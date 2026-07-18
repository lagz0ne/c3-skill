// structural-search-eval-v2 is a fail-closed adapter around the real C3 search
// boundary. The arm sees a redacted corpus and raw queries and returns only raw
// rows. The controller retains the oracle and owns probes, database inspection,
// scoring, provenance, reports, and history.
package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
	"golang.org/x/sys/unix"
)

const (
	fixtureSchema               = "structural-retrieval-fixture.v2"
	benchmarkSchema             = "structural-retrieval-benchmark.v2"
	armRequestSchema            = "structural-retrieval-arm-request.v2"
	armResponseSchema           = "structural-retrieval-arm-response.v2"
	reportSchema                = "structural-retrieval-report.v2"
	durableReportSchema         = "structural-retrieval-durable-report.v2"
	durableReportV4Schema       = "structural-retrieval-durable-report.v4"
	controllerOutputSchema      = "structural-retrieval-controller-output.v2"
	historySchema               = "structural-retrieval-history.v2"
	historyV4Schema             = "structural-retrieval-history.v4"
	sourceCapsuleSchema         = "structural-retrieval-source-capsule.v2"
	controllerAuthoritySchema   = "structural-retrieval-controller-authority.v2"
	controllerAuthorityV3Schema = "structural-retrieval-controller-authority.v3"
	controllerAuthorityV4Schema = "structural-retrieval-controller-authority.v4"

	familyWrongLayer = "wrong_layer_structural_owner"
	familyRoute      = "behavioral_route_regression"

	corpusIsolated   = "isolated"
	corpusCombined   = "combined"
	semanticDisabled = "no-semantic"
	semanticDefault  = "default-hybrid"

	decisionDiscard = "discard"
	decisionKeep    = "keep"

	registeredWallTimeMillis     int64 = 60_000
	confinementStartupMargin           = 5 * time.Second
	controllerEnvelopeOverhead         = 4 * time.Minute
	cgroupResolveDeadline              = 2 * time.Second
	cgroupResolveAttemptTimeout        = 250 * time.Millisecond
	cgroupResolveRetryInterval         = 10 * time.Millisecond
	cgroupSampleInterval               = 10 * time.Millisecond
	cgroupSampleTimeout                = 50 * time.Millisecond
	cgroupMonitorJoinTimeout           = 100 * time.Millisecond
	completionMarkerPollInterval       = 2 * time.Millisecond
	completionRetention                = 250 * time.Millisecond
	completionMarkerMaxBytes           = 256

	reconstructedSamplerMainSHA256 = "1e45eb75c0bbdf8bfe427957565b92bb6036ec1346da42d210874a5bbe2f366a"
	reconstructedSamplerTestSHA256 = "baf439870b25a09becd2ec3e9093f6745eb72f7c036ee666ae468b60ed818082"
)

const canonicalRowBytesDefinition = "UTF-8 byte length of canonical compact JSON for the exact returned SearchResultRow values, including parent-id snippets and Route fields"

var protocolPreflightSerial sync.Mutex
var protocolCommandState = struct {
	sync.RWMutex
	ctx context.Context
}{ctx: context.Background()}

func withProtocolCommandContext(ctx context.Context, operation func() error) error {
	protocolPreflightSerial.Lock()
	defer protocolPreflightSerial.Unlock()
	protocolCommandState.Lock()
	protocolCommandState.ctx = ctx
	protocolCommandState.Unlock()
	defer func() {
		protocolCommandState.Lock()
		protocolCommandState.ctx = context.Background()
		protocolCommandState.Unlock()
	}()
	return operation()
}

func protocolCommand(name string, args ...string) *exec.Cmd {
	protocolCommandState.RLock()
	ctx := protocolCommandState.ctx
	protocolCommandState.RUnlock()
	command := exec.CommandContext(ctx, name, args...)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error {
		if command.Process == nil {
			return nil
		}
		err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	command.WaitDelay = 2 * time.Second
	return command
}

type entityInput struct {
	ID, Type, Title, Slug, Category, ParentID, Goal, Status, Boundary, Date, Metadata, Markdown string
}

func (e *entityInput) UnmarshalJSON(data []byte) error {
	type wire struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Title    string `json:"title"`
		Slug     string `json:"slug"`
		Category string `json:"category"`
		ParentID string `json:"parent_id"`
		Goal     string `json:"goal"`
		Status   string `json:"status"`
		Boundary string `json:"boundary"`
		Date     string `json:"date"`
		Metadata string `json:"metadata"`
		Markdown string `json:"markdown"`
	}
	var w wire
	if err := decodeStrictBytes(data, &w); err != nil {
		return err
	}
	*e = entityInput{w.ID, w.Type, w.Title, w.Slug, w.Category, w.ParentID, w.Goal, w.Status, w.Boundary, w.Date, w.Metadata, w.Markdown}
	return nil
}

func (e entityInput) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID       string `json:"id"`
		Type     string `json:"type"`
		Title    string `json:"title"`
		Slug     string `json:"slug"`
		Category string `json:"category,omitempty"`
		ParentID string `json:"parent_id,omitempty"`
		Goal     string `json:"goal,omitempty"`
		Status   string `json:"status"`
		Boundary string `json:"boundary,omitempty"`
		Date     string `json:"date,omitempty"`
		Metadata string `json:"metadata"`
		Markdown string `json:"markdown"`
	}{e.ID, e.Type, e.Title, e.Slug, e.Category, e.ParentID, e.Goal, e.Status, e.Boundary, e.Date, e.Metadata, e.Markdown})
}

type relationshipInput struct {
	FromID  string `json:"from_id"`
	ToID    string `json:"to_id"`
	RelType string `json:"rel_type"`
}
type corpusInput struct {
	Entities      []entityInput       `json:"entities"`
	Relationships []relationshipInput `json:"relationships,omitempty"`
}
type relationshipWitness struct {
	ExpectedEntityID     string `json:"expected_entity_id"`
	FromID               string `json:"from_id"`
	ToID                 string `json:"to_id"`
	RelType              string `json:"rel_type"`
	ExpectedMatchSource  string `json:"expected_match_source"`
	RequireDirectFTSMiss bool   `json:"require_direct_fts_miss"`
}
type oracleSpec struct {
	RequiredOwnerFactIDs []string             `json:"required_owner_fact_ids"`
	AllowedExtraFactIDs  []string             `json:"allowed_extra_fact_ids"`
	ForbiddenFactIDs     []string             `json:"forbidden_fact_ids"`
	FactBindings         map[string][]string  `json:"fact_bindings"`
	RelationshipWitness  *relationshipWitness `json:"relationship_witness,omitempty"`
	RequiredRouteFields  []string             `json:"required_route_fields,omitempty"`
}
type fixtureCase struct {
	Schema string      `json:"$schema"`
	CaseID string      `json:"case_id"`
	Family string      `json:"family"`
	Query  string      `json:"query"`
	Corpus corpusInput `json:"corpus"`
	Oracle oracleSpec  `json:"oracle"`
}
type scaleConfig struct {
	Seed                  int      `json:"seed"`
	Multiplier            int      `json:"multiplier"`
	Tokens                []string `json:"tokens"`
	MaxRelationshipDegree int      `json:"max_relationship_degree"`
}
type thresholds struct {
	OwnerRecallAt5Delta      float64 `json:"owner_recall_at_5_delta"`
	StructuralOwnerPrecision float64 `json:"structural_owner_precision"`
	CanonicalRowBytesRatio   float64 `json:"canonical_row_bytes_ratio"`
}
type thresholdAuthority struct {
	CheckinRef       string `json:"checkin_ref"`
	CheckinSHA256    string `json:"checkin_sha256"`
	RecordHash       string `json:"record_hash"`
	DefinitionSHA256 string `json:"definition_sha256"`
}
type benchmarkConfig struct {
	Schema                    string              `json:"$schema"`
	FixtureSHA256             string              `json:"fixture_sha256"`
	FixtureCount              int                 `json:"fixture_count"`
	K                         int                 `json:"k"`
	SemanticMode              string              `json:"semantic_mode"`
	Scale                     scaleConfig         `json:"scale"`
	Thresholds                thresholds          `json:"thresholds"`
	ContextThresholdAuthority *thresholdAuthority `json:"context_threshold_authority,omitempty"`
}

type armQuery struct {
	CaseID string `json:"case_id"`
	Query  string `json:"query"`
}
type armRequest struct {
	Schema       string      `json:"$schema"`
	Corpus       corpusInput `json:"corpus"`
	Queries      []armQuery  `json:"queries"`
	SemanticMode string      `json:"semantic_mode"`
}
type armCaseResult struct {
	CaseID string                `json:"case_id"`
	Query  string                `json:"query,omitempty"`
	Rows   []cmd.SearchResultRow `json:"rows"`
}
type armResponse struct {
	Schema string          `json:"$schema"`
	Cases  []armCaseResult `json:"cases"`
}
type directProbes struct {
	EntityFTSIDs  []string `json:"entity_fts_ids"`
	ContentFTSIDs []string `json:"content_fts_ids"`
}
type logicalEntity struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Title      string `json:"title"`
	Slug       string `json:"slug"`
	ParentID   string `json:"parent_id,omitempty"`
	Goal       string `json:"goal"`
	Status     string `json:"status"`
	Metadata   string `json:"metadata"`
	RootMerkle string `json:"root_merkle"`
	Version    int    `json:"version"`
}
type logicalRelationship struct {
	FromID  string `json:"from_id"`
	ToID    string `json:"to_id"`
	RelType string `json:"rel_type"`
}
type logicalNode struct {
	EntityID string `json:"entity_id"`
	ParentID int64  `json:"parent_id"`
	Type     string `json:"type"`
	Level    int    `json:"level"`
	Seq      int    `json:"seq"`
	Content  string `json:"content"`
	Hash     string `json:"hash"`
}
type logicalDump struct {
	Integrity         string                `json:"integrity"`
	SchemaSQL         []string              `json:"schema_sql"`
	Entities          []logicalEntity       `json:"entities"`
	Relationships     []logicalRelationship `json:"relationships"`
	Nodes             []logicalNode         `json:"nodes"`
	EntityCount       int                   `json:"entity_count"`
	RelationshipCount int                   `json:"relationship_count"`
	SQLiteRowCount    int                   `json:"sqlite_row_count"`
	LogicalBytes      int                   `json:"logical_bytes"`
	LogicalSHA256     string                `json:"logical_sha256"`
}
type controllerResult struct {
	Cases        []armCaseResult         `json:"cases"`
	Database     logicalDump             `json:"database"`
	DirectProbes map[string]directProbes `json:"direct_probes"`
}
type controllerRun struct {
	Mode                  string                      `json:"mode"`
	CaseID                string                      `json:"case_id,omitempty"`
	Result                controllerResult            `json:"-"`
	RawResult             []byte                      `json:"-"`
	RawStderr             []byte                      `json:"-"`
	Report                evalReport                  `json:"-"`
	ActualBudget          resourceBudget              `json:"actual_budget"`
	AccountingDiagnostics cgroupAccountingDiagnostics `json:"-"`
	ProjectDirSHA256      string                      `json:"-"`
	C3DirSHA256           string                      `json:"-"`
	Fixtures              []fixtureCase               `json:"-"`
}
type controllerRunRef struct {
	Mode                 string         `json:"mode"`
	CaseID               string         `json:"case_id,omitempty"`
	ResultPath           string         `json:"result_path"`
	ResultSHA256         string         `json:"result_sha256"`
	ReportPath           string         `json:"report_path"`
	ReportSHA256         string         `json:"report_sha256"`
	HistoryRecordHash    string         `json:"history_record_hash"`
	ActualBudget         resourceBudget `json:"actual_budget"`
	CandidateDeltaSHA256 string         `json:"candidate_delta_sha256,omitempty"`
	BundleSHA256         string         `json:"bundle_sha256,omitempty"`
}
type controllerFailureRef struct {
	Mode                 string         `json:"mode"`
	CaseID               string         `json:"case_id,omitempty"`
	ErrorClass           string         `json:"error_class"`
	ErrorSHA256          string         `json:"error_sha256"`
	EvidencePath         string         `json:"evidence_path"`
	EvidenceSHA256       string         `json:"evidence_sha256"`
	HistoryRecordHash    string         `json:"history_record_hash"`
	ActualBudget         resourceBudget `json:"actual_budget"`
	CandidateDeltaSHA256 string         `json:"candidate_delta_sha256,omitempty"`
	BundleSHA256         string         `json:"bundle_sha256,omitempty"`
}
type controllerOutput struct {
	Schema                   string                `json:"$schema"`
	Runs                     []controllerRunRef    `json:"runs"`
	HistoryPath              string                `json:"history_path"`
	HistorySHA256            string                `json:"history_sha256"`
	PrivacyManifestPath      string                `json:"privacy_manifest_path,omitempty"`
	PrivacyManifestSHA256    string                `json:"privacy_manifest_sha256,omitempty"`
	OrderedRunManifestSHA256 string                `json:"ordered_run_manifest_sha256,omitempty"`
	Admitted                 bool                  `json:"admitted"`
	Admission                string                `json:"admission"`
	Failure                  *controllerFailureRef `json:"failure,omitempty"`
}

func main() {
	if err := runCLI(os.Args[1:], os.Stdin, os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runCLI(args []string, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("usage: structural-search-eval-v2 <--arm|--controller>")
	}
	switch args[0] {
	case "--arm":
		dbPath, err := requiredArg(args[1:], "--db")
		if err != nil {
			return err
		}
		projectDir, err := requiredArg(args[1:], "--project")
		if err != nil {
			return err
		}
		c3Dir, err := requiredArg(args[1:], "--c3")
		if err != nil {
			return err
		}
		req, err := decodeArmRequest(stdin, 16<<20)
		if err != nil {
			return err
		}
		response, err := executeArm(req, dbPath, projectDir, c3Dir)
		if err != nil {
			return err
		}
		return json.NewEncoder(stdout).Encode(response)
	case "--controller":
		return runControllerCLI(args[1:], stdout)
	default:
		return fmt.Errorf("unknown mode %q", args[0])
	}
}

func requiredArg(args []string, name string) (string, error) {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name && strings.TrimSpace(args[i+1]) != "" {
			return args[i+1], nil
		}
	}
	return "", fmt.Errorf("%s is required", name)
}

func hasArg(args []string, name string) bool {
	for _, arg := range args {
		if arg == name {
			return true
		}
	}
	return false
}

func runControllerCLI(args []string, stdout io.Writer) error {
	runtimePath, err := requiredArg(args, "--runtime")
	if err != nil {
		return err
	}
	fixturesPath, err := requiredArg(args, "--fixtures")
	if err != nil {
		return err
	}
	benchmarkPath, err := requiredArg(args, "--benchmark")
	if err != nil {
		return err
	}
	workRoot, err := requiredArg(args, "--work-root")
	if err != nil {
		return err
	}
	authorityPath, err := requiredArg(args, "--authority")
	if err != nil {
		return err
	}
	authorityBytes, err := readBoundedStandaloneRegularFile(authorityPath, 4<<20)
	if err != nil {
		return err
	}
	authoritySchema, err := controllerAuthoritySchemaBytes(authorityBytes)
	if err != nil {
		return err
	}
	if authoritySchema != controllerAuthorityV4Schema {
		return errors.New("protocol-v7 required: legacy controller authorities are read-only evidence")
	}
	controllerSourceRoot, err := requiredArg(args, "--controller-source-root")
	if err != nil {
		return err
	}
	privacyPolicyPath, err := requiredArg(args, "--privacy-policy")
	if err != nil {
		return err
	}
	runtimeSourceRoot, err := requiredArg(args, "--runtime-source-root")
	if err != nil {
		return err
	}
	bundlePath, err := requiredArg(args, "--bundle")
	if err != nil {
		return err
	}
	scorerPath, err := requiredArg(args, "--scorer-source")
	if err != nil {
		return err
	}
	outputDir, err := requiredArg(args, "--output-dir")
	if err != nil {
		return err
	}
	var loadedV4 controllerAuthorityV4
	if authoritySchema == controllerAuthorityV4Schema {
		loadedV4, err = decodeControllerAuthorityV4(bytes.NewReader(authorityBytes))
	}
	if err != nil {
		return err
	}
	fixtures, fixtureHash, err := loadFixtures(fixturesPath)
	if err != nil {
		return err
	}
	bench, err := loadBenchmark(benchmarkPath)
	if err != nil {
		return err
	}
	if bench.FixtureSHA256 != fixtureHash || bench.FixtureCount != len(fixtures) {
		return errors.New("fixture freeze mismatch")
	}
	controllerPath, err := os.Executable()
	if err != nil {
		return err
	}
	controllerPath, err = filepath.EvalSymlinks(controllerPath)
	if err != nil {
		return err
	}
	var parentFiles parentBaselineFiles
	if loadedV4.Mode == "candidate" {
		parentFiles.Root, err = requiredArg(args, "--parent-baseline-root")
		if err != nil {
			return err
		}
		parentFiles.Authority, err = requiredArg(args, "--parent-baseline-authority")
		if err != nil {
			return err
		}
		parentFiles.Output, err = requiredArg(args, "--parent-baseline-output")
		if err != nil {
			return err
		}
		parentFiles.ValidatorStore, err = requiredArg(args, "--parent-baseline-validator-store")
		if err != nil {
			return err
		}
	} else {
		for _, forbidden := range []string{"--parent-baseline-root", "--parent-baseline-authority", "--parent-baseline-output", "--parent-baseline-validator-store"} {
			if hasArg(args, forbidden) {
				return errors.New("baseline authority cannot name parent-baseline files")
			}
		}
	}
	var authority controllerAuthority
	frozenInputs := []string{authorityPath, fixturesPath, benchmarkPath, scorerPath}
	var protocolV7Scanner *privacyScanner
	if authoritySchema == controllerAuthorityV4Schema {
		policyBytes, openErr := readBoundedStandaloneRegularFile(privacyPolicyPath, privacyPolicyBytesMax)
		if openErr != nil {
			return openErr
		}
		policy, policyErr := decodePrivacyPolicy(bytes.NewReader(policyBytes))
		if policyErr != nil {
			return policyErr
		}
		protocolV7Scanner, err = newPrivacyScanner(policy)
		if err != nil {
			return err
		}
		if err := verifyControllerAuthorityV4(loadedV4, authorityBytes, runtimePath, controllerPath, controllerSourceRoot, runtimeSourceRoot, bundlePath, privacyPolicyPath, fixturesPath, benchmarkPath, scorerPath, bench, policy, protocolV7Scanner, parentFiles); err != nil {
			return fmt.Errorf("controller preflight: %w", err)
		}
		authority = loadedV4.persistenceAuthority()
		frozenInputs = append(frozenInputs, controllerSourceRoot, runtimeSourceRoot, bundlePath, privacyPolicyPath)
		if loadedV4.Mode == "candidate" {
			frozenInputs = append(frozenInputs, parentFiles.Root, parentFiles.Authority, parentFiles.Output, parentFiles.ValidatorStore)
		}
	}
	for _, left := range []string{workRoot, outputDir} {
		for _, right := range frozenInputs {
			if pathsOverlap(left, right) {
				return errors.New("controller work/output overlaps a frozen input")
			}
		}
	}
	if pathsOverlap(workRoot, outputDir) {
		return errors.New("controller work and output directories overlap")
	}
	if authoritySchema == controllerAuthorityV4Schema {
		return runControllerV4Execution(runtimePath, fixtures, bench, workRoot, outputDir, loadedV4, protocolV7Scanner, stdout)
	}
	controllerTimeout, err := controllerTimeoutFor(len(fixtures)+2, authority.BudgetLimits)
	if err != nil {
		return fmt.Errorf("controller wall envelope: %w", err)
	}
	persistence, err := newControllerPersistence(outputDir)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), controllerTimeout)
	defer cancel()
	runIndex := 0
	runPlan := func(fixtures []fixtureCase, requestMode, reportMode, caseID, root string) error {
		planned := controllerRun{Mode: reportMode, CaseID: caseID, Fixtures: append([]fixtureCase(nil), fixtures...)}
		run, runErr := executeControllerRun(ctx, runtimePath, fixtures, requestMode, reportMode, bench, root, authority.BudgetLimits)
		run.CaseID = caseID
		if run.Mode == "" {
			run.Mode = planned.Mode
		}
		if len(run.Fixtures) == 0 {
			run.Fixtures = planned.Fixtures
		}
		if runErr != nil {
			return persistence.fail(stdout, run, runIndex, "runtime_or_inspection", "crash", runErr, authority, bench)
		}
		if persistErr := persistence.persistSuccessfulRun(run, runIndex, authority, bench); persistErr != nil {
			return persistence.fail(stdout, run, runIndex, "persistence", "invalid", persistErr, authority, bench)
		}
		runIndex++
		return nil
	}
	for _, fixture := range fixtures {
		if err := runPlan([]fixtureCase{fixture}, corpusIsolated, corpusIsolated, fixture.CaseID, filepath.Join(workRoot, "isolated-"+safePathComponent(fixture.CaseID))); err != nil {
			return err
		}
	}
	if err := runPlan(fixtures, corpusCombined, corpusCombined, "", filepath.Join(workRoot, "combined")); err != nil {
		return err
	}
	scaled := append([]fixtureCase(nil), fixtures...)
	for i := range scaled {
		cfg := bench.Scale
		cfg.Seed += i * 1000
		scaled[i].Corpus = generateScaleCorpus(scaled[i].Corpus, cfg)
	}
	if err := runPlan(scaled, corpusCombined, "scale", "", filepath.Join(workRoot, "scale")); err != nil {
		return err
	}
	return persistence.emit(stdout)
}

func runControllerV4Execution(runtimePath string, fixtures []fixtureCase, bench benchmarkConfig, workRoot, outputDir string, authority controllerAuthorityV4, scanner *privacyScanner, stdout io.Writer) error {
	for _, path := range []string{workRoot, outputDir} {
		if _, err := os.Lstat(path); err == nil {
			return fmt.Errorf("protocol-v7 caller path already exists: %s", filepath.Base(path))
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	timeout, err := controllerTimeoutFor(len(fixtures)+2, authority.BudgetLimits)
	if err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "c3-v7-controller-work-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	persistence, err := newControllerMemoryPersistence(outputDir, scanner, authority, bench, scanner.sourceObjectCount, scanner.sourceObjectBytes)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	runIndex := 0
	runPlan := func(selected []fixtureCase, requestMode, reportMode, caseID, root string) error {
		planned := controllerRun{Mode: reportMode, CaseID: caseID, Fixtures: append([]fixtureCase(nil), selected...)}
		run, runErr := executeControllerRun(ctx, runtimePath, selected, requestMode, reportMode, bench, root, authority.BudgetLimits)
		run.CaseID = caseID
		if run.Mode == "" {
			run.Mode = planned.Mode
		}
		if len(run.Fixtures) == 0 {
			run.Fixtures = planned.Fixtures
		}
		if runErr != nil {
			if err := persistence.persistFailure(run, runIndex, "runtime_or_inspection", "crash", runErr); err != nil {
				if persistence.tx.tainted {
					return errors.New("privacy policy violation")
				}
				return err
			}
			if err := persistence.finish(stdout); err != nil {
				return err
			}
			return fmt.Errorf("runtime_or_inspection: %w", runErr)
		}
		if err := persistence.persistSuccessfulRun(run, runIndex); err != nil {
			if persistence.tx.tainted {
				return errors.New("privacy policy violation")
			}
			return err
		}
		runIndex++
		return nil
	}
	for _, fixture := range fixtures {
		if err := runPlan([]fixtureCase{fixture}, corpusIsolated, corpusIsolated, fixture.CaseID, filepath.Join(scratch, "isolated-"+safePathComponent(fixture.CaseID))); err != nil {
			return err
		}
	}
	if err := runPlan(fixtures, corpusCombined, corpusCombined, "", filepath.Join(scratch, "combined")); err != nil {
		return err
	}
	scaled := append([]fixtureCase(nil), fixtures...)
	for i := range scaled {
		cfg := bench.Scale
		cfg.Seed += i * 1000
		scaled[i].Corpus = generateScaleCorpus(scaled[i].Corpus, cfg)
	}
	if err := runPlan(scaled, corpusCombined, "scale", "", filepath.Join(scratch, "scale")); err != nil {
		return err
	}
	return persistence.finish(stdout)
}

func pathsOverlap(left, right string) bool {
	a, err := resolvedAbsolutePath(left)
	if err != nil {
		return true
	}
	b, err := resolvedAbsolutePath(right)
	if err != nil {
		return true
	}
	return a == b || strings.HasPrefix(a, b+string(os.PathSeparator)) || strings.HasPrefix(b, a+string(os.PathSeparator))
}

func resolvedAbsolutePath(path string) (string, error) {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	absolute = filepath.Clean(absolute)
	if resolved, err := filepath.EvalSymlinks(absolute); err == nil {
		return filepath.Clean(resolved), nil
	}
	current := absolute
	var suffix []string
	for {
		parent := filepath.Dir(current)
		if parent == current {
			return "", fmt.Errorf("cannot resolve existing ancestor for %s", path)
		}
		suffix = append(suffix, filepath.Base(current))
		current = parent
		resolved, err := filepath.EvalSymlinks(current)
		if err != nil {
			continue
		}
		for i := len(suffix) - 1; i >= 0; i-- {
			resolved = filepath.Join(resolved, suffix[i])
		}
		return filepath.Clean(resolved), nil
	}
}

func executeControllerRun(ctx context.Context, runtimePath string, fixtures []fixtureCase, requestMode, reportMode string, bench benchmarkConfig, root string, limits resourceBudget) (controllerRun, error) {
	run := controllerRun{Mode: reportMode, Fixtures: append([]fixtureCase(nil), fixtures...)}
	req, err := buildArmRequest(fixtures, requestMode, bench.SemanticMode, bench)
	if err != nil {
		return run, err
	}
	result, invocation, projectHash, c3Hash, err := runConfinedArmMeasured(ctx, runtimePath, req, root, limits)
	run.Result = result
	run.RawResult = invocation.RawStdout
	run.RawStderr = invocation.RawStderr
	run.ActualBudget = invocation.ActualBudget
	run.AccountingDiagnostics = invocation.AccountingDiagnostics
	run.ProjectDirSHA256 = projectHash
	run.C3DirSHA256 = c3Hash
	if err != nil {
		return run, err
	}
	report, err := scoreArmResponseWithProbes(fixtures, armResponse{Schema: armResponseSchema, Cases: result.Cases}, result.Database, reportMode, result.DirectProbes)
	if err != nil {
		return run, err
	}
	run.Report = report
	return run, nil
}

type durableReport struct {
	Schema                              string                      `json:"$schema"`
	Report                              evalReport                  `json:"report"`
	Database                            logicalDump                 `json:"database"`
	DirectProbes                        map[string]directProbes     `json:"direct_probes"`
	ActualBudget                        resourceBudget              `json:"actual_budget"`
	AccountingDiagnostics               cgroupAccountingDiagnostics `json:"accounting_diagnostics"`
	BudgetLimits                        resourceBudget              `json:"budget_limits"`
	BudgetVerdict                       string                      `json:"budget_verdict"`
	ResultPath                          string                      `json:"result_path"`
	ResultSHA256                        string                      `json:"result_sha256"`
	CanonicalRowBytesDefinition         string                      `json:"canonical_row_bytes_definition"`
	ContextThresholdAuthority           *thresholdAuthority         `json:"context_threshold_authority"`
	ContextThresholdAuthorityLineSHA256 string                      `json:"context_threshold_authority_line_sha256"`
}

type controllerPersistence struct {
	outputDir        string
	historyPath      string
	output           controllerOutput
	snapshotSeq      int
	lastSnapshotPath string
	lastSnapshotHash string
}

type privacyManifest struct {
	Schema              string            `json:"$schema"`
	PrivacyPolicySHA256 string            `json:"privacy_policy_sha256"`
	PrivacyTermCount    int               `json:"privacy_term_count"`
	DetectorVersion     string            `json:"detector_version"`
	DetectorSHA256      string            `json:"detector_definition_sha256"`
	ScanCaps            privacyScanCaps   `json:"scan_caps"`
	SourceObjectCount   int               `json:"source_object_count"`
	SourceObjectBytes   int64             `json:"source_object_uncompressed_bytes"`
	ArtifactCount       int               `json:"artifact_count"`
	TotalScannedBytes   int64             `json:"total_scanned_bytes"`
	Artifacts           []privacyArtifact `json:"artifacts"`
	Hits                int               `json:"hits"`
}

type controllerMemoryPersistence struct {
	tx                   *privacyTransaction
	authority            controllerAuthority
	authorityV4          controllerAuthorityV4
	bench                benchmarkConfig
	history              []historyRecord
	output               controllerOutput
	sourceObjectCount    int
	sourceObjectBytes    int64
	durableArtifactBytes int64
}

type controllerFailureEvidence struct {
	Schema                string                      `json:"$schema"`
	Mode                  string                      `json:"mode"`
	CaseID                string                      `json:"case_id,omitempty"`
	ErrorClass            string                      `json:"error_class"`
	Error                 string                      `json:"error"`
	ErrorSHA256           string                      `json:"error_sha256"`
	RawResultPath         string                      `json:"raw_result_path"`
	RawResultSHA256       string                      `json:"raw_result_sha256"`
	ActualBudget          resourceBudget              `json:"actual_budget"`
	AccountingDiagnostics cgroupAccountingDiagnostics `json:"accounting_diagnostics"`
	PriorHistoryTail      string                      `json:"prior_history_tail"`
	Provenance            provenance                  `json:"provenance"`
}

func populateDurableAccounting(report *durableReport, run controllerRun) {
	report.ActualBudget = run.ActualBudget
	report.AccountingDiagnostics = run.AccountingDiagnostics
}

func newControllerMemoryPersistence(outputRoot string, scanner *privacyScanner, authorityV4 controllerAuthorityV4, bench benchmarkConfig, sourceObjectCount int, sourceObjectBytes int64) (*controllerMemoryPersistence, error) {
	if scanner == nil || authorityV4.Schema != controllerAuthorityV4Schema {
		return nil, errors.New("invalid protocol-v7 memory persistence authority")
	}
	return &controllerMemoryPersistence{
		tx: newPrivacyTransaction(outputRoot, scanner), authority: authorityV4.persistenceAuthority(), authorityV4: authorityV4,
		bench: bench, sourceObjectCount: sourceObjectCount, sourceObjectBytes: sourceObjectBytes,
		output: controllerOutput{Schema: "structural-retrieval-controller-output.v4", HistoryPath: "history.jsonl", Admitted: false, Admission: "diagnostic_unadmitted"},
	}, nil
}

func (s *controllerMemoryPersistence) appendHistory(record historyRecord) error {
	if len(s.history) == 0 {
		record.PrevHash = "GENESIS"
	} else {
		record.PrevHash = s.history[len(s.history)-1].RecordHash
	}
	record.RecordHash = hashHistoryRecord(record)
	if err := verifyHistorySchema(append(append([]historyRecord(nil), s.history...), record), historyV4Schema); err != nil {
		return err
	}
	s.history = append(s.history, record)
	return nil
}

func (s *controllerMemoryPersistence) historyBytes() ([]byte, error) {
	var output bytes.Buffer
	for _, record := range s.history {
		data, err := json.Marshal(record)
		if err != nil {
			return nil, err
		}
		output.Write(data)
		output.WriteByte('\n')
	}
	return output.Bytes(), nil
}

func (s *controllerMemoryPersistence) historyChange() (string, string, []string) {
	if s.authorityV4.Mode == "candidate" && s.authorityV4.ParentBaseline != nil {
		return s.authorityV4.ParentBaseline.HistoryTailRecordHash, "direct_hit_containment_owner_substitution", append([]string(nil), s.authorityV4.CandidateDelta.AllowedPaths...)
	}
	return "GENESIS", "baseline", []string{}
}

func (s *controllerMemoryPersistence) persistSuccessfulRun(run controllerRun, index int) error {
	if !actualBudgetComplete(run.ActualBudget) {
		return fmt.Errorf("run %d has incomplete actual budget", index)
	}
	if err := s.tx.Add("runtime_stderr", fmt.Sprintf("runtime/%02d.stderr", index+1), run.RawStderr); err != nil {
		s.tx.tainted = true
		return err
	}
	response, err := decodeArmResponse(bytes.NewReader(run.RawResult), 16<<20)
	if err != nil || !reflect.DeepEqual(response.Cases, run.Result.Cases) {
		return fmt.Errorf("re-read exact arm result %d", index)
	}
	if err := verifyReportWithProbes(run.Fixtures, response, run.Result.Database, run.Mode, run.Result.DirectProbes, run.Report); err != nil {
		return fmt.Errorf("recompute report %d: %w", index, err)
	}
	stem := controllerRunStem(index, run)
	resultRel := filepath.ToSlash(filepath.Join("results", stem+".json"))
	if err := s.tx.Add("result", resultRel, run.RawResult); err != nil {
		return err
	}
	resultHash := shaString(string(run.RawResult))
	p := completedRunProvenance(s.authority, run, index)
	if err := validateBaselineAdmission(s.bench, p); err != nil {
		return err
	}
	report := run.Report
	report.Provenance = &p
	budgetVerdict := "within_limits"
	if classifyBudget(s.authority.BudgetLimits, run.ActualBudget) != "score" {
		budgetVerdict = "over_limit"
	}
	envelope := durableReport{
		Schema: durableReportV4Schema, Report: report, Database: run.Result.Database, DirectProbes: run.Result.DirectProbes,
		BudgetLimits: s.authority.BudgetLimits, BudgetVerdict: budgetVerdict, ResultPath: resultRel, ResultSHA256: resultHash,
		CanonicalRowBytesDefinition: s.authority.CanonicalRowBytesDefinition, ContextThresholdAuthority: s.bench.ContextThresholdAuthority,
		ContextThresholdAuthorityLineSHA256: shaString(s.authority.ContextThresholdAuthorityRecord),
	}
	populateDurableAccounting(&envelope, run)
	reportBytes, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	reportBytes = append(reportBytes, '\n')
	reportRel := filepath.ToSlash(filepath.Join("reports", stem+".json"))
	if err := s.tx.Add("report", reportRel, reportBytes); err != nil {
		return err
	}
	reportHash := shaString(string(reportBytes))
	parent, variable, paths := s.historyChange()
	record := historyRecord{
		Schema: historyV4Schema, ExperimentID: p.ExperimentID, ArmID: p.ArmID, ParentKeep: parent, ChangedVariable: variable,
		ChangedPaths: paths, ResultPath: resultRel, ResultSHA256: resultHash, Provenance: p, Budgets: run.ActualBudget,
		Status: "invalid", Reason: "diagnostic_unadmitted", Evidence: []string{reportRel + "#sha256=" + reportHash, s.bench.ContextThresholdAuthority.CheckinRef + "#sha256=" + s.bench.ContextThresholdAuthority.CheckinSHA256},
	}
	if err := s.appendHistory(record); err != nil {
		return err
	}
	last := s.history[len(s.history)-1]
	s.output.Runs = append(s.output.Runs, controllerRunRef{Mode: run.Mode, CaseID: run.CaseID, ResultPath: resultRel, ResultSHA256: resultHash, ReportPath: reportRel, ReportSHA256: reportHash, HistoryRecordHash: last.RecordHash, ActualBudget: run.ActualBudget, CandidateDeltaSHA256: p.CandidateDeltaSHA256, BundleSHA256: p.BundleSHA256})
	return nil
}

func (s *controllerMemoryPersistence) persistFailure(run controllerRun, index int, errorClass, status string, cause error) error {
	if err := s.tx.Add("runtime_stderr", fmt.Sprintf("runtime/%02d.stderr", index+1), run.RawStderr); err != nil {
		s.tx.tainted = true
		return err
	}
	if err := s.tx.scanner.Scan("failure_error", fmt.Sprintf("failure/%02d.error", index+1), []byte(cause.Error())); err != nil {
		s.tx.tainted = true
		return err
	}
	stem := controllerRunStem(index, run)
	rawRel := filepath.ToSlash(filepath.Join("failures", stem+"-raw.bin"))
	if err := s.tx.Add("failure_raw", rawRel, run.RawResult); err != nil {
		return err
	}
	rawHash := shaString(string(run.RawResult))
	priorTail := "GENESIS"
	if len(s.history) > 0 {
		priorTail = s.history[len(s.history)-1].RecordHash
	}
	errorHash := shaString(cause.Error())
	p := completedRunProvenance(s.authority, run, index)
	evidence := controllerFailureEvidence{Schema: "structural-retrieval-controller-failure.v4", Mode: run.Mode, CaseID: run.CaseID, ErrorClass: errorClass, Error: cause.Error(), ErrorSHA256: errorHash, RawResultPath: rawRel, RawResultSHA256: rawHash, ActualBudget: run.ActualBudget, AccountingDiagnostics: run.AccountingDiagnostics, PriorHistoryTail: priorTail, Provenance: p}
	evidenceBytes, err := json.Marshal(evidence)
	if err != nil {
		return err
	}
	evidenceBytes = append(evidenceBytes, '\n')
	evidenceRel := filepath.ToSlash(filepath.Join("failures", stem+".json"))
	if err := s.tx.Add("failure_evidence", evidenceRel, evidenceBytes); err != nil {
		return err
	}
	evidenceHash := shaString(string(evidenceBytes))
	parent, variable, paths := s.historyChange()
	record := historyRecord{
		Schema: historyV4Schema, ExperimentID: p.ExperimentID, ArmID: p.ArmID, ParentKeep: parent, ChangedVariable: variable,
		ChangedPaths: paths, ResultPath: rawRel, ResultSHA256: rawHash, Provenance: p, Budgets: run.ActualBudget,
		Status: status, Reason: "controller_" + errorClass, ErrorClass: errorClass, ErrorSHA256: errorHash,
		Evidence: []string{evidenceRel + "#sha256=" + evidenceHash, s.bench.ContextThresholdAuthority.CheckinRef + "#sha256=" + s.bench.ContextThresholdAuthority.CheckinSHA256},
	}
	if err := s.appendHistory(record); err != nil {
		return err
	}
	last := s.history[len(s.history)-1]
	s.output.Failure = &controllerFailureRef{Mode: run.Mode, CaseID: run.CaseID, ErrorClass: errorClass, ErrorSHA256: errorHash, EvidencePath: evidenceRel, EvidenceSHA256: evidenceHash, HistoryRecordHash: last.RecordHash, ActualBudget: run.ActualBudget, CandidateDeltaSHA256: p.CandidateDeltaSHA256, BundleSHA256: p.BundleSHA256}
	return nil
}

func (s *controllerMemoryPersistence) finish(stdout io.Writer) error {
	historyBytes, err := s.historyBytes()
	if err != nil {
		return err
	}
	if err := s.tx.Add("history", s.output.HistoryPath, historyBytes); err != nil {
		return err
	}
	s.output.HistorySHA256 = shaString(string(historyBytes))
	s.output.OrderedRunManifestSHA256 = canonicalSHA256(s.output.Runs)
	entries := append([]privacyArtifact(nil), s.tx.scanner.entries...)
	sort.Slice(entries, func(i, j int) bool { return privacyArtifactLess(entries[i], entries[j]) })
	var total int64
	for _, entry := range entries {
		total += int64(entry.Bytes)
	}
	manifest := privacyManifest{
		Schema: "structural-retrieval-privacy-scan.v1", PrivacyPolicySHA256: s.tx.scanner.policy.SHA256,
		PrivacyTermCount: len(s.tx.scanner.policy.DenyTerms), DetectorVersion: s.tx.scanner.detector.Version,
		DetectorSHA256: s.tx.scanner.detector.DefinitionSHA256, ScanCaps: s.authorityV4.ScanCaps,
		SourceObjectCount: s.sourceObjectCount, SourceObjectBytes: s.sourceObjectBytes,
		ArtifactCount: len(entries), TotalScannedBytes: total,
		Artifacts: entries, Hits: 0,
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	manifestBytes = append(manifestBytes, '\n')
	manifestPath := "privacy-scan.json"
	if err := s.tx.Add("privacy_manifest", manifestPath, manifestBytes); err != nil {
		return err
	}
	s.output.PrivacyManifestPath = manifestPath
	s.output.PrivacyManifestSHA256 = shaString(string(manifestBytes))
	outputBytes, err := json.Marshal(s.output)
	if err != nil {
		return err
	}
	outputBytes = append(outputBytes, '\n')
	return s.tx.Publish(stdout, outputBytes)
}

func newControllerPersistence(outputDir string) (*controllerPersistence, error) {
	if err := os.Mkdir(outputDir, 0o700); err != nil {
		return nil, fmt.Errorf("create exclusive output directory: %w", err)
	}
	s := &controllerPersistence{
		outputDir:   outputDir,
		historyPath: filepath.Join(outputDir, "history.jsonl"),
		output:      controllerOutput{Schema: controllerOutputSchema, HistoryPath: "history.jsonl", Admitted: false, Admission: "diagnostic_unadmitted"},
	}
	if err := writeExclusive(s.historyPath, nil); err != nil {
		return nil, fmt.Errorf("initialize append-only history: %w", err)
	}
	if _, _, err := s.writeSnapshot(); err != nil {
		return nil, fmt.Errorf("initialize durable controller output: %w", err)
	}
	return s, nil
}

func (s *controllerPersistence) writeSnapshot() (string, string, error) {
	historyHash, err := fileSHA256(s.historyPath)
	if err != nil {
		return "", "", err
	}
	s.output.HistorySHA256 = historyHash
	data, err := json.Marshal(s.output)
	if err != nil {
		return "", "", err
	}
	data = append(data, '\n')
	rel := filepath.ToSlash(filepath.Join("controller-output", fmt.Sprintf("%03d.json", s.snapshotSeq)))
	path := filepath.Join(s.outputDir, filepath.FromSlash(rel))
	if err := writeExclusive(path, data); err != nil {
		return "", "", err
	}
	hash, err := fileSHA256(path)
	if err != nil {
		return "", "", err
	}
	s.snapshotSeq++
	s.lastSnapshotPath = rel
	s.lastSnapshotHash = hash
	return rel, hash, nil
}

func (s *controllerPersistence) emit(stdout io.Writer) error {
	return json.NewEncoder(stdout).Encode(s.output)
}

func (s *controllerPersistence) persistSuccessfulRun(run controllerRun, index int, authority controllerAuthority, bench benchmarkConfig) error {
	if !actualBudgetComplete(run.ActualBudget) {
		return fmt.Errorf("run %d has incomplete actual budget", index)
	}
	response, err := decodeArmResponse(bytes.NewReader(run.RawResult), 16<<20)
	if err != nil {
		return fmt.Errorf("re-read exact arm result %d: %w", index, err)
	}
	if !reflect.DeepEqual(response.Cases, run.Result.Cases) {
		return fmt.Errorf("exact arm result %d differs from inspected rows", index)
	}
	if err := verifyReportWithProbes(run.Fixtures, response, run.Result.Database, run.Mode, run.Result.DirectProbes, run.Report); err != nil {
		return fmt.Errorf("recompute report %d: %w", index, err)
	}
	stem := controllerRunStem(index, run)
	resultRel := filepath.ToSlash(filepath.Join("results", stem+".json"))
	resultPath := filepath.Join(s.outputDir, filepath.FromSlash(resultRel))
	if err := writeExclusive(resultPath, run.RawResult); err != nil {
		return fmt.Errorf("write exact result %d: %w", index, err)
	}
	resultHash, err := fileSHA256(resultPath)
	if err != nil {
		return err
	}
	p := completedRunProvenance(authority, run, index)
	if err := validateBaselineAdmission(bench, p); err != nil {
		return fmt.Errorf("run %d provenance/authority: %w", index, err)
	}
	report := run.Report
	report.Provenance = &p
	budgetVerdict := "within_limits"
	if classifyBudget(authority.BudgetLimits, run.ActualBudget) != "score" {
		budgetVerdict = "over_limit"
	}
	envelope := durableReport{
		Schema: durableReportSchema, Report: report, Database: run.Result.Database,
		DirectProbes: run.Result.DirectProbes,
		BudgetLimits: authority.BudgetLimits, BudgetVerdict: budgetVerdict,
		ResultPath: resultRel, ResultSHA256: resultHash,
		CanonicalRowBytesDefinition:         authority.CanonicalRowBytesDefinition,
		ContextThresholdAuthority:           bench.ContextThresholdAuthority,
		ContextThresholdAuthorityLineSHA256: shaString(authority.ContextThresholdAuthorityRecord),
	}
	populateDurableAccounting(&envelope, run)
	reportBytes, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	reportBytes = append(reportBytes, '\n')
	reportRel := filepath.ToSlash(filepath.Join("reports", stem+".json"))
	reportPath := filepath.Join(s.outputDir, filepath.FromSlash(reportRel))
	if err := writeExclusive(reportPath, reportBytes); err != nil {
		return fmt.Errorf("write report %d: %w", index, err)
	}
	reportHash, err := fileSHA256(reportPath)
	if err != nil {
		return err
	}
	record := historyRecord{
		Schema: historySchema, ExperimentID: p.ExperimentID, ArmID: p.ArmID,
		ParentKeep: "GENESIS", ChangedVariable: "none", ChangedPaths: []string{},
		ResultPath: resultRel, ResultSHA256: resultHash, Provenance: p,
		Budgets: run.ActualBudget, Status: "invalid", Reason: "diagnostic_unadmitted",
		Evidence: []string{reportRel + "#sha256=" + reportHash, bench.ContextThresholdAuthority.CheckinRef + "#sha256=" + bench.ContextThresholdAuthority.CheckinSHA256},
	}
	if err := appendHistory(s.historyPath, record); err != nil {
		return fmt.Errorf("append verified history %d: %w", index, err)
	}
	records, err := readHistory(s.historyPath)
	if err != nil {
		return err
	}
	last := records[len(records)-1]
	for _, exact := range []struct{ path, hash string }{{resultPath, resultHash}, {reportPath, reportHash}} {
		got, err := fileSHA256(exact.path)
		if err != nil || got != exact.hash {
			return fmt.Errorf("durable artifact hash mismatch for %s", exact.path)
		}
	}
	s.output.Runs = append(s.output.Runs, controllerRunRef{Mode: run.Mode, CaseID: run.CaseID, ResultPath: resultRel, ResultSHA256: resultHash, ReportPath: reportRel, ReportSHA256: reportHash, HistoryRecordHash: last.RecordHash, ActualBudget: run.ActualBudget, CandidateDeltaSHA256: p.CandidateDeltaSHA256, BundleSHA256: p.BundleSHA256})
	_, _, err = s.writeSnapshot()
	return err
}

func controllerRunStem(index int, run controllerRun) string {
	stem := fmt.Sprintf("%02d-%s", index+1, safePathComponent(run.Mode))
	if run.CaseID != "" {
		stem += "-" + safePathComponent(run.CaseID)
	}
	return stem
}

func completedRunProvenance(authority controllerAuthority, run controllerRun, index int) provenance {
	p := authority.Expected
	p.ExperimentID = fmt.Sprintf("%s-%02d", authority.Expected.ExperimentID, index+1)
	p.ArmID = run.Mode
	if run.CaseID != "" {
		p.ArmID += ":" + run.CaseID
	}
	p.LogicalDumpSHA256 = run.Result.Database.LogicalSHA256
	p.CorpusMode = run.Mode
	p.ProjectDirSHA256 = run.ProjectDirSHA256
	p.C3DirSHA256 = run.C3DirSHA256
	p.ContextThresholdAuthorityRecord = []byte(authority.ContextThresholdAuthorityRecord)
	return p
}

func (s *controllerPersistence) fail(stdout io.Writer, run controllerRun, index int, errorClass, status string, cause error, authority controllerAuthority, bench benchmarkConfig) error {
	failure, persistErr := s.persistFailure(run, index, errorClass, status, cause, authority, bench)
	durableSnapshotPath := filepath.Join(s.outputDir, filepath.FromSlash(s.lastSnapshotPath))
	if persistErr != nil {
		_ = s.emit(stdout)
		return fmt.Errorf("%s: %w; failure evidence persistence failed: %v; prior durable output: %s#sha256=%s", errorClass, cause, persistErr, durableSnapshotPath, s.lastSnapshotHash)
	}
	_ = s.emit(stdout)
	return fmt.Errorf("%s: %w; durable failure evidence: %s#sha256=%s history_record=%s", errorClass, cause, durableSnapshotPath, s.lastSnapshotHash, failure.HistoryRecordHash)
}

func (s *controllerPersistence) persistFailure(run controllerRun, index int, errorClass, status string, cause error, authority controllerAuthority, bench benchmarkConfig) (controllerFailureRef, error) {
	stem := controllerRunStem(index, run)
	rawRel := filepath.ToSlash(filepath.Join("failures", stem+"-raw.bin"))
	rawPath := filepath.Join(s.outputDir, filepath.FromSlash(rawRel))
	if err := writeExclusive(rawPath, run.RawResult); err != nil {
		return controllerFailureRef{}, err
	}
	rawHash, err := fileSHA256(rawPath)
	if err != nil {
		return controllerFailureRef{}, err
	}
	priorTail := "GENESIS"
	prior, err := readHistory(s.historyPath)
	if err != nil {
		return controllerFailureRef{}, err
	}
	if len(prior) > 0 {
		priorTail = prior[len(prior)-1].RecordHash
	}
	errorHash := shaString(cause.Error())
	p := completedRunProvenance(authority, run, index)
	evidence := controllerFailureEvidence{Schema: "structural-retrieval-controller-failure.v2", Mode: run.Mode, CaseID: run.CaseID, ErrorClass: errorClass, Error: cause.Error(), ErrorSHA256: errorHash, RawResultPath: rawRel, RawResultSHA256: rawHash, ActualBudget: run.ActualBudget, AccountingDiagnostics: run.AccountingDiagnostics, PriorHistoryTail: priorTail, Provenance: p}
	evidenceBytes, err := json.Marshal(evidence)
	if err != nil {
		return controllerFailureRef{}, err
	}
	evidenceBytes = append(evidenceBytes, '\n')
	evidenceRel := filepath.ToSlash(filepath.Join("failures", stem+".json"))
	evidencePath := filepath.Join(s.outputDir, filepath.FromSlash(evidenceRel))
	if err := writeExclusive(evidencePath, evidenceBytes); err != nil {
		return controllerFailureRef{}, err
	}
	evidenceHash, err := fileSHA256(evidencePath)
	if err != nil {
		return controllerFailureRef{}, err
	}
	record := historyRecord{
		Schema: historySchema, ExperimentID: p.ExperimentID, ArmID: p.ArmID,
		ParentKeep: priorTail, ChangedVariable: "none", ChangedPaths: []string{},
		ResultPath: rawRel, ResultSHA256: rawHash, Provenance: p, Budgets: run.ActualBudget,
		Status: status, Reason: "controller_" + errorClass, ErrorClass: errorClass, ErrorSHA256: errorHash,
		Evidence: []string{evidenceRel + "#sha256=" + evidenceHash, bench.ContextThresholdAuthority.CheckinRef + "#sha256=" + bench.ContextThresholdAuthority.CheckinSHA256},
	}
	if err := appendHistory(s.historyPath, record); err != nil {
		return controllerFailureRef{}, err
	}
	records, err := readHistory(s.historyPath)
	if err != nil {
		return controllerFailureRef{}, err
	}
	if err := verifyHistory(records); err != nil {
		return controllerFailureRef{}, err
	}
	last := records[len(records)-1]
	ref := controllerFailureRef{Mode: run.Mode, CaseID: run.CaseID, ErrorClass: errorClass, ErrorSHA256: errorHash, EvidencePath: evidenceRel, EvidenceSHA256: evidenceHash, HistoryRecordHash: last.RecordHash, ActualBudget: run.ActualBudget, CandidateDeltaSHA256: p.CandidateDeltaSHA256, BundleSHA256: p.BundleSHA256}
	s.output.Failure = &ref
	if _, _, err := s.writeSnapshot(); err != nil {
		return controllerFailureRef{}, err
	}
	return ref, nil
}
func safePathComponent(value string) string {
	var b strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-' || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteByte('_')
		}
	}
	if b.Len() == 0 {
		return "case"
	}
	return b.String()
}

func buildArmRequest(fixtures []fixtureCase, corpusMode, semanticMode string, bench benchmarkConfig) (armRequest, error) {
	if len(fixtures) == 0 {
		return armRequest{}, errors.New("no fixtures")
	}
	if corpusMode != corpusIsolated && corpusMode != corpusCombined {
		return armRequest{}, fmt.Errorf("invalid corpus mode %q", corpusMode)
	}
	if semanticMode != semanticDisabled && semanticMode != semanticDefault {
		return armRequest{}, fmt.Errorf("invalid semantic mode %q", semanticMode)
	}
	if err := validateAdapterImplementationConfig(bench); err != nil {
		return armRequest{}, err
	}
	selected := fixtures
	if corpusMode == corpusIsolated {
		selected = fixtures[:1]
	}
	req := armRequest{Schema: armRequestSchema, SemanticMode: semanticMode}
	seenEntity := map[string]bool{}
	seenRel := map[string]bool{}
	for _, fixture := range selected {
		if err := validateFixture(fixture); err != nil {
			return armRequest{}, err
		}
		req.Queries = append(req.Queries, armQuery{CaseID: fixture.CaseID, Query: fixture.Query})
		for _, e := range fixture.Corpus.Entities {
			if !seenEntity[e.ID] {
				req.Corpus.Entities = append(req.Corpus.Entities, e)
				seenEntity[e.ID] = true
			}
		}
		for _, rel := range fixture.Corpus.Relationships {
			key := rel.FromID + "\x00" + rel.ToID + "\x00" + rel.RelType
			if !seenRel[key] {
				req.Corpus.Relationships = append(req.Corpus.Relationships, rel)
				seenRel[key] = true
			}
		}
	}
	return req, nil
}

func runArm(req armRequest, dbPath, projectDir, c3Dir string) (controllerResult, error) {
	arm, err := executeArm(req, dbPath, projectDir, c3Dir)
	if err != nil {
		return controllerResult{}, err
	}
	return inspectArmResult(req, arm, dbPath)
}

// executeArm is the oracle-blind runtime half. It receives only the redacted
// request and returns only raw ordered search rows.
func executeArm(req armRequest, dbPath, projectDir, c3Dir string) (armResponse, error) {
	if req.Schema != armRequestSchema {
		return armResponse{}, errors.New("arm request schema mismatch")
	}
	for _, path := range []string{filepath.Dir(dbPath), projectDir, c3Dir} {
		if err := os.MkdirAll(path, 0o700); err != nil {
			return armResponse{}, err
		}
	}
	s, err := store.Open(dbPath)
	if err != nil {
		return armResponse{}, err
	}
	closed := false
	defer func() {
		if !closed {
			_ = s.Close()
		}
	}()
	if err := s.WithTx(func(ts *store.Store) error {
		if err := insertEntitiesParentFirst(ts, req.Corpus.Entities); err != nil {
			return err
		}
		for _, in := range req.Corpus.Entities {
			if strings.TrimSpace(in.Markdown) != "" {
				if err := content.WriteEntity(ts, in.ID, in.Markdown); err != nil {
					return err
				}
			}
		}
		for _, rel := range req.Corpus.Relationships {
			if err := ts.AddRelationship(&store.Relationship{FromID: rel.FromID, ToID: rel.ToID, RelType: rel.RelType}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return armResponse{}, err
	}
	arm := armResponse{Schema: armResponseSchema}
	for _, q := range req.Queries {
		var raw bytes.Buffer
		err = cmd.RunSearch(cmd.SearchOptions{Store: s, Query: q.Query, JSON: true, Limit: 5, NoSemantic: req.SemanticMode == semanticDisabled, ProjectDir: projectDir, C3Dir: c3Dir}, &raw)
		if err != nil {
			return armResponse{}, fmt.Errorf("RunSearch %s: %w", q.CaseID, err)
		}
		var output cmd.SearchOutput
		if err := decodeStrictBytes(raw.Bytes(), &output); err != nil {
			return armResponse{}, fmt.Errorf("RunSearch output %s: %w", q.CaseID, err)
		}
		arm.Cases = append(arm.Cases, armCaseResult{CaseID: q.CaseID, Query: output.Query, Rows: output.Results})
	}
	if _, err := s.DB().Exec(`PRAGMA wal_checkpoint(TRUNCATE)`); err != nil {
		return armResponse{}, err
	}
	if err := s.Close(); err != nil {
		return armResponse{}, err
	}
	closed = true
	return arm, nil
}

func insertEntitiesParentFirst(s *store.Store, inputs []entityInput) error {
	pending := append([]entityInput(nil), inputs...)
	inserted := map[string]bool{}
	for len(pending) > 0 {
		next := pending[:0]
		progress := false
		for _, in := range pending {
			if in.ParentID != "" && !inserted[in.ParentID] {
				next = append(next, in)
				continue
			}
			metadata := in.Metadata
			if metadata == "" {
				metadata = "{}"
			}
			status := in.Status
			if status == "" {
				status = "active"
			}
			e := &store.Entity{ID: in.ID, Type: in.Type, Title: in.Title, Slug: in.Slug, Category: in.Category, ParentID: in.ParentID, Goal: in.Goal, Status: status, Boundary: in.Boundary, Date: in.Date, Metadata: metadata}
			if err := s.InsertEntity(e); err != nil {
				return err
			}
			inserted[in.ID] = true
			progress = true
		}
		if !progress {
			return errors.New("entity parent graph has a missing parent or cycle")
		}
		pending = append([]entityInput(nil), next...)
	}
	return nil
}

func inspectArmResult(req armRequest, arm armResponse, dbPath string) (controllerResult, error) {
	encoded, err := json.Marshal(arm)
	if err != nil {
		return controllerResult{}, err
	}
	decoded, err := decodeArmResponse(bytes.NewReader(encoded), 16<<20)
	if err != nil {
		return controllerResult{}, err
	}
	s, err := store.Open(dbPath)
	if err != nil {
		return controllerResult{}, err
	}
	defer s.Close()
	result := controllerResult{Cases: decoded.Cases, DirectProbes: map[string]directProbes{}}
	for _, q := range req.Queries {
		probe, err := runDirectProbes(s, q.Query)
		if err != nil {
			return controllerResult{}, err
		}
		result.DirectProbes[q.CaseID] = probe
	}
	result.Database, err = inspectLogicalDump(s)
	if err != nil {
		return controllerResult{}, err
	}
	return result, nil
}

func runDirectProbes(s *store.Store, query string) (directProbes, error) {
	var out directProbes
	entities, err := s.SearchWithLimit(query, "", 20)
	if err != nil {
		return out, err
	}
	contentRows, err := s.SearchContent(query, 20)
	if err != nil {
		return out, err
	}
	for _, hit := range entities {
		out.EntityFTSIDs = append(out.EntityFTSIDs, hit.ID)
	}
	for _, hit := range contentRows {
		out.ContentFTSIDs = append(out.ContentFTSIDs, hit.ID)
	}
	return out, nil
}

func decodeArmResponse(r io.Reader, maxBytes int64) (armResponse, error) {
	var out armResponse
	if maxBytes <= 0 {
		return out, errors.New("invalid output limit")
	}
	limited := io.LimitReader(r, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return out, err
	}
	if int64(len(data)) > maxBytes {
		return out, errors.New("arm output exceeds limit")
	}
	if err := decodeStrictBytes(data, &out); err != nil {
		return out, err
	}
	if out.Schema != armResponseSchema {
		return out, errors.New("arm response schema mismatch")
	}
	return out, nil
}

func decodeArmRequest(r io.Reader, maxBytes int64) (armRequest, error) {
	var out armRequest
	if maxBytes <= 0 {
		return out, errors.New("invalid input limit")
	}
	data, err := io.ReadAll(io.LimitReader(r, maxBytes+1))
	if err != nil {
		return out, err
	}
	if int64(len(data)) > maxBytes {
		return out, errors.New("arm input exceeds limit")
	}
	if err := decodeStrictBytes(data, &out); err != nil {
		return out, err
	}
	if out.Schema != armRequestSchema {
		return out, errors.New("arm request schema mismatch")
	}
	return out, nil
}

func decodeStrictBytes(data []byte, dst any) error {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return fmt.Errorf("trailing bytes: %w", err)
	}
	return nil
}

func rejectDuplicateJSONKeys(data []byte) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var parseValue func() error
	parseValue = func() error {
		token, err := dec.Token()
		if err != nil {
			return err
		}
		delim, ok := token.(json.Delim)
		if !ok {
			return nil
		}
		switch delim {
		case '{':
			seen := map[string]bool{}
			for dec.More() {
				keyToken, err := dec.Token()
				if err != nil {
					return err
				}
				key, ok := keyToken.(string)
				if !ok || seen[key] {
					return errors.New("duplicate or invalid JSON object key")
				}
				seen[key] = true
				if err := parseValue(); err != nil {
					return err
				}
			}
			closing, err := dec.Token()
			if err != nil || closing != json.Delim('}') {
				return errors.New("invalid JSON object")
			}
		case '[':
			for dec.More() {
				if err := parseValue(); err != nil {
					return err
				}
			}
			closing, err := dec.Token()
			if err != nil || closing != json.Delim(']') {
				return errors.New("invalid JSON array")
			}
		default:
			return errors.New("invalid JSON delimiter")
		}
		return nil
	}
	if err := parseValue(); err != nil {
		return err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func inspectLogicalDump(s *store.Store) (logicalDump, error) {
	d := logicalDump{}
	if err := s.DB().QueryRow(`PRAGMA integrity_check`).Scan(&d.Integrity); err != nil {
		return d, err
	}
	rows, err := s.DB().Query(`SELECT sql FROM sqlite_master WHERE sql IS NOT NULL ORDER BY type, name`)
	if err != nil {
		return d, err
	}
	for rows.Next() {
		var sql string
		if err := rows.Scan(&sql); err != nil {
			rows.Close()
			return d, err
		}
		d.SchemaSQL = append(d.SchemaSQL, sql)
	}
	if err := rows.Close(); err != nil {
		return d, err
	}
	entities, err := s.AllEntities()
	if err != nil {
		return d, err
	}
	for _, e := range entities {
		d.Entities = append(d.Entities, logicalEntity{e.ID, e.Type, e.Title, e.Slug, e.ParentID, e.Goal, e.Status, e.Metadata, e.RootMerkle, e.Version})
	}
	rr, err := s.DB().Query(`SELECT from_id, to_id, rel_type FROM relationships ORDER BY from_id, to_id, rel_type`)
	if err != nil {
		return d, err
	}
	for rr.Next() {
		var x logicalRelationship
		if err := rr.Scan(&x.FromID, &x.ToID, &x.RelType); err != nil {
			rr.Close()
			return d, err
		}
		d.Relationships = append(d.Relationships, x)
	}
	if err := rr.Close(); err != nil {
		return d, err
	}
	nr, err := s.DB().Query(`SELECT entity_id, COALESCE(parent_id,0), type, level, seq, content, hash FROM nodes ORDER BY entity_id, id`)
	if err != nil {
		return d, err
	}
	for nr.Next() {
		var x logicalNode
		if err := nr.Scan(&x.EntityID, &x.ParentID, &x.Type, &x.Level, &x.Seq, &x.Content, &x.Hash); err != nil {
			nr.Close()
			return d, err
		}
		d.Nodes = append(d.Nodes, x)
	}
	if err := nr.Close(); err != nil {
		return d, err
	}
	d.EntityCount = len(d.Entities)
	d.RelationshipCount = len(d.Relationships)
	d.SQLiteRowCount = len(d.Entities) + len(d.Relationships) + len(d.Nodes)
	canonical := d
	canonical.LogicalSHA256 = ""
	canonical.LogicalBytes = 0
	data, err := json.Marshal(canonical)
	if err != nil {
		return d, err
	}
	d.LogicalBytes = len(data)
	sum := sha256.Sum256(data)
	d.LogicalSHA256 = hex.EncodeToString(sum[:])
	return d, nil
}

func validateLogicalDump(d logicalDump) error {
	if d.Integrity != "ok" || d.EntityCount != len(d.Entities) || d.RelationshipCount != len(d.Relationships) || d.SQLiteRowCount != len(d.Entities)+len(d.Relationships)+len(d.Nodes) {
		return errors.New("invalid logical dump counts or integrity")
	}
	for index := 1; index < len(d.Entities); index++ {
		before, after := d.Entities[index-1], d.Entities[index]
		if before.Type > after.Type || before.Type == after.Type && before.ID >= after.ID {
			return errors.New("logical dump entities are not canonical")
		}
	}
	for index := 1; index < len(d.Relationships); index++ {
		before, after := d.Relationships[index-1], d.Relationships[index]
		beforeKey := before.FromID + "\x00" + before.ToID + "\x00" + before.RelType
		afterKey := after.FromID + "\x00" + after.ToID + "\x00" + after.RelType
		if beforeKey >= afterKey {
			return errors.New("logical dump relationships are not canonical")
		}
	}
	canonical := d
	canonical.LogicalSHA256 = ""
	canonical.LogicalBytes = 0
	data, err := json.Marshal(canonical)
	if err != nil {
		return err
	}
	if d.LogicalBytes != len(data) || d.LogicalSHA256 != shaString(string(data)) {
		return errors.New("logical dump canonical hash mismatch")
	}
	return nil
}

func setRawDiagnosticMarker(path string, marker byte) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteAt([]byte{0, 0, 0, marker}, 68)
	return err
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func validateFixture(f fixtureCase) error {
	if f.Schema != fixtureSchema || strings.TrimSpace(f.CaseID) == "" || strings.TrimSpace(f.Query) == "" {
		return errors.New("invalid fixture identity")
	}
	if f.Family != familyWrongLayer && f.Family != familyRoute {
		return fmt.Errorf("invalid fixture family %q", f.Family)
	}
	ids := map[string]bool{}
	for _, e := range f.Corpus.Entities {
		if e.ID == "" || e.Type == "" || e.Title == "" || ids[e.ID] {
			return errors.New("invalid or duplicate fixture entity")
		}
		ids[e.ID] = true
	}
	for entityID := range f.Oracle.FactBindings {
		if !ids[entityID] {
			return fmt.Errorf("fact binding references missing entity %s", entityID)
		}
	}
	if f.Family == familyRoute {
		return validateRouteFixture(f)
	}
	if len(f.Oracle.RequiredOwnerFactIDs) == 0 {
		return errors.New("wrong-layer fixture has no required owner facts")
	}
	return nil
}

func validateRouteFixture(f fixtureCase) error {
	w := f.Oracle.RelationshipWitness
	if w == nil || w.ExpectedEntityID == "" || w.FromID == "" || w.ToID == "" || w.RelType == "" {
		return errors.New("route fixture has no complete witness")
	}
	if w.ExpectedEntityID != w.FromID || w.ExpectedMatchSource != "graph:"+w.RelType+":"+w.ToID {
		return errors.New("route witness does not match C3 graph tag contract")
	}
	found := false
	for _, rel := range f.Corpus.Relationships {
		if rel.FromID == w.FromID && rel.ToID == w.ToID && rel.RelType == w.RelType {
			found = true
		}
	}
	if !found {
		return errors.New("route witness absent from controller corpus")
	}
	return nil
}

func validateExpansionSpecificFixture(f fixtureCase) error {
	if err := validateRouteFixture(f); err != nil {
		return err
	}
	w := f.Oracle.RelationshipWitness
	if !w.RequireDirectFTSMiss {
		return errors.New("expansion-specific fixture does not require direct FTS miss")
	}
	queryTerms := terms(f.Query)
	for _, e := range f.Corpus.Entities {
		if e.ID != w.ExpectedEntityID {
			continue
		}
		ownerTerms := terms(strings.Join([]string{e.Title, e.Slug, e.Goal, e.Markdown}, " "))
		for term := range queryTerms {
			if ownerTerms[term] {
				return fmt.Errorf("expected route owner has direct lexical term %q", term)
			}
		}
		return nil
	}
	return errors.New("expected route owner missing")
}

func terms(value string) map[string]bool {
	out := map[string]bool{}
	for _, field := range strings.FieldsFunc(strings.ToLower(value), func(r rune) bool { return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9') }) {
		if field != "" {
			out[field] = true
		}
	}
	return out
}

func relationshipRouteHit(f fixtureCase, result armCaseResult, dump logicalDump) bool {
	w := f.Oracle.RelationshipWitness
	if w == nil {
		return false
	}
	witness := false
	for _, rel := range dump.Relationships {
		if rel.FromID == w.FromID && rel.ToID == w.ToID && rel.RelType == w.RelType {
			witness = true
			break
		}
	}
	if !witness {
		return false
	}
	for i, row := range result.Rows {
		if i >= 5 {
			break
		}
		if row.ID == w.ExpectedEntityID && containsString(row.MatchSources, w.ExpectedMatchSource) {
			return true
		}
	}
	return false
}

type caseMetrics struct {
	CaseID                            string             `json:"case_id"`
	Family                            string             `json:"family"`
	OwnerRecallAt1                    float64            `json:"owner_recall_at_1"`
	OwnerRecallAt3                    float64            `json:"owner_recall_at_3"`
	OwnerRecallAt5                    float64            `json:"owner_recall_at_5"`
	MRR                               float64            `json:"mrr"`
	StructuralOwnerPrecision          float64            `json:"structural_owner_precision"`
	ForbiddenStructuralRetrievalCount int                `json:"forbidden_structural_retrieval_count"`
	RelationshipRouteHit              bool               `json:"relationship_route_hit"`
	ExpansionSpecificDirectMiss       bool               `json:"expansion_specific_direct_miss"`
	RelationshipRouteMRR              float64            `json:"relationship_route_mrr"`
	CanonicalRowBytes                 int                `json:"canonical_row_bytes"`
	RowCount                          int                `json:"row_count"`
	RouteCoverage                     map[string]float64 `json:"route_coverage"`
}

type metrics struct {
	OwnerRecallAt1                    float64            `json:"owner_recall_at_1"`
	OwnerRecallAt3                    float64            `json:"owner_recall_at_3"`
	OwnerRecallAt5                    float64            `json:"owner_recall_at_5"`
	MRR                               float64            `json:"mrr"`
	WrongLayerMRR                     float64            `json:"wrong_layer_mrr"`
	StructuralOwnerPrecision          float64            `json:"structural_owner_precision"`
	ForbiddenStructuralRetrievalCount int                `json:"forbidden_structural_retrieval_count"`
	RelationshipRouteRecallAt5        float64            `json:"relationship_route_recall_at_5"`
	RelationshipRouteMRR              float64            `json:"relationship_route_mrr"`
	CanonicalRowBytesPerCase          map[string]int     `json:"canonical_row_bytes_per_case"`
	RouteCoverage                     map[string]float64 `json:"route_coverage"`
}

type evalReport struct {
	Schema     string        `json:"$schema"`
	CorpusMode string        `json:"corpus_mode"`
	Metrics    metrics       `json:"metrics"`
	Cases      []caseMetrics `json:"cases"`
	Provenance *provenance   `json:"provenance,omitempty"`
}
type gateVerdict struct {
	Keep                bool     `json:"keep"`
	WouldPassIfRatified bool     `json:"would_pass_if_ratified"`
	Status              string   `json:"status"`
	Failed              []string `json:"failed"`
	RecallDelta         float64  `json:"owner_recall_at_5_delta"`
}

func scoreArmResponse(fixtures []fixtureCase, response armResponse, dump logicalDump, corpusMode string) (evalReport, error) {
	return scoreArmResponseWithProbes(fixtures, response, dump, corpusMode, nil)
}

func scoreArmResponseWithProbes(fixtures []fixtureCase, response armResponse, dump logicalDump, corpusMode string, probes map[string]directProbes) (evalReport, error) {
	if response.Schema != armResponseSchema {
		return evalReport{}, errors.New("response schema mismatch")
	}
	byID := map[string]armCaseResult{}
	for _, c := range response.Cases {
		if _, exists := byID[c.CaseID]; exists {
			return evalReport{}, fmt.Errorf("duplicate case %s", c.CaseID)
		}
		byID[c.CaseID] = c
	}
	report := evalReport{Schema: reportSchema, CorpusMode: corpusMode}
	for _, fixture := range fixtures {
		result, ok := byID[fixture.CaseID]
		if !ok {
			return evalReport{}, fmt.Errorf("missing case %s", fixture.CaseID)
		}
		probe, probePresent := probes[fixture.CaseID]
		cm := scoreCase(fixture, result, dump, probe, probePresent)
		report.Cases = append(report.Cases, cm)
	}
	report.Metrics = summarizeMetrics(report.Cases)
	return report, nil
}

func scoreCase(f fixtureCase, result armCaseResult, dump logicalDump, probe directProbes, probePresent bool) caseMetrics {
	rows := result.Rows
	cm := caseMetrics{CaseID: f.CaseID, Family: f.Family, CanonicalRowBytes: canonicalRowBytes(rows), RowCount: len(rows), RouteCoverage: map[string]float64{}}
	cm.OwnerRecallAt1 = recallAt(rows, 1, f.Oracle)
	cm.OwnerRecallAt3 = recallAt(rows, 3, f.Oracle)
	cm.OwnerRecallAt5 = recallAt(rows, 5, f.Oracle)
	cm.MRR = reciprocalRank(rows, f.Oracle.RequiredOwnerFactIDs, f.Oracle.FactBindings)
	required := stringSet(f.Oracle.RequiredOwnerFactIDs)
	allowed := stringSet(f.Oracle.AllowedExtraFactIDs)
	forbidden := stringSet(f.Oracle.ForbiddenFactIDs)
	seenStructural := map[string]bool{}
	seenRequired := map[string]bool{}
	seenForbidden := map[string]bool{}
	for i, row := range rows {
		if i >= 5 {
			break
		}
		for _, fact := range f.Oracle.FactBindings[row.ID] {
			if required[fact] {
				seenRequired[fact] = true
				seenStructural[fact] = true
			}
			if allowed[fact] {
				seenStructural[fact] = true
			}
			if forbidden[fact] {
				seenForbidden[fact] = true
				seenStructural[fact] = true
			}
		}
	}
	if len(seenStructural) > 0 {
		cm.StructuralOwnerPrecision = float64(len(seenRequired)) / float64(len(seenStructural))
	}
	cm.ForbiddenStructuralRetrievalCount = len(seenForbidden)
	if f.Oracle.RelationshipWitness != nil {
		cm.ExpansionSpecificDirectMiss = true
		if f.Oracle.RelationshipWitness.RequireDirectFTSMiss {
			expected := f.Oracle.RelationshipWitness.ExpectedEntityID
			cm.ExpansionSpecificDirectMiss = probePresent && !containsString(probe.EntityFTSIDs, expected) && !containsString(probe.ContentFTSIDs, expected)
		}
		cm.RelationshipRouteHit = cm.ExpansionSpecificDirectMiss && relationshipRouteHit(f, result, dump)
		if cm.RelationshipRouteHit {
			for i, row := range rows {
				if i >= 5 {
					break
				}
				if row.ID == f.Oracle.RelationshipWitness.ExpectedEntityID {
					cm.RelationshipRouteMRR = 1 / float64(i+1)
					break
				}
			}
		}
	}
	for _, field := range f.Oracle.RequiredRouteFields {
		cm.RouteCoverage[field] = routeFieldCoverage(rows, field)
	}
	return cm
}

func routeFieldCoverage(rows []cmd.SearchResultRow, field string) float64 {
	if len(rows) == 0 {
		return 0
	}
	hits := 0
	for _, row := range rows {
		present := false
		switch field {
		case "facts":
			present = len(row.Route.Facts) > 0
		case "graph":
			present = len(row.Route.Graph) > 0
		case "anchors":
			present = len(row.Route.Anchors) > 0
		case "lanes":
			present = len(row.Route.Lanes) > 0
		case "drift":
			present = len(row.Route.Drift) > 0
		case "hash":
			present = row.Route.Hash != ""
		}
		if present {
			hits++
		}
	}
	return float64(hits) / float64(len(rows))
}

func summarizeMetrics(cases []caseMetrics) metrics {
	m := metrics{CanonicalRowBytesPerCase: map[string]int{}, RouteCoverage: map[string]float64{}}
	if len(cases) == 0 {
		return m
	}
	wrong, route := 0, 0
	for _, c := range cases {
		m.StructuralOwnerPrecision += c.StructuralOwnerPrecision
		m.ForbiddenStructuralRetrievalCount += c.ForbiddenStructuralRetrievalCount
		m.CanonicalRowBytesPerCase[c.CaseID] = c.CanonicalRowBytes
		for k, v := range c.RouteCoverage {
			m.RouteCoverage[k] += v
		}
		if c.Family == familyWrongLayer {
			m.OwnerRecallAt1 += c.OwnerRecallAt1
			m.OwnerRecallAt3 += c.OwnerRecallAt3
			m.OwnerRecallAt5 += c.OwnerRecallAt5
			m.MRR += c.MRR
			m.WrongLayerMRR += c.MRR
			wrong++
		}
		if c.Family == familyRoute {
			if c.RelationshipRouteHit {
				m.RelationshipRouteRecallAt5++
			}
			m.RelationshipRouteMRR += c.RelationshipRouteMRR
			route++
		}
	}
	n := float64(len(cases))
	m.StructuralOwnerPrecision /= n
	for k := range m.RouteCoverage {
		m.RouteCoverage[k] /= n
	}
	if wrong > 0 {
		m.OwnerRecallAt1 /= float64(wrong)
		m.OwnerRecallAt3 /= float64(wrong)
		m.OwnerRecallAt5 /= float64(wrong)
		m.MRR /= float64(wrong)
		m.WrongLayerMRR /= float64(wrong)
	}
	if route > 0 {
		m.RelationshipRouteRecallAt5 /= float64(route)
		m.RelationshipRouteMRR /= float64(route)
	}
	return m
}

func recallAt(rows []cmd.SearchResultRow, k int, oracle oracleSpec) float64 {
	req := stringSet(oracle.RequiredOwnerFactIDs)
	if len(req) == 0 {
		return 0
	}
	seen := map[string]bool{}
	for i, row := range rows {
		if i >= k {
			break
		}
		for _, f := range oracle.FactBindings[row.ID] {
			if req[f] {
				seen[f] = true
			}
		}
	}
	return float64(len(seen)) / float64(len(req))
}
func reciprocalRank(rows []cmd.SearchResultRow, required []string, bindings map[string][]string) float64 {
	req := stringSet(required)
	for i, row := range rows {
		for _, f := range bindings[row.ID] {
			if req[f] {
				return 1 / float64(i+1)
			}
		}
	}
	return 0
}
func stringSet(values []string) map[string]bool {
	m := map[string]bool{}
	for _, v := range values {
		m[v] = true
	}
	return m
}
func canonicalRowBytes(rows []cmd.SearchResultRow) int {
	data, _ := json.Marshal(rows)
	return len(data)
}

func verifyReport(fixtures []fixtureCase, response armResponse, dump logicalDump, mode string, reported evalReport) error {
	return verifyReportWithProbes(fixtures, response, dump, mode, nil, reported)
}

func verifyReportWithProbes(fixtures []fixtureCase, response armResponse, dump logicalDump, mode string, probes map[string]directProbes, reported evalReport) error {
	want, err := scoreArmResponseWithProbes(fixtures, response, dump, mode, probes)
	if err != nil {
		return err
	}
	reported.Provenance = nil
	if !reflect.DeepEqual(want, reported) {
		return errors.New("report does not match controller recomputation")
	}
	return nil
}

func evaluateGate(baseline, candidate evalReport, bench benchmarkConfig) gateVerdict {
	v := gateVerdict{Status: decisionDiscard, RecallDelta: candidate.Metrics.OwnerRecallAt5 - baseline.Metrics.OwnerRecallAt5}
	fail := func(reason string) { v.Keep = false; v.Status = decisionDiscard; v.Failed = append(v.Failed, reason) }
	if v.RecallDelta+1e-12 < bench.Thresholds.OwnerRecallAt5Delta {
		fail("owner recall delta")
	}
	if candidate.Metrics.StructuralOwnerPrecision+1e-12 < bench.Thresholds.StructuralOwnerPrecision {
		fail("structural precision")
	}
	if candidate.Metrics.ForbiddenStructuralRetrievalCount != 0 {
		fail("forbidden structural retrieval")
	}
	if candidate.Metrics.RelationshipRouteRecallAt5 < baseline.Metrics.RelationshipRouteRecallAt5 {
		fail("route recall")
	}
	if candidate.Metrics.RelationshipRouteMRR < baseline.Metrics.RelationshipRouteMRR {
		fail("route MRR")
	}
	if candidate.Metrics.WrongLayerMRR < baseline.Metrics.WrongLayerMRR {
		fail("wrong-layer MRR")
	}
	for id, b := range baseline.Metrics.CanonicalRowBytesPerCase {
		c, ok := candidate.Metrics.CanonicalRowBytesPerCase[id]
		if !ok || b == 0 || float64(c) > float64(b)*bench.Thresholds.CanonicalRowBytesRatio+1e-12 {
			fail("canonical row bytes " + id)
		}
	}
	for field, b := range baseline.Metrics.RouteCoverage {
		if candidate.Metrics.RouteCoverage[field] < b {
			fail("route field " + field)
		}
	}
	if len(v.Failed) == 0 {
		v.WouldPassIfRatified = true
		v.Status = "would_pass_if_ratified"
	}
	return v
}

func generateScaleCorpus(base corpusInput, cfg scaleConfig) corpusInput {
	out := corpusInput{Entities: append([]entityInput(nil), base.Entities...), Relationships: append([]relationshipInput(nil), base.Relationships...)}
	mult := cfg.Multiplier
	if mult < 1 {
		mult = 1
	}
	tokens := cfg.Tokens
	if len(tokens) == 0 {
		tokens = []string{"generic"}
	}
	for i := 0; i < len(base.Entities)*(mult-1); i++ {
		token := tokens[(cfg.Seed+i)%len(tokens)]
		out.Entities = append(out.Entities, entityInput{ID: fmt.Sprintf("scale-%d-%d", cfg.Seed, i), Type: "component", Title: strings.Title(token) + " Synthetic Decoy", Slug: fmt.Sprintf("%s-decoy-%d", token, i), Goal: "Generic deterministic scale record", Status: "active", Metadata: "{}", Markdown: "# Synthetic Decoy\n\nGeneric deterministic scale record."})
	}
	return out
}

type provenance struct {
	ExperimentID                    string `json:"experiment_id"`
	ArmID                           string `json:"arm_id"`
	Commit                          string `json:"commit"`
	Tree                            string `json:"tree"`
	ControllerCommit                string `json:"controller_commit,omitempty"`
	ControllerTree                  string `json:"controller_tree,omitempty"`
	ControllerSourceCapsuleSHA256   string `json:"controller_source_capsule_sha256,omitempty"`
	CandidateDeltaSHA256            string `json:"candidate_delta_sha256,omitempty"`
	BundleSHA256                    string `json:"bundle_sha256,omitempty"`
	BenchmarkSHA256                 string `json:"benchmark_sha256,omitempty"`
	SourceCapsuleSHA256             string `json:"source_capsule_sha256"`
	DiffSHA256                      string `json:"diff_sha256"`
	FixtureSHA256                   string `json:"fixture_sha256"`
	ScorerSHA256                    string `json:"scorer_sha256"`
	ControllerSHA256                string `json:"controller_sha256"`
	RuntimeSHA256                   string `json:"runtime_sha256"`
	LogicalDumpSHA256               string `json:"logical_dump_sha256"`
	EnvironmentSHA256               string `json:"environment_sha256"`
	ModuleGraphSHA256               string `json:"module_graph_sha256"`
	BudgetSHA256                    string `json:"budget_sha256"`
	ActionEnvelopeSHA256            string `json:"action_envelope_sha256"`
	CorpusMode                      string `json:"corpus_mode"`
	SemanticMode                    string `json:"semantic_mode"`
	ProjectDirSHA256                string `json:"project_dir_sha256"`
	C3DirSHA256                     string `json:"c3_dir_sha256"`
	ContextThresholdAuthoritySHA256 string `json:"context_threshold_authority_sha256"`
	PrivacyPolicySHA256             string `json:"privacy_policy_sha256,omitempty"`
	PrivacyTermCount                int    `json:"privacy_term_count,omitempty"`
	PrivacyDetectorSHA256           string `json:"privacy_detector_sha256,omitempty"`
	GoExecutableSHA256              string `json:"go_executable_sha256,omitempty"`
	GoModVerifySHA256               string `json:"go_mod_verify_sha256,omitempty"`
	ScanCapsSHA256                  string `json:"scan_caps_sha256,omitempty"`
	SourceBundleHeadsSHA256         string `json:"source_bundle_heads_sha256,omitempty"`
	ProtocolTestSHA256              string `json:"protocol_test_sha256,omitempty"`
	ParentBaselineOutputSHA256      string `json:"parent_baseline_output_sha256,omitempty"`
	ParentBaselineAuthoritySHA256   string `json:"parent_baseline_authority_sha256,omitempty"`
	ParentBaselineOrderedRunSHA256  string `json:"parent_baseline_ordered_run_manifest_sha256,omitempty"`
	ParentBaselineRunCount          int    `json:"parent_baseline_run_count,omitempty"`
	ParentBaselineHistorySHA256     string `json:"parent_baseline_history_sha256,omitempty"`
	ParentBaselineHistoryTailSHA256 string `json:"parent_baseline_history_tail_record_hash,omitempty"`
	ParentBaselinePrivacySHA256     string `json:"parent_baseline_privacy_manifest_sha256,omitempty"`
	ParentBaselineValidatorHash     string `json:"parent_baseline_validator_record_hash,omitempty"`
	ParentBaselineValidatorPayload  string `json:"parent_baseline_validator_payload_sha256,omitempty"`
	ContextThresholdAuthorityRecord []byte `json:"-"`
}

type controllerAuthority struct {
	Schema                          string              `json:"$schema"`
	Expected                        provenance          `json:"expected_provenance"`
	SourceCapsule                   sourceCapsule       `json:"source_capsule"`
	BuildReplay                     buildReplayManifest `json:"build_replay"`
	BudgetLimits                    resourceBudget      `json:"budget_limits"`
	ActionEnvelope                  string              `json:"action_envelope"`
	CanonicalRowBytesDefinition     string              `json:"canonical_row_bytes_definition"`
	ContextThresholdAuthorityRecord string              `json:"context_threshold_authority_record"`
}

type candidateDelta struct {
	Variable          string            `json:"variable"`
	BaselineCommit    string            `json:"baseline_commit"`
	BaselineTree      string            `json:"baseline_tree"`
	CandidateCommit   string            `json:"candidate_commit"`
	CandidateTree     string            `json:"candidate_tree"`
	DiffSHA256        string            `json:"diff_sha256"`
	NameStatusSHA256  string            `json:"name_status_sha256"`
	NameStatus        []string          `json:"name_status"`
	AllowedPaths      []string          `json:"allowed_paths"`
	BeforeBlobSHA256  map[string]string `json:"before_blob_sha256"`
	AfterBlobSHA256   map[string]string `json:"after_blob_sha256"`
	BundleSHA256      string            `json:"bundle_sha256"`
	BundleHeadsSHA256 string            `json:"bundle_heads_sha256"`
}

type controllerAuthorityV3 struct {
	Schema                          string              `json:"$schema"`
	Mode                            string              `json:"mode"`
	Expected                        provenance          `json:"expected_provenance"`
	ControllerSourceCapsule         sourceCapsule       `json:"controller_source_capsule"`
	RuntimeSourceCapsule            sourceCapsule       `json:"runtime_source_capsule"`
	CandidateDelta                  *candidateDelta     `json:"candidate_delta,omitempty"`
	BuildReplay                     buildReplayManifest `json:"build_replay"`
	BudgetLimits                    resourceBudget      `json:"budget_limits"`
	ActionEnvelope                  string              `json:"action_envelope"`
	CanonicalRowBytesDefinition     string              `json:"canonical_row_bytes_definition"`
	ContextThresholdAuthorityRecord string              `json:"context_threshold_authority_record"`
}

type privacyScanCaps struct {
	BundleFileBytesMax               int64 `json:"bundle_file_bytes_max"`
	SourceObjectCountMax             int   `json:"source_object_count_max"`
	SourceObjectUncompressedBytesMax int64 `json:"source_object_uncompressed_bytes_total_max"`
	SingleBlobBytesMax               int64 `json:"single_blob_bytes_max"`
	SingleCommitOrTagBytesMax        int64 `json:"single_commit_or_tag_bytes_max"`
	TreePathCountMax                 int   `json:"tree_path_count_max"`
	SinglePathUTF8BytesMax           int   `json:"single_path_utf8_bytes_max"`
	DurableArtifactCountMax          int   `json:"durable_artifact_count_max"`
	SingleDurableArtifactBytesMax    int64 `json:"single_durable_artifact_bytes_max"`
	DurableArtifactBytesTotalMax     int64 `json:"durable_artifact_bytes_total_max"`
	PreflightWallTimeMillis          int64 `json:"preflight_wall_time_millis"`
}

type parentBaselineBinding struct {
	AuthoritySHA256        string `json:"authority_sha256"`
	OutputSHA256           string `json:"output_sha256"`
	OrderedRunSHA256       string `json:"ordered_run_manifest_sha256"`
	RunCount               int    `json:"run_count"`
	HistorySHA256          string `json:"history_sha256"`
	HistoryTailRecordHash  string `json:"history_tail_record_hash"`
	PrivacyManifestSHA256  string `json:"privacy_manifest_sha256"`
	ValidatorRecordRef     string `json:"validator_record_ref"`
	ValidatorRecordHash    string `json:"validator_record_hash"`
	ValidatorPayloadSHA256 string `json:"validator_payload_sha256"`
}

type parentBaselineFiles struct {
	Root           string
	Authority      string
	Output         string
	ValidatorStore string
}

type baselineAcceptance struct {
	Schema                    string `json:"$schema"`
	Verdict                   string `json:"verdict"`
	AuthoritySHA256           string `json:"authority_sha256"`
	OutputSHA256              string `json:"output_sha256"`
	OrderedRunManifestSHA256  string `json:"ordered_run_manifest_sha256"`
	RunCount                  int    `json:"run_count"`
	HistorySHA256             string `json:"history_sha256"`
	HistoryTailRecordHash     string `json:"history_tail_record_hash"`
	PrivacyManifestSHA256     string `json:"privacy_manifest_sha256"`
	ValidatedSourceMainSHA256 string `json:"validated_source_main_sha256"`
	ValidatedSourceTestSHA256 string `json:"validated_source_test_sha256"`
}

type baselineValidatorPayload struct {
	Event              string             `json:"event"`
	WorkerID           string             `json:"worker_id"`
	Role               string             `json:"role"`
	Status             string             `json:"status"`
	EffectClaim        bool               `json:"effect_claim"`
	BaselineAcceptance baselineAcceptance `json:"baseline_acceptance"`
}

type controllerAuthorityV4 struct {
	Schema                          string                 `json:"$schema"`
	Mode                            string                 `json:"mode"`
	Expected                        provenance             `json:"expected_provenance"`
	ControllerSourceCapsule         sourceCapsule          `json:"controller_source_capsule"`
	RuntimeSourceCapsule            sourceCapsule          `json:"runtime_source_capsule"`
	CandidateDelta                  *candidateDelta        `json:"candidate_delta,omitempty"`
	ParentBaseline                  *parentBaselineBinding `json:"parent_baseline,omitempty"`
	BuildReplay                     buildReplayManifest    `json:"build_replay"`
	BudgetLimits                    resourceBudget         `json:"budget_limits"`
	ScanCaps                        privacyScanCaps        `json:"scan_caps"`
	PrivacyPolicySHA256             string                 `json:"privacy_policy_sha256"`
	PrivacyTermCount                int                    `json:"privacy_term_count"`
	PrivacyDetectorSHA256           string                 `json:"privacy_detector_sha256"`
	GoExecutableSHA256              string                 `json:"go_executable_sha256"`
	GoModVerifySHA256               string                 `json:"go_mod_verify_sha256"`
	SourceBundleHeadsSHA256         string                 `json:"source_bundle_heads_sha256"`
	ProtocolTestSHA256              string                 `json:"protocol_test_sha256"`
	ActionEnvelope                  string                 `json:"action_envelope"`
	CanonicalRowBytesDefinition     string                 `json:"canonical_row_bytes_definition"`
	ContextThresholdAuthorityRecord string                 `json:"context_threshold_authority_record"`
}

func protocolV7ScanCaps() privacyScanCaps {
	return privacyScanCaps{
		BundleFileBytesMax: 67_108_864, SourceObjectCountMax: 50_000,
		SourceObjectUncompressedBytesMax: 134_217_728, SingleBlobBytesMax: 16_777_216,
		SingleCommitOrTagBytesMax: 1_048_576, TreePathCountMax: 100_000, SinglePathUTF8BytesMax: 4_096,
		DurableArtifactCountMax: 2_048, SingleDurableArtifactBytesMax: 16_777_216,
		DurableArtifactBytesTotalMax: 268_435_456, PreflightWallTimeMillis: 180_000,
	}
}

func (a controllerAuthorityV4) persistenceAuthority() controllerAuthority {
	return controllerAuthority{
		Schema: controllerAuthoritySchema, Expected: a.Expected, SourceCapsule: a.RuntimeSourceCapsule,
		BuildReplay: a.BuildReplay, BudgetLimits: a.BudgetLimits, ActionEnvelope: a.ActionEnvelope,
		CanonicalRowBytesDefinition:     a.CanonicalRowBytesDefinition,
		ContextThresholdAuthorityRecord: a.ContextThresholdAuthorityRecord,
	}
}

const (
	privacyPolicySchema       = "structural-retrieval-privacy-policy.v1"
	privacyDetectorVersion    = "structural-retrieval-generic-privacy-detector.v1"
	privacyPolicyBytesMax     = 128 << 10
	privacyPolicyTermsMax     = 512
	privacyPolicyTermBytesMax = 64 << 10
)

type privacyPolicy struct {
	Schema    string   `json:"$schema"`
	DenyTerms []string `json:"deny_terms"`
	SHA256    string   `json:"-"`
}

type privacyPattern struct {
	ID      string `json:"id"`
	Pattern string `json:"re2"`
}

type compiledPrivacyPattern struct {
	ID string
	RE *regexp.Regexp
}

type genericPrivacyDetector struct {
	Version          string
	DefinitionSHA256 string
	patterns         []compiledPrivacyPattern
}

type privacyArtifact struct {
	Role   string `json:"role"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int    `json:"bytes"`
}

func privacyArtifactLess(left, right privacyArtifact) bool {
	leftSource := strings.HasPrefix(left.Role, "source_object_")
	rightSource := strings.HasPrefix(right.Role, "source_object_")
	if leftSource != rightSource {
		return leftSource
	}
	if leftSource && left.Path != right.Path {
		return left.Path < right.Path
	}
	if left.Role != right.Role {
		return left.Role < right.Role
	}
	return left.Path < right.Path
}

type privacyScanner struct {
	policy            privacyPolicy
	detector          genericPrivacyDetector
	entries           []privacyArtifact
	tainted           bool
	sourceObjectCount int
	sourceObjectBytes int64
}

type privacyTransaction struct {
	outputRoot string
	scanner    *privacyScanner
	files      map[string][]byte
	tainted    bool
	caps       privacyScanCaps
	totalBytes int64
}

func decodePrivacyPolicy(r io.Reader) (privacyPolicy, error) {
	data, err := io.ReadAll(io.LimitReader(r, privacyPolicyBytesMax+1))
	if err != nil {
		return privacyPolicy{}, err
	}
	if len(data) == 0 || len(data) > privacyPolicyBytesMax || !utf8.Valid(data) {
		return privacyPolicy{}, errors.New("invalid privacy policy bytes")
	}
	var policy privacyPolicy
	if err := decodeStrictBytes(data, &policy); err != nil {
		return privacyPolicy{}, err
	}
	if policy.Schema != privacyPolicySchema || len(policy.DenyTerms) == 0 || len(policy.DenyTerms) > privacyPolicyTermsMax {
		return privacyPolicy{}, errors.New("invalid privacy policy schema or term count")
	}
	totalBytes := 0
	previous := ""
	for _, term := range policy.DenyTerms {
		runeCount := utf8.RuneCountInString(term)
		totalBytes += len(term)
		if !utf8.ValidString(term) || runeCount < 4 || runeCount > 256 || len(term) > 1024 || strings.TrimSpace(term) != term || strings.ToLower(term) != term || term <= previous {
			return privacyPolicy{}, errors.New("invalid canonical privacy term")
		}
		for _, value := range term {
			if unicode.IsControl(value) {
				return privacyPolicy{}, errors.New("privacy term contains a control code point")
			}
		}
		previous = term
	}
	if totalBytes > privacyPolicyTermBytesMax {
		return privacyPolicy{}, errors.New("privacy terms exceed byte cap")
	}
	canonical, err := json.Marshal(struct {
		Schema    string   `json:"$schema"`
		DenyTerms []string `json:"deny_terms"`
	}{Schema: policy.Schema, DenyTerms: policy.DenyTerms})
	if err != nil || !bytes.Equal(canonical, data) {
		return privacyPolicy{}, errors.New("privacy policy is not canonical compact JSON")
	}
	policy.SHA256 = shaString(string(data))
	return policy, nil
}

func privacyDetectorDefinition() ([]privacyPattern, []string, []string) {
	patterns := []privacyPattern{
		{ID: "unix_machine_path", Pattern: "(?i)/(?:ho" + "me|users|root|tmp)/"},
		{ID: "windows_profile_path", Pattern: `(?i)[a-z]:\\` + `users\\`},
		{ID: "private_key_pem", Pattern: "-----BEGIN (?:RSA |EC |DSA |OPENSSH )?PRIVATE " + "KEY-----"},
		{ID: "github_classic", Pattern: "gh" + `[pousr]_[A-Za-z0-9_]{20,255}`},
		{ID: "github_fine_grained", Pattern: "github" + `_pat_[A-Za-z0-9_]{20,255}`},
		{ID: "openai_or_anthropic", Pattern: "s" + `k-(?:ant-)?[A-Za-z0-9_-]{20,255}`},
		{ID: "aws_access_key", Pattern: `(?:AK` + `IA|ASIA)[A-Z0-9]{16}`},
		{ID: "slack_token", Pattern: "xo" + `x[baprs]-[A-Za-z0-9-]{10,255}`},
		{ID: "bearer_header", Pattern: `(?i)\bauthorization[ \t]*:[ \t]*bearer[ \t]+[A-Za-z0-9._~+/=-]{12,512}`},
		{ID: "secret_assignment", Pattern: `(?im)^(?:[ \t]*(?:"(?:password|passwd|secret|token|api[_-]?key)"|'(?:password|passwd|secret|token|api[_-]?key)'|(?:password|passwd|secret|token|api[_-]?key))[ \t]*[:=][ \t]*(?:"[^"\r\n]{8,512}"|'[^'\r\n]{8,512}')[ \t]*[,;]?[ \t]*|(?:password|passwd|secret|token|api[_-]?key)=[A-Za-z0-9_+=~-]{8,512})$`},
	}
	positives := []string{
		"/ho" + "me/example/work", "/ro" + "ot/private", `C:\` + `Users\Example\repo`,
		"-----BEGIN PRIVATE " + "KEY-----", "gh" + "p_abcdefghijklmnopqrstuvwxyz123456",
		"github" + "_pat_abcdefghijklmnopqrstuvwxyz", "s" + "k-ant-abcdefghijklmnopqrstuvwxyz1234",
		"AK" + "IAABCDEFGHIJKLMNOP", "xo" + "xb-1234567890-abcdef",
		"Authorization: " + "Bearer abcdefghijklmnop", "api_" + "key=abcdefghijklmnop",
		`token = "abcdefghijklmnop";`, `password: "abcdefghijklmnop"`, `'secret': 'abcdefghijklmnop'`,
		`"api_key": "abcdefghijklmnop",`, `  "token": "abc-def-12345678"  `, `"secret": "密密密密密密密密"`,
	}
	negatives := []string{
		"/template/path", "/homework/example", `C:\UserGuide\repo`, "-----BEGIN PUBLIC KEY-----", "gh" + "p_example", "sk-short", "AKIAEXAMPLE", "Authorization: Basic abcdefghijklmnop",
		"token_count=12", "api_key_name=generic", "token = strings.TrimPrefix(token, prefix)", "token = normalizeProjectPath(token)",
		"secret = config.Value", "password = make([]byte, 32)", "secret=config/value", "token=config.Value", "prefix api_key=abcdefghijklmnop",
		`"secret': "abcdefghijklmnop"`, `'secret": 'abcdefghijklmnop'`, `"secret": "abc\"defghijklmnop"`,
	}
	return patterns, positives, negatives
}

func newGenericPrivacyDetector() (genericPrivacyDetector, error) {
	patterns, positives, negatives := privacyDetectorDefinition()
	detector := genericPrivacyDetector{Version: privacyDetectorVersion}
	for _, pattern := range patterns {
		compiled, err := regexp.Compile(pattern.Pattern)
		if err != nil {
			return genericPrivacyDetector{}, fmt.Errorf("compile privacy detector %s: %w", pattern.ID, err)
		}
		detector.patterns = append(detector.patterns, compiledPrivacyPattern{ID: pattern.ID, RE: compiled})
	}
	detector.DefinitionSHA256 = canonicalSHA256(map[string]any{"detector_version": detector.Version, "patterns": patterns, "positive_vectors": positives, "negative_vectors": negatives})
	return detector, nil
}

func (d genericPrivacyDetector) Match(data []byte) string {
	if !utf8.Valid(data) {
		return "invalid_utf8"
	}
	for _, pattern := range d.patterns {
		if pattern.RE.Find(data) != nil {
			return pattern.ID
		}
	}
	return ""
}

func newPrivacyScanner(policy privacyPolicy) (*privacyScanner, error) {
	if policy.Schema != privacyPolicySchema || !validSHA256(policy.SHA256) || len(policy.DenyTerms) == 0 {
		return nil, errors.New("invalid privacy policy binding")
	}
	detector, err := newGenericPrivacyDetector()
	if err != nil {
		return nil, err
	}
	return &privacyScanner{policy: policy, detector: detector}, nil
}

func (s *privacyScanner) Scan(role, path string, data []byte) error {
	if s == nil || s.tainted {
		return errors.New("privacy policy violation")
	}
	lower := strings.ToLower(string(data))
	for _, term := range s.policy.DenyTerms {
		if strings.Contains(lower, term) {
			s.tainted = true
			return errors.New("privacy policy violation")
		}
	}
	if s.detector.Match(data) != "" {
		s.tainted = true
		return errors.New("privacy policy violation")
	}
	s.entries = append(s.entries, privacyArtifact{Role: role, Path: path, SHA256: shaString(string(data)), Bytes: len(data)})
	return nil
}

func newPrivacyTransaction(outputRoot string, scanner *privacyScanner) *privacyTransaction {
	return &privacyTransaction{outputRoot: outputRoot, scanner: scanner, files: map[string][]byte{}, caps: protocolV7ScanCaps()}
}

func (t *privacyTransaction) Add(role, path string, data []byte) error {
	clean := filepath.ToSlash(filepath.Clean(path))
	if t == nil || t.scanner == nil || clean == "." || clean != path || strings.HasPrefix(clean, "../") || filepath.IsAbs(path) {
		if t != nil {
			t.tainted = true
		}
		return errors.New("invalid privacy transaction artifact")
	}
	if _, exists := t.files[clean]; exists {
		t.tainted = true
		return errors.New("duplicate privacy transaction artifact")
	}
	if len(t.files)+1 > t.caps.DurableArtifactCountMax || int64(len(data)) > t.caps.SingleDurableArtifactBytesMax || t.totalBytes+int64(len(data)) > t.caps.DurableArtifactBytesTotalMax {
		t.tainted = true
		return errors.New("protocol-v7 durable artifact cap exceeded")
	}
	if err := t.scanner.Scan(role, clean, data); err != nil {
		t.tainted = true
		return err
	}
	t.files[clean] = append([]byte(nil), data...)
	t.totalBytes += int64(len(data))
	return nil
}

func (t *privacyTransaction) Publish(stdout io.Writer, output []byte) error {
	if t == nil || t.scanner == nil || t.tainted || t.scanner.tainted {
		return errors.New("privacy policy violation")
	}
	if err := t.scanner.Scan("controller_output", "stdout", output); err != nil {
		t.tainted = true
		return err
	}
	parent := filepath.Dir(t.outputRoot)
	stage, err := os.MkdirTemp(parent, ".c3-v7-output-")
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = os.RemoveAll(stage)
		}
	}()
	paths := make([]string, 0, len(t.files))
	for path := range t.files {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := writeExclusive(filepath.Join(stage, filepath.FromSlash(path)), t.files[path]); err != nil {
			return err
		}
	}
	if err := syncDirectoryTree(stage); err != nil {
		return err
	}
	if _, err := os.Lstat(t.outputRoot); err == nil {
		return errors.New("controller output already exists")
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if err := os.Rename(stage, t.outputRoot); err != nil {
		return err
	}
	if err := syncDirectory(parent); err != nil {
		_ = os.RemoveAll(t.outputRoot)
		_ = syncDirectory(parent)
		return err
	}
	committed = true
	written, writeErr := stdout.Write(output)
	if writeErr != nil || written != len(output) {
		_ = os.RemoveAll(t.outputRoot)
		_ = syncDirectory(parent)
		if writeErr != nil {
			return writeErr
		}
		return io.ErrShortWrite
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	defer directory.Close()
	return directory.Sync()
}

func syncDirectoryTree(root string) error {
	var directories []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			directories = append(directories, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Slice(directories, func(i, j int) bool { return len(directories[i]) > len(directories[j]) })
	for _, directory := range directories {
		if err := syncDirectory(directory); err != nil {
			return err
		}
	}
	return nil
}

type buildPackage struct {
	Dir        string
	GoFiles    []string
	CgoFiles   []string
	EmbedFiles []string
}

func validateAdapterImplementationConfig(b benchmarkConfig) error {
	if b.Schema != benchmarkSchema {
		return errors.New("benchmark schema mismatch")
	}
	if b.K != 5 {
		return errors.New("benchmark k must be 5")
	}
	if b.SemanticMode != semanticDisabled && b.SemanticMode != semanticDefault {
		return errors.New("invalid semantic mode")
	}
	if b.Thresholds.OwnerRecallAt5Delta <= 0 || b.Thresholds.StructuralOwnerPrecision <= 0 || b.Thresholds.CanonicalRowBytesRatio <= 0 {
		return errors.New("invalid thresholds")
	}
	if b.Scale.Multiplier < 1 || len(b.Scale.Tokens) == 0 {
		return errors.New("invalid scale config")
	}
	return nil
}

func hashThresholdAuthority(a *thresholdAuthority) string {
	if a == nil {
		return ""
	}
	data, _ := json.Marshal(a)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func validateBaselineAdmission(b benchmarkConfig, p provenance) error {
	if err := validateAdapterImplementationConfig(b); err != nil {
		return err
	}
	if b.ContextThresholdAuthority == nil {
		return errors.New("canonical_row_bytes threshold is not externally re-ratified")
	}
	a := b.ContextThresholdAuthority
	if !strings.HasPrefix(a.CheckinRef, "checkins.jsonl#seq=") || !validSHA256(a.CheckinSHA256) || allZero(a.CheckinSHA256) || !validSHA256(a.RecordHash) || allZero(a.RecordHash) || !validSHA256(a.DefinitionSHA256) || allZero(a.DefinitionSHA256) {
		return errors.New("invalid external threshold authority")
	}
	if len(p.ContextThresholdAuthorityRecord) == 0 {
		return errors.New("external authority record is missing")
	}
	if err := validateThresholdAuthorityRecord(a, p.ContextThresholdAuthorityRecord); err != nil {
		return err
	}
	if p.ContextThresholdAuthoritySHA256 != hashThresholdAuthority(a) {
		return errors.New("threshold authority provenance mismatch")
	}
	return validateProvenance(p)
}

func validateThresholdAuthorityRecord(a *thresholdAuthority, line []byte) error {
	sum := sha256.Sum256(line)
	if hex.EncodeToString(sum[:]) != a.CheckinSHA256 {
		return errors.New("external authority line hash mismatch")
	}
	var envelope struct {
		Payload struct {
			CheckinID     string `json:"checkin_id"`
			DecisionDelta string `json:"decision_delta"`
			Type          string `json:"type"`
		} `json:"payload"`
		RecordHash string `json:"record_hash"`
		Seq        int    `json:"seq"`
	}
	if err := json.Unmarshal(bytes.TrimSpace(line), &envelope); err != nil {
		return fmt.Errorf("external authority record: %w", err)
	}
	wantSeq, err := strconv.Atoi(strings.TrimPrefix(a.CheckinRef, "checkins.jsonl#seq="))
	if err != nil || envelope.Seq != wantSeq {
		return errors.New("external authority sequence mismatch")
	}
	if envelope.RecordHash != a.RecordHash {
		return errors.New("external authority record hash mismatch")
	}
	decision := strings.ToLower(envelope.Payload.DecisionDelta)
	if envelope.Payload.Type != "steering_checkin" || envelope.Payload.CheckinID == "" || !strings.Contains(decision, "re-ratify") || !strings.Contains(decision, "canonical") || !strings.Contains(decision, "1.05") {
		return errors.New("external authority does not ratify canonical row bytes 1.05")
	}
	return nil
}
func validateProvenance(p provenance) error {
	if p.ExperimentID == "" || p.ArmID == "" || !validHexN(p.Commit, 40) || !validHexN(p.Tree, 40) {
		return errors.New("invalid provenance identity")
	}
	values := []string{p.SourceCapsuleSHA256, p.DiffSHA256, p.FixtureSHA256, p.ScorerSHA256, p.ControllerSHA256, p.RuntimeSHA256, p.LogicalDumpSHA256, p.EnvironmentSHA256, p.ModuleGraphSHA256, p.BudgetSHA256, p.ActionEnvelopeSHA256, p.ProjectDirSHA256, p.C3DirSHA256, p.ContextThresholdAuthoritySHA256}
	for _, v := range values {
		if !validSHA256(v) || allZero(v) {
			return errors.New("invalid provenance hash")
		}
	}
	if p.ControllerCommit != "" || p.ControllerTree != "" || p.ControllerSourceCapsuleSHA256 != "" || p.CandidateDeltaSHA256 != "" || p.BundleSHA256 != "" || p.BenchmarkSHA256 != "" {
		if !validHexN(p.ControllerCommit, 40) || !validHexN(p.ControllerTree, 40) {
			return errors.New("invalid controller provenance identity")
		}
		for _, value := range []string{p.ControllerSourceCapsuleSHA256, p.BundleSHA256, p.BenchmarkSHA256} {
			if !validSHA256(value) || allZero(value) {
				return errors.New("invalid dual-source provenance hash")
			}
		}
		if p.ControllerCommit != p.Commit || p.ControllerTree != p.Tree {
			if !validSHA256(p.CandidateDeltaSHA256) || allZero(p.CandidateDeltaSHA256) {
				return errors.New("candidate provenance is missing its delta hash")
			}
		} else if p.CandidateDeltaSHA256 != "" {
			return errors.New("baseline provenance unexpectedly has a candidate delta")
		}
	}
	if p.CorpusMode != corpusIsolated && p.CorpusMode != corpusCombined && p.CorpusMode != "scale" {
		return errors.New("invalid provenance corpus mode")
	}
	if p.SemanticMode != semanticDisabled && p.SemanticMode != semanticDefault {
		return errors.New("invalid provenance semantic mode")
	}
	return validateProtocolV7Provenance(p)
}

func validateProtocolV7Provenance(p provenance) error {
	v4 := p.PrivacyPolicySHA256 != "" || p.PrivacyTermCount != 0 || p.PrivacyDetectorSHA256 != "" || p.GoExecutableSHA256 != "" || p.GoModVerifySHA256 != "" || p.ScanCapsSHA256 != "" || p.SourceBundleHeadsSHA256 != "" || p.ProtocolTestSHA256 != ""
	if v4 {
		if p.PrivacyTermCount <= 0 {
			return errors.New("invalid protocol-v7 privacy term count")
		}
		for _, hash := range []string{p.PrivacyPolicySHA256, p.PrivacyDetectorSHA256, p.GoExecutableSHA256, p.GoModVerifySHA256, p.ScanCapsSHA256, p.SourceBundleHeadsSHA256, p.ProtocolTestSHA256} {
			if !validSHA256(hash) || allZero(hash) {
				return errors.New("invalid protocol-v7 provenance hash")
			}
		}
		parentHashes := []string{p.ParentBaselineAuthoritySHA256, p.ParentBaselineOutputSHA256, p.ParentBaselineOrderedRunSHA256, p.ParentBaselineHistorySHA256, p.ParentBaselineHistoryTailSHA256, p.ParentBaselinePrivacySHA256, p.ParentBaselineValidatorHash, p.ParentBaselineValidatorPayload}
		hasParent := p.ParentBaselineRunCount != 0
		for _, hash := range parentHashes {
			hasParent = hasParent || hash != ""
		}
		if hasParent {
			if p.ParentBaselineRunCount <= 0 || p.Commit == p.ControllerCommit {
				return errors.New("invalid protocol-v7 parent provenance")
			}
			for _, hash := range parentHashes {
				if !validSHA256(hash) || allZero(hash) {
					return errors.New("incomplete protocol-v7 parent provenance")
				}
			}
		}
	}
	return nil
}
func validHexN(v string, n int) bool {
	if len(v) != n {
		return false
	}
	_, err := hex.DecodeString(v)
	return err == nil
}
func validSHA256(v string) bool { return validHexN(v, 64) }
func allZero(v string) bool     { return strings.Trim(v, "0") == "" }

type resourceBudget struct {
	WallTimeMillis   int64 `json:"wall_time_millis"`
	CPUTimeMillis    int64 `json:"cpu_time_millis"`
	MaxRSSBytes      int64 `json:"max_rss_bytes"`
	ProcessCount     int   `json:"process_count"`
	SQLiteRowCount   int   `json:"sqlite_row_count"`
	LogicalDumpBytes int   `json:"logical_dump_bytes"`
	StdoutBytes      int   `json:"stdout_bytes"`
	StderrBytes      int   `json:"stderr_bytes"`
	CaseCount        int   `json:"case_count"`
}
type historyRecord struct {
	Schema          string         `json:"$schema"`
	ExperimentID    string         `json:"experiment_id"`
	ArmID           string         `json:"arm_id"`
	ParentKeep      string         `json:"parent_keep"`
	ChangedVariable string         `json:"changed_variable"`
	ChangedPaths    []string       `json:"changed_paths"`
	ResultPath      string         `json:"result_path"`
	ResultSHA256    string         `json:"result_sha256"`
	Provenance      provenance     `json:"provenance"`
	Budgets         resourceBudget `json:"budgets"`
	Status          string         `json:"status"`
	Reason          string         `json:"reason"`
	ErrorClass      string         `json:"error_class,omitempty"`
	ErrorSHA256     string         `json:"error_sha256,omitempty"`
	Evidence        []string       `json:"evidence"`
	PrevHash        string         `json:"prev_hash"`
	RecordHash      string         `json:"record_hash"`
}

func hashHistoryRecord(r historyRecord) string {
	r.RecordHash = ""
	data, _ := json.Marshal(r)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
func budgetLimitsComplete(b resourceBudget) bool {
	return b.WallTimeMillis > 0 && b.CPUTimeMillis > 0 && b.MaxRSSBytes > 0 && b.ProcessCount > 0 && b.SQLiteRowCount > 0 && b.LogicalDumpBytes > 0 && b.StdoutBytes > 0 && b.StderrBytes > 0 && b.CaseCount > 0
}

func validateControllerBudgetLimits(b resourceBudget) error {
	if !budgetLimitsComplete(b) {
		return errors.New("authority has incomplete budget limits")
	}
	if b.CPUTimeMillis != 10_000 || b.MaxRSSBytes != 536_870_912 || b.ProcessCount != 16 || b.SQLiteRowCount != 1_000_000 || b.LogicalDumpBytes != 64<<20 || b.StdoutBytes != 16<<20 || b.StderrBytes != 1<<20 || b.CaseCount != 100 {
		return errors.New("authority changes a frozen non-wall budget")
	}
	if b.WallTimeMillis%1000 != 0 {
		return errors.New("authority wall must be an exact number of seconds")
	}
	wall := time.Duration(b.WallTimeMillis) * time.Millisecond
	cpu := time.Duration(b.CPUTimeMillis) * time.Millisecond
	if wall <= cpu+confinementStartupMargin {
		return errors.New("authority wall must exceed CPU plus startup margin")
	}
	if b.WallTimeMillis != registeredWallTimeMillis {
		return errors.New("authority wall does not match the registered wall envelope")
	}
	return nil
}

func runtimeMaxSeconds(b resourceBudget) (int64, error) {
	if err := validateControllerBudgetLimits(b); err != nil {
		return 0, err
	}
	return b.WallTimeMillis / 1000, nil
}

func controllerTimeoutFor(armCount int, limits resourceBudget) (time.Duration, error) {
	if armCount <= 0 {
		return 0, errors.New("controller arm count must be positive")
	}
	if _, err := runtimeMaxSeconds(limits); err != nil {
		return 0, err
	}
	wall := time.Duration(limits.WallTimeMillis) * time.Millisecond
	timeout := time.Duration(armCount)*wall + controllerEnvelopeOverhead
	if err := validateControllerTimeout(timeout, armCount, limits); err != nil {
		return 0, err
	}
	return timeout, nil
}

func validateControllerTimeout(timeout time.Duration, armCount int, limits resourceBudget) error {
	if armCount <= 0 {
		return errors.New("controller arm count must be positive")
	}
	if _, err := runtimeMaxSeconds(limits); err != nil {
		return err
	}
	aggregateArmWall := time.Duration(armCount) * time.Duration(limits.WallTimeMillis) * time.Millisecond
	minimum := aggregateArmWall + controllerEnvelopeOverhead
	if timeout < minimum || timeout <= aggregateArmWall {
		return errors.New("controller timeout must cover aggregate arm wall plus overhead")
	}
	return nil
}

func actualBudgetComplete(b resourceBudget) bool {
	return b.WallTimeMillis > 0 && b.CPUTimeMillis >= 0 && b.MaxRSSBytes > 0 && b.ProcessCount > 0 && b.SQLiteRowCount > 0 && b.LogicalDumpBytes > 0 && b.StdoutBytes > 0 && b.StderrBytes >= 0 && b.CaseCount > 0
}
func partialBudgetValid(b resourceBudget) bool {
	return b.WallTimeMillis >= 0 && b.CPUTimeMillis >= 0 && b.MaxRSSBytes >= 0 && b.ProcessCount >= 0 && b.SQLiteRowCount >= 0 && b.LogicalDumpBytes >= 0 && b.StdoutBytes >= 0 && b.StderrBytes >= 0 && b.CaseCount >= 0
}
func validateFailureProvenance(p provenance) error {
	if p.ExperimentID == "" || p.ArmID == "" || !validHexN(p.Commit, 40) || !validHexN(p.Tree, 40) {
		return errors.New("invalid failure provenance identity")
	}
	required := []string{p.SourceCapsuleSHA256, p.DiffSHA256, p.FixtureSHA256, p.ScorerSHA256, p.ControllerSHA256, p.RuntimeSHA256, p.EnvironmentSHA256, p.ModuleGraphSHA256, p.BudgetSHA256, p.ActionEnvelopeSHA256, p.ContextThresholdAuthoritySHA256}
	for _, value := range required {
		if !validSHA256(value) || allZero(value) {
			return errors.New("invalid failure provenance hash")
		}
	}
	if p.ControllerCommit != "" || p.ControllerTree != "" || p.ControllerSourceCapsuleSHA256 != "" || p.CandidateDeltaSHA256 != "" || p.BundleSHA256 != "" || p.BenchmarkSHA256 != "" {
		if !validHexN(p.ControllerCommit, 40) || !validHexN(p.ControllerTree, 40) {
			return errors.New("invalid failure controller identity")
		}
		for _, value := range []string{p.ControllerSourceCapsuleSHA256, p.BundleSHA256, p.BenchmarkSHA256} {
			if !validSHA256(value) || allZero(value) {
				return errors.New("invalid failure dual-source hash")
			}
		}
		if p.ControllerCommit != p.Commit || p.ControllerTree != p.Tree {
			if !validSHA256(p.CandidateDeltaSHA256) || allZero(p.CandidateDeltaSHA256) {
				return errors.New("candidate failure provenance is missing its delta hash")
			}
		} else if p.CandidateDeltaSHA256 != "" {
			return errors.New("baseline failure provenance unexpectedly has a candidate delta")
		}
	}
	for _, value := range []string{p.LogicalDumpSHA256, p.ProjectDirSHA256, p.C3DirSHA256} {
		if value != "" && (!validSHA256(value) || allZero(value)) {
			return errors.New("invalid optional failure provenance hash")
		}
	}
	if p.CorpusMode != corpusIsolated && p.CorpusMode != corpusCombined && p.CorpusMode != "scale" {
		return errors.New("invalid failure corpus mode")
	}
	if p.SemanticMode != semanticDisabled && p.SemanticMode != semanticDefault {
		return errors.New("invalid failure semantic mode")
	}
	return validateProtocolV7Provenance(p)
}
func verifyHistory(records []historyRecord) error {
	return verifyHistorySchema(records, historySchema)
}

func verifyHistorySchema(records []historyRecord, schema string) error {
	if schema != historySchema && schema != historyV4Schema {
		return errors.New("unsupported history schema")
	}
	prev := "GENESIS"
	for i, r := range records {
		if r.Schema != schema || r.PrevHash != prev || !validSHA256(r.ResultSHA256) || r.RecordHash != hashHistoryRecord(r) {
			return fmt.Errorf("invalid history record %d", i)
		}
		if r.ErrorClass == "" {
			if !actualBudgetComplete(r.Budgets) {
				return fmt.Errorf("incomplete successful-run budget in history record %d", i)
			}
			if err := validateProvenance(r.Provenance); err != nil {
				return err
			}
		} else {
			if !partialBudgetValid(r.Budgets) || !validSHA256(r.ErrorSHA256) {
				return fmt.Errorf("invalid failure actuals in history record %d", i)
			}
			if err := validateFailureProvenance(r.Provenance); err != nil {
				return err
			}
		}
		if r.Status != "keep" && r.Status != "discard" && r.Status != "crash" && r.Status != "invalid" {
			return errors.New("invalid history status")
		}
		prev = r.RecordHash
	}
	return nil
}
func writeExclusive(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(data); err != nil {
		return err
	}
	return f.Sync()
}
func appendHistory(path string, r historyRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	existing, err := readHistory(path)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	if len(existing) == 0 {
		r.PrevHash = "GENESIS"
	} else {
		r.PrevHash = existing[len(existing)-1].RecordHash
	}
	r.RecordHash = hashHistoryRecord(r)
	if err := verifyHistory(append(existing, r)); err != nil {
		return err
	}
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err = f.Write(append(data, '\n')); err != nil {
		return err
	}
	return f.Sync()
}
func readHistory(path string) ([]historyRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return decodeHistoryBytes(data)
}

func decodeHistoryBytes(data []byte) ([]historyRecord, error) {
	var out []historyRecord
	s := bufio.NewScanner(bytes.NewReader(data))
	s.Buffer(make([]byte, 1024), 4<<20)
	for s.Scan() {
		var r historyRecord
		if err := decodeStrictBytes(s.Bytes(), &r); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, s.Err()
}
func classifyBudget(limit, actual resourceBudget) string {
	if actual.WallTimeMillis > limit.WallTimeMillis || actual.CPUTimeMillis > limit.CPUTimeMillis || actual.MaxRSSBytes > limit.MaxRSSBytes || actual.ProcessCount > limit.ProcessCount || actual.SQLiteRowCount > limit.SQLiteRowCount || actual.LogicalDumpBytes > limit.LogicalDumpBytes || actual.StdoutBytes > limit.StdoutBytes || actual.StderrBytes > limit.StderrBytes || actual.CaseCount > limit.CaseCount {
		return "crash"
	}
	return "score"
}

type sourceInput struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Origin string `json:"origin"`
}
type sourceCapsule struct {
	Schema                    string        `json:"$schema"`
	HeadCommit                string        `json:"head_commit"`
	HeadTree                  string        `json:"head_tree"`
	RepositoryBuildInputCount int           `json:"repository_build_input_count"`
	Inputs                    []sourceInput `json:"inputs"`
	DirtyPatchSHA256          string        `json:"dirty_patch_sha256"`
}
type buildReplayManifest struct {
	ControllerSHA256                 string `json:"controller_sha256"`
	RuntimeSHA256                    string `json:"runtime_sha256"`
	RebuiltControllerSHA256          string `json:"rebuilt_controller_sha256"`
	RebuiltRuntimeSHA256             string `json:"rebuilt_runtime_sha256"`
	SourceCapsuleRebuildVerified     bool   `json:"source_capsule_rebuild_verified"`
	ControllerCapsuleRebuildVerified bool   `json:"controller_capsule_rebuild_verified,omitempty"`
	RuntimeCapsuleRebuildVerified    bool   `json:"runtime_capsule_rebuild_verified,omitempty"`
	BundleVerified                   bool   `json:"bundle_verified"`
}

const semanticSourceSHA256 = "7b79f5d218fb422654174c5d651a55b0ac2b3fb1f38f9bb048f03492afc34883"
const authoritativeDirtyPatchSHA256 = "def9ef26b435525e0ba8b9dcb704bc55ce1c51eb93f51e12ee372d05a72036af"

func validateSourceCapsule(c sourceCapsule) error {
	if c.Schema != sourceCapsuleSchema || !validHexN(c.HeadCommit, 40) || !validHexN(c.HeadTree, 40) || c.RepositoryBuildInputCount != len(c.Inputs) || c.RepositoryBuildInputCount == 0 {
		return errors.New("invalid source capsule")
	}
	if c.DirtyPatchSHA256 != authoritativeDirtyPatchSHA256 {
		return errors.New("dirty patch mismatch")
	}
	seen := map[string]bool{}
	for _, in := range c.Inputs {
		clean := filepath.ToSlash(filepath.Clean(in.Path))
		if strings.HasPrefix(clean, "../") || !strings.HasPrefix(clean, "cli/") || strings.Contains(clean, ".okra") || strings.Contains(clean, "private") || seen[clean] || !validSHA256(in.SHA256) {
			return fmt.Errorf("invalid build input %q", in.Path)
		}
		if clean == "cli/internal/store/semantic.go" && in.SHA256 != semanticSourceSHA256 {
			return errors.New("semantic.go hash mismatch")
		}
		if in.Origin != "head" && in.Origin != "working-tree" {
			return errors.New("invalid source origin")
		}
		seen[clean] = true
	}
	return nil
}
func validateBuildReplay(m buildReplayManifest) error {
	dualSource := m.ControllerCapsuleRebuildVerified || m.RuntimeCapsuleRebuildVerified
	if dualSource {
		if !m.ControllerCapsuleRebuildVerified || !m.RuntimeCapsuleRebuildVerified || !m.BundleVerified {
			return errors.New("dual-source capsule or bundle rebuild not verified")
		}
	} else if !m.SourceCapsuleRebuildVerified {
		return errors.New("source capsule rebuild not verified")
	}
	for _, h := range []string{m.ControllerSHA256, m.RuntimeSHA256, m.RebuiltControllerSHA256, m.RebuiltRuntimeSHA256} {
		if !validSHA256(h) || allZero(h) {
			return errors.New("invalid build replay hash")
		}
	}
	if m.ControllerSHA256 != m.RebuiltControllerSHA256 || m.RuntimeSHA256 != m.RebuiltRuntimeSHA256 {
		return errors.New("bundle rebuild mismatch")
	}
	return nil
}

func loadControllerAuthority(path string) (controllerAuthority, error) {
	f, err := os.Open(path)
	if err != nil {
		return controllerAuthority{}, err
	}
	defer f.Close()
	var a controllerAuthority
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&a); err != nil {
		return a, err
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return a, errors.New("trailing authority JSON")
		}
		return a, err
	}
	if a.Schema != controllerAuthoritySchema {
		return a, errors.New("controller authority schema mismatch")
	}
	return a, nil
}

func controllerAuthoritySchemaAt(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return controllerAuthoritySchemaBytes(data)
}

func controllerAuthoritySchemaBytes(data []byte) (string, error) {
	if err := rejectDuplicateJSONKeys(data); err != nil {
		return "", err
	}
	var envelope struct {
		Schema string `json:"$schema"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		return "", err
	}
	if envelope.Schema != controllerAuthoritySchema && envelope.Schema != controllerAuthorityV3Schema && envelope.Schema != controllerAuthorityV4Schema {
		return "", errors.New("unsupported controller authority schema")
	}
	return envelope.Schema, nil
}

func loadControllerAuthorityV3(path string) (controllerAuthorityV3, error) {
	file, err := os.Open(path)
	if err != nil {
		return controllerAuthorityV3{}, err
	}
	defer file.Close()
	return decodeControllerAuthorityV3(file)
}

func decodeControllerAuthorityV4(r io.Reader) (controllerAuthorityV4, error) {
	var authority controllerAuthorityV4
	data, err := io.ReadAll(io.LimitReader(r, (4<<20)+1))
	if err != nil || len(data) > 4<<20 {
		return authority, errors.New("controller authority v4 exceeds byte cap")
	}
	if err := decodeStrictBytes(data, &authority); err != nil {
		return authority, err
	}
	if authority.Schema != controllerAuthorityV4Schema {
		return authority, errors.New("controller authority v4 schema mismatch")
	}
	return authority, nil
}

func loadControllerAuthorityV4(path string) (controllerAuthorityV4, error) {
	file, err := os.Open(path)
	if err != nil {
		return controllerAuthorityV4{}, err
	}
	defer file.Close()
	return decodeControllerAuthorityV4(file)
}

func (a controllerAuthorityV3) persistenceAuthority() controllerAuthority {
	return controllerAuthority{
		Schema:                          controllerAuthoritySchema,
		Expected:                        a.Expected,
		SourceCapsule:                   a.RuntimeSourceCapsule,
		BuildReplay:                     a.BuildReplay,
		BudgetLimits:                    a.BudgetLimits,
		ActionEnvelope:                  a.ActionEnvelope,
		CanonicalRowBytesDefinition:     a.CanonicalRowBytesDefinition,
		ContextThresholdAuthorityRecord: a.ContextThresholdAuthorityRecord,
	}
}
func fileSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
func canonicalSHA256(value any) string {
	data, _ := json.Marshal(value)
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func canonicalCandidateDeltaSHA256(delta candidateDelta) string {
	payload := map[string]any{
		"variable":            delta.Variable,
		"baseline_commit":     delta.BaselineCommit,
		"baseline_tree":       delta.BaselineTree,
		"candidate_commit":    delta.CandidateCommit,
		"candidate_tree":      delta.CandidateTree,
		"diff_sha256":         delta.DiffSHA256,
		"name_status_sha256":  delta.NameStatusSHA256,
		"name_status":         delta.NameStatus,
		"allowed_paths":       delta.AllowedPaths,
		"before_blob_sha256":  delta.BeforeBlobSHA256,
		"after_blob_sha256":   delta.AfterBlobSHA256,
		"bundle_sha256":       delta.BundleSHA256,
		"bundle_heads_sha256": delta.BundleHeadsSHA256,
	}
	return canonicalSHA256(payload)
}

type workerProgressRecord struct {
	Seq           int            `json:"seq"`
	RecordedAt    string         `json:"recorded_at"`
	PrevHash      string         `json:"prev_hash"`
	PayloadSHA256 string         `json:"payload_sha256"`
	Payload       map[string]any `json:"payload"`
	RecordHash    string         `json:"record_hash"`
}

func verifyWorkerRecordPrefix(r io.Reader, targetSeq int, wantRecordHash, wantPayloadHash string) (map[string]any, error) {
	if targetSeq <= 0 || !validSHA256(wantRecordHash) || !validSHA256(wantPayloadHash) {
		return nil, errors.New("invalid governed validator record authority")
	}
	scanner := bufio.NewScanner(io.LimitReader(r, 4<<20))
	scanner.Buffer(make([]byte, 4096), 1<<20)
	previous := "GENESIS"
	expectedSeq := 1
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			return nil, errors.New("blank governed validator record")
		}
		var record workerProgressRecord
		if err := decodeStrictBytes(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		when, err := time.Parse(time.RFC3339, record.RecordedAt)
		if err != nil || when.Location() != time.UTC || record.Seq != expectedSeq || record.PrevHash != previous {
			return nil, errors.New("invalid governed validator record chain")
		}
		if canonicalSHA256(record.Payload) != record.PayloadSHA256 {
			return nil, errors.New("governed validator payload hash mismatch")
		}
		withoutHash := map[string]any{
			"seq": record.Seq, "recorded_at": record.RecordedAt, "prev_hash": record.PrevHash,
			"payload_sha256": record.PayloadSHA256, "payload": record.Payload,
		}
		if canonicalSHA256(withoutHash) != record.RecordHash {
			return nil, errors.New("governed validator record hash mismatch")
		}
		if record.Seq == targetSeq {
			if record.RecordHash != wantRecordHash || record.PayloadSHA256 != wantPayloadHash {
				return nil, errors.New("governed validator trust root mismatch")
			}
			return record.Payload, nil
		}
		previous = record.RecordHash
		expectedSeq++
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return nil, errors.New("governed validator record is missing")
}

func hermeticGoEnvironment(home, moduleCache, buildCache string) []string {
	return []string{
		"PATH=" + os.Getenv("PATH"),
		"HOME=" + home,
		"LC_ALL=C",
		"LANG=C",
		"TZ=UTC",
		"GOENV=off",
		"GOWORK=off",
		"GOFLAGS=",
		"GOTOOLCHAIN=local",
		"CGO_ENABLED=0",
		"GOPROXY=off",
		"GOSUMDB=off",
		"GONOSUMDB=",
		"GOPRIVATE=",
		"GOMODCACHE=" + moduleCache,
		"GOCACHE=" + buildCache,
	}
}

type protocolGoEnvironment struct {
	GOOS        *string `json:"GOOS"`
	GOARCH      *string `json:"GOARCH"`
	GOVERSION   *string `json:"GOVERSION"`
	GOROOT      *string `json:"GOROOT"`
	CGOEnabled  *string `json:"CGO_ENABLED"`
	GOENV       *string `json:"GOENV"`
	GOWORK      *string `json:"GOWORK"`
	GOFLAGS     *string `json:"GOFLAGS"`
	GOTOOLCHAIN *string `json:"GOTOOLCHAIN"`
	GOMODCACHE  *string `json:"GOMODCACHE"`
}

func protocolEnvironmentIdentitySHA256(goEnvironmentJSON []byte, goExecutableHash string, allowedEnvironment []string) (string, error) {
	if !validSHA256(goExecutableHash) || allZero(goExecutableHash) || goExecutableHash != strings.ToLower(goExecutableHash) {
		return "", errors.New("invalid Go executable hash")
	}

	var goEnvironment protocolGoEnvironment
	if err := decodeStrictBytes(goEnvironmentJSON, &goEnvironment); err != nil {
		return "", fmt.Errorf("invalid Go environment: %w", err)
	}
	goValues := map[string]*string{
		"GOOS": goEnvironment.GOOS, "GOARCH": goEnvironment.GOARCH, "GOVERSION": goEnvironment.GOVERSION,
		"GOROOT": goEnvironment.GOROOT, "CGO_ENABLED": goEnvironment.CGOEnabled, "GOENV": goEnvironment.GOENV,
		"GOWORK": goEnvironment.GOWORK, "GOFLAGS": goEnvironment.GOFLAGS, "GOTOOLCHAIN": goEnvironment.GOTOOLCHAIN,
		"GOMODCACHE": goEnvironment.GOMODCACHE,
	}
	for key, value := range goValues {
		if value == nil {
			return "", fmt.Errorf("Go environment is missing %s", key)
		}
	}
	for _, key := range []string{"GOOS", "GOARCH", "GOVERSION", "GOROOT", "CGO_ENABLED", "GOWORK", "GOTOOLCHAIN", "GOMODCACHE"} {
		if *goValues[key] == "" {
			return "", fmt.Errorf("Go environment has empty %s", key)
		}
	}

	expectedAllowedKeys := []string{
		"PATH", "HOME", "LC_ALL", "LANG", "TZ", "GOENV", "GOWORK", "GOFLAGS", "GOTOOLCHAIN", "CGO_ENABLED",
		"GOPROXY", "GOSUMDB", "GONOSUMDB", "GOPRIVATE", "GOMODCACHE", "GOCACHE",
	}
	expected := make(map[string]bool, len(expectedAllowedKeys))
	for _, key := range expectedAllowedKeys {
		expected[key] = true
	}
	allowedValues := make(map[string]string, len(allowedEnvironment))
	for _, assignment := range allowedEnvironment {
		key, value, ok := strings.Cut(assignment, "=")
		if !ok || key == "" || !expected[key] {
			return "", fmt.Errorf("invalid allowed environment assignment %q", assignment)
		}
		if _, duplicate := allowedValues[key]; duplicate {
			return "", fmt.Errorf("duplicate allowed environment key %s", key)
		}
		allowedValues[key] = value
	}
	for _, key := range expectedAllowedKeys {
		if _, ok := allowedValues[key]; !ok {
			return "", fmt.Errorf("allowed environment is missing %s", key)
		}
	}
	for _, key := range []string{"PATH", "HOME", "LC_ALL", "LANG", "TZ", "GOENV", "GOWORK", "GOTOOLCHAIN", "CGO_ENABLED", "GOPROXY", "GOSUMDB", "GOMODCACHE", "GOCACHE"} {
		if allowedValues[key] == "" {
			return "", fmt.Errorf("allowed environment has empty %s", key)
		}
	}

	for _, key := range []string{"CGO_ENABLED", "GOWORK", "GOFLAGS", "GOTOOLCHAIN", "GOMODCACHE"} {
		if allowedValues[key] != *goValues[key] {
			return "", fmt.Errorf("Go environment and allowed environment disagree on %s", key)
		}
	}
	if allowedValues["GOENV"] != *goEnvironment.GOENV && !(allowedValues["GOENV"] == "off" && *goEnvironment.GOENV == "") {
		return "", errors.New("Go environment and allowed environment disagree on GOENV")
	}

	canonicalGoEnvironment := map[string]string{
		"GOOS": *goEnvironment.GOOS, "GOARCH": *goEnvironment.GOARCH, "GOVERSION": *goEnvironment.GOVERSION,
		"GOROOT": "<go-root-role>", "CGO_ENABLED": *goEnvironment.CGOEnabled, "GOENV": *goEnvironment.GOENV,
		"GOWORK": *goEnvironment.GOWORK, "GOFLAGS": *goEnvironment.GOFLAGS, "GOTOOLCHAIN": *goEnvironment.GOTOOLCHAIN,
		"GOMODCACHE": "<module-cache-role>",
	}
	canonicalAllowedEnvironment := make(map[string]string, len(allowedValues))
	for key, value := range allowedValues {
		canonicalAllowedEnvironment[key] = value
	}
	canonicalAllowedEnvironment["PATH"] = "<resolved-go-executable-path-role>"
	canonicalAllowedEnvironment["HOME"] = "<private-home-role>"
	canonicalAllowedEnvironment["GOMODCACHE"] = "<module-cache-role>"
	canonicalAllowedEnvironment["GOCACHE"] = "<build-cache-role>"

	return canonicalSHA256(map[string]any{
		"$schema":              "structural-retrieval-environment-identity.v1",
		"go_environment":       canonicalGoEnvironment,
		"go_executable_sha256": goExecutableHash,
		"allowed_environment":  canonicalAllowedEnvironment,
	}), nil
}

func decodeControllerAuthorityV3(r io.Reader) (controllerAuthorityV3, error) {
	var authority controllerAuthorityV3
	data, err := io.ReadAll(io.LimitReader(r, (4<<20)+1))
	if err != nil || len(data) > 4<<20 {
		return authority, errors.New("controller authority v3 exceeds byte cap")
	}
	if err := decodeStrictBytes(data, &authority); err != nil {
		return authority, err
	}
	if authority.Schema != controllerAuthorityV3Schema {
		return authority, errors.New("controller authority v3 schema mismatch")
	}
	return authority, nil
}

func scoringRegionSHA256(path string) (string, error) {
	source, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, path, source, parser.AllErrors)
	if err != nil {
		return "", err
	}
	var first, last *ast.FuncDecl
	var names []string
	inside := false
	for _, declaration := range parsed.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok || function.Recv != nil {
			continue
		}
		if function.Name.Name == "scoreCase" {
			first = function
			inside = true
		}
		if inside {
			names = append(names, function.Name.Name)
		}
		if function.Name.Name == "evaluateGate" {
			last = function
			break
		}
	}
	wantNames := []string{"scoreCase", "routeFieldCoverage", "summarizeMetrics", "recallAt", "reciprocalRank", "stringSet", "canonicalRowBytes", "verifyReport", "verifyReportWithProbes", "evaluateGate"}
	if first == nil || last == nil || !reflect.DeepEqual(names, wantNames) {
		return "", fmt.Errorf("scoring function boundary mismatch: %v", names)
	}
	tokenFile := fset.File(parsed.Pos())
	if tokenFile == nil {
		return "", errors.New("scoring source token file missing")
	}
	start := tokenFile.Offset(first.Pos())
	end := tokenFile.Offset(last.End())
	if start < 0 || end <= start || end > len(source) {
		return "", errors.New("invalid scoring source offsets")
	}
	sum := sha256.Sum256(source[start:end])
	return hex.EncodeToString(sum[:]), nil
}

func validateCleanSourceCapsule(c sourceCapsule) error {
	if c.Schema != sourceCapsuleSchema || !validHexN(c.HeadCommit, 40) || !validHexN(c.HeadTree, 40) || c.RepositoryBuildInputCount != len(c.Inputs) || len(c.Inputs) == 0 {
		return errors.New("invalid clean source capsule")
	}
	if c.DirtyPatchSHA256 != shaString("") {
		return errors.New("source capsule is not clean")
	}
	previous := ""
	for _, input := range c.Inputs {
		clean := filepath.ToSlash(filepath.Clean(input.Path))
		if clean != input.Path || !strings.HasPrefix(clean, "cli/") || strings.Contains(clean, ".okra") || strings.Contains(clean, "private") || clean <= previous || !validSHA256(input.SHA256) || input.Origin != "head" {
			return fmt.Errorf("invalid clean build input %q", input.Path)
		}
		previous = clean
		if clean == "cli/internal/store/semantic.go" && input.SHA256 != semanticSourceSHA256 {
			return errors.New("semantic.go hash mismatch")
		}
	}
	return nil
}

func validateDualSourceCapsules(baseline, candidate sourceCapsule, allowedPaths []string) error {
	if err := validateCleanSourceCapsule(baseline); err != nil {
		return fmt.Errorf("baseline capsule: %w", err)
	}
	if err := validateCleanSourceCapsule(candidate); err != nil {
		return fmt.Errorf("candidate capsule: %w", err)
	}
	allowed := stringSet(allowedPaths)
	base := make(map[string]sourceInput, len(baseline.Inputs))
	for _, input := range baseline.Inputs {
		base[input.Path] = input
	}
	if len(candidate.Inputs) != len(base) {
		return errors.New("candidate build-input closure changed")
	}
	changed := map[string]bool{}
	for _, input := range candidate.Inputs {
		before, ok := base[input.Path]
		if !ok {
			return fmt.Errorf("candidate added build input %q", input.Path)
		}
		if input.SHA256 != before.SHA256 {
			if !allowed[input.Path] {
				return fmt.Errorf("unregistered build input changed: %s", input.Path)
			}
			changed[input.Path] = true
		}
	}
	for path := range allowed {
		if path == "cli/cmd/search_test.go" {
			continue
		}
		if !changed[path] {
			return fmt.Errorf("registered runtime path did not change: %s", path)
		}
	}
	return nil
}

func validateCandidateDeltaAuthorityBinding(a controllerAuthorityV3) error {
	if a.Mode != "candidate" || a.CandidateDelta == nil {
		return errors.New("candidate delta authority binding requires candidate mode and delta")
	}
	delta := a.CandidateDelta
	if delta.BaselineCommit != a.ControllerSourceCapsule.HeadCommit ||
		delta.BaselineCommit != a.Expected.ControllerCommit ||
		delta.BaselineTree != a.ControllerSourceCapsule.HeadTree ||
		delta.BaselineTree != a.Expected.ControllerTree ||
		delta.CandidateCommit != a.RuntimeSourceCapsule.HeadCommit ||
		delta.CandidateCommit != a.Expected.Commit ||
		delta.CandidateTree != a.RuntimeSourceCapsule.HeadTree ||
		delta.CandidateTree != a.Expected.Tree {
		return errors.New("candidate delta identity does not match selected B/C source roots")
	}
	return nil
}

func gitCommandBytes(dir string, args ...string) ([]byte, error) {
	home, err := os.MkdirTemp("", "c3-v6-git-home-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(home)
	common := []string{"-c", "core.hooksPath=/dev/null", "-c", "core.attributesFile=/dev/null", "-c", "protocol.file.allow=never"}
	command := protocolCommand("git", append(common, args...)...)
	command.Dir = dir
	command.Env = []string{
		"PATH=" + os.Getenv("PATH"),
		"LC_ALL=C",
		"LANG=C",
		"TZ=UTC",
		"HOME=" + home,
		"XDG_CONFIG_HOME=" + home,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=/dev/null",
		"GIT_OPTIONAL_LOCKS=0",
		"GIT_NO_REPLACE_OBJECTS=1",
		"GIT_ALTERNATE_OBJECT_DIRECTORIES=",
		"GIT_ATTR_NOSYSTEM=1",
		"GIT_CEILING_DIRECTORIES=",
		"GIT_DISCOVERY_ACROSS_FILESYSTEM=0",
	}
	if tmp := os.Getenv("TMPDIR"); tmp != "" {
		command.Env = append(command.Env, "TMPDIR="+tmp)
	}
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return output, nil
}

func gitSHA256(dir string, args ...string) (string, []byte, error) {
	data, err := gitCommandBytes(dir, args...)
	if err != nil {
		return "", nil, err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), data, nil
}

func verifyCandidateDeltaGit(repoRoot, bundlePath string, delta candidateDelta) error {
	if delta.Variable != "direct_hit_containment_owner_substitution" || !validHexN(delta.BaselineCommit, 40) || !validHexN(delta.CandidateCommit, 40) || !validHexN(delta.BaselineTree, 40) || !validHexN(delta.CandidateTree, 40) {
		return errors.New("invalid candidate delta identity")
	}
	registered := []string{"cli/cmd/search.go"}
	if reflect.DeepEqual(delta.AllowedPaths, []string{"cli/cmd/search.go", "cli/cmd/search_test.go"}) {
		registered = delta.AllowedPaths
	} else if !reflect.DeepEqual(delta.AllowedPaths, registered) {
		return errors.New("candidate allowed paths do not match the registered variable")
	}
	status, err := gitCommandBytes(repoRoot, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return err
	}
	if len(status) != 0 {
		return errors.New("candidate repository is dirty")
	}
	parents, err := gitCommandBytes(repoRoot, "rev-list", "--parents", "-n", "1", delta.CandidateCommit)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(parents)) != delta.CandidateCommit+" "+delta.BaselineCommit {
		return errors.New("candidate is not a clean direct child of baseline")
	}
	for commit, tree := range map[string]string{delta.BaselineCommit: delta.BaselineTree, delta.CandidateCommit: delta.CandidateTree} {
		kind, err := gitCommandBytes(repoRoot, "cat-file", "-t", commit)
		if err != nil || strings.TrimSpace(string(kind)) != "commit" {
			return errors.New("candidate delta commit object missing")
		}
		gotTree, err := gitCommandBytes(repoRoot, "rev-parse", commit+"^{tree}")
		if err != nil || strings.TrimSpace(string(gotTree)) != tree {
			return errors.New("candidate delta tree mismatch")
		}
	}
	rangeArg := delta.BaselineCommit + ".." + delta.CandidateCommit
	diffHash, _, err := gitSHA256(repoRoot, "diff", "--binary", "--full-index", "--no-ext-diff", "--no-textconv", "--no-renames", rangeArg)
	if err != nil || diffHash != delta.DiffSHA256 {
		return errors.New("candidate binary diff mismatch")
	}
	nameHash, nameBytes, err := gitSHA256(repoRoot, "diff", "--name-status", "--no-renames", rangeArg)
	if err != nil || nameHash != delta.NameStatusSHA256 {
		return errors.New("candidate name-status mismatch")
	}
	var names []string
	for _, line := range strings.Split(strings.TrimSuffix(string(nameBytes), "\n"), "\n") {
		if line != "" {
			names = append(names, line)
		}
	}
	if !reflect.DeepEqual(names, delta.NameStatus) || len(names) != len(registered) {
		return errors.New("candidate changed-path list mismatch")
	}
	for i, path := range registered {
		if names[i] != "M\t"+path {
			return fmt.Errorf("candidate path is not an exact modification: %s", path)
		}
		for commit, wantHash := range map[string]string{delta.BaselineCommit: delta.BeforeBlobSHA256[path], delta.CandidateCommit: delta.AfterBlobSHA256[path]} {
			entry, err := gitCommandBytes(repoRoot, "ls-tree", commit, "--", path)
			fields := strings.Fields(string(entry))
			if err != nil || len(fields) < 3 || fields[0] != "100644" || fields[1] != "blob" || !validSHA256(wantHash) {
				return fmt.Errorf("candidate path mode/blob invalid: %s", path)
			}
			content, err := gitCommandBytes(repoRoot, "show", commit+":"+path)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(content)
			if hex.EncodeToString(sum[:]) != wantHash {
				return fmt.Errorf("candidate blob SHA-256 mismatch: %s", path)
			}
		}
	}
	if len(delta.BeforeBlobSHA256) != len(registered) || len(delta.AfterBlobSHA256) != len(registered) {
		return errors.New("candidate blob maps contain unregistered paths")
	}
	for path := range delta.BeforeBlobSHA256 {
		if !containsString(registered, path) {
			return errors.New("candidate before-blob map contains an unregistered path")
		}
	}
	for path := range delta.AfterBlobSHA256 {
		if !containsString(registered, path) {
			return errors.New("candidate after-blob map contains an unregistered path")
		}
	}
	summary, err := gitCommandBytes(repoRoot, "diff", "--summary", rangeArg)
	if err != nil || len(summary) != 0 {
		return errors.New("candidate includes a mode or structural path change")
	}
	bundleHash, err := fileSHA256(bundlePath)
	if err != nil || bundleHash != delta.BundleSHA256 {
		return errors.New("candidate bundle SHA-256 mismatch")
	}
	headsHash, headsBytes, err := gitSHA256(repoRoot, "bundle", "list-heads", bundlePath)
	if err != nil || headsHash != delta.BundleHeadsSHA256 {
		return errors.New("candidate bundle heads mismatch")
	}
	headLines := strings.Split(strings.TrimSuffix(string(headsBytes), "\n"), "\n")
	if len(headLines) != 1 {
		return errors.New("candidate bundle must expose exactly one head")
	}
	headFields := strings.Fields(headLines[0])
	if len(headFields) != 2 || headFields[0] != delta.CandidateCommit || !strings.HasPrefix(headFields[1], "refs/c3-eval/commit-pool/") {
		return errors.New("candidate bundle head is not the registered candidate")
	}
	if _, err := gitCommandBytes(repoRoot, "bundle", "verify", bundlePath); err != nil {
		return err
	}
	fresh, err := os.MkdirTemp("", "c3-v6-bundle-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(fresh)
	if _, err := gitCommandBytes(fresh, "init", "--bare", "repo.git"); err != nil {
		return err
	}
	freshGit := filepath.Join(fresh, "repo.git")
	if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "bundle", "unbundle", bundlePath); err != nil {
		return err
	}
	if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "update-ref", "refs/c3-eval/verified-candidate", delta.CandidateCommit); err != nil {
		return err
	}
	for _, object := range []string{delta.BaselineCommit, delta.CandidateCommit, delta.BaselineTree, delta.CandidateTree} {
		if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "cat-file", "-e", object); err != nil {
			return errors.New("candidate bundle is missing a required object")
		}
	}
	if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "fsck", "--full", "--strict", "--no-dangling"); err != nil {
		return err
	}
	return nil
}
func commandBytes(dir string, args ...string) ([]byte, error) {
	c := exec.Command(args[0], args[1:]...)
	c.Dir = dir
	c.Env = append(os.Environ(), "LC_ALL=C", "TZ=UTC")
	out, err := c.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", strings.Join(args, " "), err)
	}
	return out, nil
}
func commandSHA256(dir string, args ...string) (string, error) {
	data, err := commandBytes(dir, args...)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func hermeticGoPaths() (string, string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", "", err
	}
	cacheRoot, err := os.UserCacheDir()
	if err != nil {
		return "", "", "", err
	}
	return home, filepath.Join(home, "go", "pkg", "mod"), filepath.Join(cacheRoot, "go-build"), nil
}

func goCommandBytes(dir string, args ...string) ([]byte, error) {
	home, moduleCache, buildCache, err := hermeticGoPaths()
	if err != nil {
		return nil, err
	}
	goPath, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}
	command := protocolCommand(goPath, args...)
	command.Dir = dir
	command.Env = hermeticGoEnvironment(home, moduleCache, buildCache)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("go %s: %w", strings.Join(args, " "), err)
	}
	if stderr.Len() != 0 {
		return nil, errors.New("hermetic Go command wrote stderr")
	}
	return output, nil
}

func goCommandSHA256(dir string, args ...string) (string, []byte, error) {
	data, err := goCommandBytes(dir, args...)
	if err != nil {
		return "", nil, err
	}
	return shaString(string(data)), data, nil
}

func goExecutableSHA256() (string, error) {
	path, err := exec.LookPath("go")
	if err != nil {
		return "", err
	}
	path, err = filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return fileSHA256(path)
}

func goModVerifySHA256(sourceRoot string) (string, error) {
	hash, output, err := goCommandSHA256(filepath.Join(sourceRoot, "cli"), "mod", "verify")
	if err != nil {
		return "", err
	}
	if string(output) != "all modules verified\n" {
		return "", errors.New("go mod verify returned unexpected output")
	}
	return hash, nil
}

var authoritativeDirtyInputs = []string{"cli/cmd/change.go", "cli/cmd/check_enhanced.go", "cli/cmd/graph.go", "cli/cmd/help.go", "cli/cmd/import.go", "cli/cmd/options.go", "cli/cmd/overlay.go", "cli/cmd/search.go", "cli/internal/changeset/patch.go", "cli/internal/walker/walker.go"}

func actualDirtyPatchSHA256(sourceRoot string) (string, error) {
	args := []string{"diff", "--binary", "--full-index", "--no-ext-diff", "--no-textconv", "--"}
	args = append(args, authoritativeDirtyInputs...)
	hash, _, err := gitSHA256(sourceRoot, args...)
	return hash, err
}

func discoverRepositoryBuildInputs(sourceRoot string) ([]string, error) {
	cliRoot := filepath.Join(sourceRoot, "cli")
	home, moduleCache, buildCache, err := hermeticGoPaths()
	if err != nil {
		return nil, err
	}
	goPath, err := exec.LookPath("go")
	if err != nil {
		return nil, err
	}
	cmdline := protocolCommand(goPath, "list", "-mod=readonly", "-deps", "-json", "./tools/structural-search-eval-v2")
	cmdline.Dir = cliRoot
	cmdline.Env = hermeticGoEnvironment(home, moduleCache, buildCache)
	var stderr bytes.Buffer
	cmdline.Stderr = &stderr
	stdout, err := cmdline.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmdline.Start(); err != nil {
		return nil, err
	}
	dec := json.NewDecoder(stdout)
	set := map[string]bool{}
	cleanRoot := filepath.Clean(sourceRoot) + string(os.PathSeparator)
	for {
		var pkg buildPackage
		if err := dec.Decode(&pkg); errors.Is(err, io.EOF) {
			break
		} else if err != nil {
			return nil, err
		}
		dir := filepath.Clean(pkg.Dir)
		if !strings.HasPrefix(dir, cleanRoot) {
			continue
		}
		for _, name := range append(append(append([]string{}, pkg.GoFiles...), pkg.CgoFiles...), pkg.EmbedFiles...) {
			path := filepath.Join(dir, name)
			rel, err := filepath.Rel(sourceRoot, path)
			if err != nil {
				return nil, err
			}
			set[filepath.ToSlash(rel)] = true
		}
	}
	if err := cmdline.Wait(); err != nil {
		return nil, err
	}
	if stderr.Len() != 0 {
		return nil, errors.New("hermetic go list wrote stderr")
	}
	for _, name := range []string{"cli/go.mod", "cli/go.sum"} {
		set[name] = true
	}
	paths := make([]string, 0, len(set))
	for path := range set {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func captureSourceCapsule(sourceRoot string) (sourceCapsule, error) {
	paths, err := discoverRepositoryBuildInputs(sourceRoot)
	if err != nil {
		return sourceCapsule{}, err
	}
	head, err := gitCommandBytes(sourceRoot, "rev-parse", "HEAD")
	if err != nil {
		return sourceCapsule{}, err
	}
	tree, err := gitCommandBytes(sourceRoot, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return sourceCapsule{}, err
	}
	patch, err := actualDirtyPatchSHA256(sourceRoot)
	if err != nil {
		return sourceCapsule{}, err
	}
	capsule := sourceCapsule{Schema: sourceCapsuleSchema, HeadCommit: strings.TrimSpace(string(head)), HeadTree: strings.TrimSpace(string(tree)), DirtyPatchSHA256: patch}
	for _, path := range paths {
		hash, err := fileSHA256(filepath.Join(sourceRoot, filepath.FromSlash(path)))
		if err != nil {
			return sourceCapsule{}, err
		}
		origin := "head"
		if _, err := gitCommandBytes(sourceRoot, "ls-files", "--error-unmatch", "--", path); err != nil {
			origin = "working-tree"
		} else {
			if _, err := gitCommandBytes(sourceRoot, "diff", "--quiet", "--no-ext-diff", "--no-textconv", "--", path); err != nil {
				origin = "working-tree"
			}
		}
		capsule.Inputs = append(capsule.Inputs, sourceInput{Path: path, SHA256: hash, Origin: origin})
	}
	capsule.RepositoryBuildInputCount = len(capsule.Inputs)
	return capsule, nil
}

func verifySourceCapsuleClosure(sourceRoot string, want sourceCapsule) error {
	got, err := captureSourceCapsule(sourceRoot)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(got, want) {
		return errors.New("source capsule does not match exact repository build closure")
	}
	return validateSourceCapsule(want)
}

func environmentSHA256(sourceRoot string) (string, error) {
	home, moduleCache, buildCache, err := hermeticGoPaths()
	if err != nil {
		return "", err
	}
	goEnv, err := goCommandBytes(filepath.Join(sourceRoot, "cli"), "env", "-json", "GOOS", "GOARCH", "GOVERSION", "GOROOT", "CGO_ENABLED", "GOENV", "GOWORK", "GOFLAGS", "GOTOOLCHAIN", "GOMODCACHE")
	if err != nil {
		return "", err
	}
	goHash, err := goExecutableSHA256()
	if err != nil {
		return "", err
	}
	return protocolEnvironmentIdentitySHA256(goEnv, goHash, hermeticGoEnvironment(home, moduleCache, buildCache))
}
func moduleGraphSHA256(sourceRoot string) (string, error) {
	hash, _, err := goCommandSHA256(filepath.Join(sourceRoot, "cli"), "list", "-mod=readonly", "-m", "all")
	return hash, err
}

func verifyControllerAuthority(a controllerAuthority, runtimePath, controllerPath, sourceRoot, fixturesPath, scorerPath string, bench benchmarkConfig) error {
	if err := validateControllerBudgetLimits(a.BudgetLimits); err != nil {
		return err
	}
	if canonicalSHA256(a.BudgetLimits) != a.Expected.BudgetSHA256 {
		return errors.New("budget contract hash mismatch")
	}
	if shaString(a.ActionEnvelope) != a.Expected.ActionEnvelopeSHA256 {
		return errors.New("action envelope hash mismatch")
	}
	if bench.ContextThresholdAuthority == nil || a.CanonicalRowBytesDefinition != canonicalRowBytesDefinition || shaString(a.CanonicalRowBytesDefinition) != bench.ContextThresholdAuthority.DefinitionSHA256 {
		return errors.New("canonical row byte definition hash mismatch")
	}
	runtimeHash, err := fileSHA256(runtimePath)
	if err != nil {
		return err
	}
	if runtimeHash != a.Expected.RuntimeSHA256 {
		return errors.New("runtime SHA-256 mismatch before spawn")
	}
	controllerHash, err := fileSHA256(controllerPath)
	if err != nil {
		return err
	}
	if controllerHash != a.Expected.ControllerSHA256 {
		return errors.New("controller SHA-256 mismatch before spawn")
	}
	fixtureHash, err := fileSHA256(fixturesPath)
	if err != nil {
		return err
	}
	if fixtureHash != a.Expected.FixtureSHA256 {
		return errors.New("fixture SHA-256 mismatch")
	}
	scorerHash, err := fileSHA256(scorerPath)
	if err != nil {
		return err
	}
	if scorerHash != a.Expected.ScorerSHA256 {
		return errors.New("scorer SHA-256 mismatch")
	}
	if canonicalSHA256(a.SourceCapsule) != a.Expected.SourceCapsuleSHA256 {
		return errors.New("source capsule hash mismatch")
	}
	if err := verifySourceCapsuleClosure(sourceRoot, a.SourceCapsule); err != nil {
		return err
	}
	if a.SourceCapsule.DirtyPatchSHA256 != a.Expected.DiffSHA256 {
		return errors.New("source diff hash mismatch")
	}
	if a.Expected.Commit != a.SourceCapsule.HeadCommit || a.Expected.Tree != a.SourceCapsule.HeadTree {
		return errors.New("source commit or tree mismatch")
	}
	envHash, err := environmentSHA256(sourceRoot)
	if err != nil {
		return err
	}
	if envHash != a.Expected.EnvironmentSHA256 {
		return errors.New("environment hash mismatch")
	}
	moduleHash, err := moduleGraphSHA256(sourceRoot)
	if err != nil {
		return err
	}
	if moduleHash != a.Expected.ModuleGraphSHA256 {
		return errors.New("module graph hash mismatch")
	}
	if a.Expected.ContextThresholdAuthoritySHA256 != hashThresholdAuthority(bench.ContextThresholdAuthority) {
		return errors.New("threshold authority metadata hash mismatch")
	}
	if err := validateThresholdAuthorityRecord(bench.ContextThresholdAuthority, []byte(a.ContextThresholdAuthorityRecord)); err != nil {
		return err
	}
	tmp, err := os.MkdirTemp("", "c3-v2-rebuild-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	rebuilt := filepath.Join(tmp, "runtime")
	build := exec.Command("go", "build", "-trimpath", "-o", rebuilt, "./tools/structural-search-eval-v2")
	build.Dir = filepath.Join(sourceRoot, "cli")
	home, moduleCache, buildCache, err := hermeticGoPaths()
	if err != nil {
		return err
	}
	build.Env = hermeticGoEnvironment(home, moduleCache, buildCache)
	if output, err := build.CombinedOutput(); err != nil {
		return fmt.Errorf("rebuild controller/runtime: %w: %s", err, output)
	}
	rebuiltHash, err := fileSHA256(rebuilt)
	if err != nil {
		return err
	}
	observed := buildReplayManifest{ControllerSHA256: controllerHash, RuntimeSHA256: runtimeHash, RebuiltControllerSHA256: rebuiltHash, RebuiltRuntimeSHA256: rebuiltHash, SourceCapsuleRebuildVerified: true, BundleVerified: false}
	if !reflect.DeepEqual(observed, a.BuildReplay) {
		return errors.New("build replay manifest mismatch")
	}
	return validateBuildReplay(observed)
}

func verifyControllerAuthorityV3(a controllerAuthorityV3, runtimePath, controllerPath, controllerSourceRoot, runtimeSourceRoot, bundlePath, fixturesPath, benchmarkPath, scorerPath string, bench benchmarkConfig) error {
	if a.Schema != controllerAuthorityV3Schema || (a.Mode != "baseline" && a.Mode != "candidate") {
		return errors.New("invalid controller authority v3 mode")
	}
	for _, path := range []string{fixturesPath, benchmarkPath, scorerPath} {
		if !pathWithin(controllerSourceRoot, path) {
			return errors.New("v3 fixture, benchmark, and scorer must come from controller root B")
		}
	}
	if a.Mode == "candidate" && pathsOverlap(controllerSourceRoot, runtimeSourceRoot) {
		return errors.New("candidate controller and runtime roots overlap")
	}
	if err := verifyPortableAuthorityPrivacy(a); err != nil {
		return err
	}
	if err := validateControllerBudgetLimits(a.BudgetLimits); err != nil {
		return err
	}
	if canonicalSHA256(a.BudgetLimits) != a.Expected.BudgetSHA256 || shaString(a.ActionEnvelope) != a.Expected.ActionEnvelopeSHA256 {
		return errors.New("v3 budget or action-envelope hash mismatch")
	}
	if bench.ContextThresholdAuthority == nil || a.CanonicalRowBytesDefinition != canonicalRowBytesDefinition || shaString(a.CanonicalRowBytesDefinition) != bench.ContextThresholdAuthority.DefinitionSHA256 {
		return errors.New("v3 canonical row byte definition mismatch")
	}
	controllerHash, err := fileSHA256(controllerPath)
	if err != nil || controllerHash != a.Expected.ControllerSHA256 {
		return errors.New("controller SHA-256 mismatch before spawn")
	}
	runtimeHash, err := fileSHA256(runtimePath)
	if err != nil || runtimeHash != a.Expected.RuntimeSHA256 {
		return errors.New("runtime SHA-256 mismatch before spawn")
	}
	fixtureHash, err := fileSHA256(fixturesPath)
	if err != nil || fixtureHash != a.Expected.FixtureSHA256 {
		return errors.New("fixture SHA-256 mismatch")
	}
	benchmarkHash, err := fileSHA256(benchmarkPath)
	if err != nil || benchmarkHash != a.Expected.BenchmarkSHA256 {
		return errors.New("benchmark SHA-256 mismatch")
	}
	scorerHash, err := fileSHA256(scorerPath)
	if err != nil || scorerHash != a.Expected.ScorerSHA256 {
		return errors.New("scorer SHA-256 mismatch")
	}
	scoringHash, err := scoringRegionSHA256(scorerPath)
	if err != nil || scoringHash != "c1669d43b13a36eb1454a01a5ea7e444fe67ce28fdae9628da083927f093cbcb" {
		return errors.New("frozen scoring region mismatch")
	}
	if canonicalSHA256(a.ControllerSourceCapsule) != a.Expected.ControllerSourceCapsuleSHA256 || canonicalSHA256(a.RuntimeSourceCapsule) != a.Expected.SourceCapsuleSHA256 {
		return errors.New("dual-source capsule hash mismatch")
	}
	if err := verifyCleanSourceCapsuleClosure(controllerSourceRoot, a.ControllerSourceCapsule); err != nil {
		return fmt.Errorf("controller source closure: %w", err)
	}
	if err := verifyCleanSourceCapsuleClosure(runtimeSourceRoot, a.RuntimeSourceCapsule); err != nil {
		return fmt.Errorf("runtime source closure: %w", err)
	}
	if a.ControllerSourceCapsule.HeadCommit != a.Expected.ControllerCommit || a.ControllerSourceCapsule.HeadTree != a.Expected.ControllerTree || a.RuntimeSourceCapsule.HeadCommit != a.Expected.Commit || a.RuntimeSourceCapsule.HeadTree != a.Expected.Tree {
		return errors.New("dual-source commit or tree mismatch")
	}
	controllerEnvironment, err := environmentSHA256(controllerSourceRoot)
	if err != nil {
		return err
	}
	runtimeEnvironment, err := environmentSHA256(runtimeSourceRoot)
	if err != nil {
		return err
	}
	if controllerEnvironment != runtimeEnvironment || controllerEnvironment != a.Expected.EnvironmentSHA256 {
		return errors.New("dual-source environment mismatch")
	}
	controllerModules, err := moduleGraphSHA256(controllerSourceRoot)
	if err != nil {
		return err
	}
	runtimeModules, err := moduleGraphSHA256(runtimeSourceRoot)
	if err != nil {
		return err
	}
	if controllerModules != runtimeModules || controllerModules != a.Expected.ModuleGraphSHA256 {
		return errors.New("dual-source module graph mismatch")
	}
	if a.Expected.ContextThresholdAuthoritySHA256 != hashThresholdAuthority(bench.ContextThresholdAuthority) {
		return errors.New("threshold authority metadata hash mismatch")
	}
	if err := validateThresholdAuthorityRecord(bench.ContextThresholdAuthority, []byte(a.ContextThresholdAuthorityRecord)); err != nil {
		return err
	}
	if a.Expected.BundleSHA256 == "" || !validSHA256(a.Expected.BundleSHA256) {
		return errors.New("portable bundle hash is missing")
	}
	if a.Mode == "candidate" {
		if a.CandidateDelta == nil {
			return errors.New("candidate authority has no delta")
		}
		if canonicalCandidateDeltaSHA256(*a.CandidateDelta) != a.Expected.CandidateDeltaSHA256 || a.CandidateDelta.BundleSHA256 != a.Expected.BundleSHA256 {
			return errors.New("candidate delta authority mismatch")
		}
		if err := validateCandidateDeltaAuthorityBinding(a); err != nil {
			return err
		}
		if err := validateDualSourceCapsules(a.ControllerSourceCapsule, a.RuntimeSourceCapsule, a.CandidateDelta.AllowedPaths); err != nil {
			return err
		}
		if err := verifyCandidateDeltaGit(runtimeSourceRoot, bundlePath, *a.CandidateDelta); err != nil {
			return err
		}
	} else {
		if a.CandidateDelta != nil || a.Expected.CandidateDeltaSHA256 != "" || a.Expected.ControllerCommit != a.Expected.Commit || a.Expected.ControllerTree != a.Expected.Tree || !reflect.DeepEqual(a.ControllerSourceCapsule, a.RuntimeSourceCapsule) {
			return errors.New("baseline authority is not an exact B/B pair")
		}
		if err := verifyBaselineBundle(runtimeSourceRoot, bundlePath, a.Expected.BundleSHA256, a.Expected.Commit); err != nil {
			return err
		}
	}
	if err := verifyPortableTreePrivacy(controllerSourceRoot, a.Expected.ControllerCommit); err != nil {
		return err
	}
	if a.Expected.Commit != a.Expected.ControllerCommit {
		if err := verifyPortableTreePrivacy(runtimeSourceRoot, a.Expected.Commit); err != nil {
			return err
		}
	}
	tmp, err := os.MkdirTemp("", "c3-v3-rebuild-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)
	rebuiltController := filepath.Join(tmp, "controller")
	rebuiltRuntime := filepath.Join(tmp, "runtime")
	if err := buildFrozenRuntime(controllerSourceRoot, rebuiltController); err != nil {
		return fmt.Errorf("rebuild controller: %w", err)
	}
	if err := buildFrozenRuntime(runtimeSourceRoot, rebuiltRuntime); err != nil {
		return fmt.Errorf("rebuild runtime: %w", err)
	}
	rebuiltControllerHash, err := fileSHA256(rebuiltController)
	if err != nil {
		return err
	}
	rebuiltRuntimeHash, err := fileSHA256(rebuiltRuntime)
	if err != nil {
		return err
	}
	observed := buildReplayManifest{
		ControllerSHA256: controllerHash, RuntimeSHA256: runtimeHash,
		RebuiltControllerSHA256: rebuiltControllerHash, RebuiltRuntimeSHA256: rebuiltRuntimeHash,
		ControllerCapsuleRebuildVerified: true, RuntimeCapsuleRebuildVerified: true, BundleVerified: true,
	}
	if !reflect.DeepEqual(observed, a.BuildReplay) {
		return errors.New("dual-source build replay manifest mismatch")
	}
	return validateBuildReplay(observed)
}

func registeredPathInside(root, path, wantRelative string) bool {
	realRoot, err := resolvedAbsolutePath(root)
	if err != nil {
		return false
	}
	realPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 {
		return false
	}
	relative, err := filepath.Rel(realRoot, realPath)
	return err == nil && filepath.ToSlash(relative) == wantRelative
}

func readBoundedStandaloneRegularFile(file string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("invalid standalone governed file cap")
	}
	absolute, err := filepath.Abs(file)
	if err != nil {
		return nil, errors.New("invalid standalone governed file")
	}
	absolute = filepath.Clean(absolute)
	real, err := filepath.EvalSymlinks(absolute)
	if err != nil || real != absolute {
		return nil, errors.New("standalone governed file uses a symlink alias")
	}
	info, err := os.Lstat(absolute)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 || info.Size() > maxBytes {
		return nil, errors.New("invalid or oversized standalone governed file")
	}
	fd, err := unix.Open(absolute, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, errors.New("cannot pin standalone governed file")
	}
	pinned := os.NewFile(uintptr(fd), "standalone-governed-file")
	defer pinned.Close()
	pinnedInfo, err := pinned.Stat()
	if err != nil || !pinnedInfo.Mode().IsRegular() || !os.SameFile(info, pinnedInfo) || pinnedInfo.Size() > maxBytes {
		return nil, errors.New("standalone governed file changed before read")
	}
	data, err := io.ReadAll(io.LimitReader(pinned, maxBytes+1))
	if err != nil || int64(len(data)) != pinnedInfo.Size() {
		return nil, errors.New("standalone governed file changed while reading")
	}
	return data, nil
}

func readBoundedRegularFileInside(root, file string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, errors.New("invalid governed file byte cap")
	}
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, errors.New("invalid governed root")
	}
	absoluteRoot = filepath.Clean(absoluteRoot)
	realRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil || realRoot != absoluteRoot {
		return nil, errors.New("governed root uses a symlink alias")
	}
	rootInfo, err := os.Lstat(absoluteRoot)
	if err != nil || !rootInfo.IsDir() || rootInfo.Mode()&os.ModeSymlink != 0 {
		return nil, errors.New("invalid governed root")
	}
	absoluteFile, err := filepath.Abs(file)
	if err != nil {
		return nil, errors.New("invalid governed file")
	}
	absoluteFile = filepath.Clean(absoluteFile)
	relative, err := filepath.Rel(absoluteRoot, absoluteFile)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return nil, errors.New("governed file escapes its root")
	}
	cursor := absoluteRoot
	for _, component := range strings.Split(relative, string(os.PathSeparator)) {
		cursor = filepath.Join(cursor, component)
		info, err := os.Lstat(cursor)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return nil, errors.New("governed file path contains a symlink")
		}
	}
	rootFD, err := unix.Open(absoluteRoot, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC|unix.O_NOFOLLOW, 0)
	if err != nil {
		return nil, errors.New("cannot pin governed root")
	}
	defer unix.Close(rootFD)
	fileFD, err := unix.Openat2(rootFD, relative, &unix.OpenHow{
		Flags:   unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW,
		Resolve: unix.RESOLVE_BENEATH | unix.RESOLVE_NO_SYMLINKS | unix.RESOLVE_NO_MAGICLINKS,
	})
	if err != nil {
		return nil, errors.New("cannot pin governed file")
	}
	pinned := os.NewFile(uintptr(fileFD), "governed-file")
	defer pinned.Close()
	info, err := pinned.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > maxBytes {
		return nil, errors.New("invalid or oversized governed file")
	}
	data, err := io.ReadAll(io.LimitReader(pinned, maxBytes+1))
	if err != nil || int64(len(data)) != info.Size() {
		return nil, errors.New("governed file changed while reading")
	}
	return data, nil
}

func parentRelativeArtifact(root, relative string, maxBytes int64) ([]byte, error) {
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(relative)))
	if relative == "" || relative == "." || clean != relative || filepath.IsAbs(relative) || strings.Contains(relative, `\`) || strings.HasPrefix(relative, "../") {
		return nil, errors.New("invalid parent artifact path")
	}
	return readBoundedRegularFileInside(root, filepath.Join(root, filepath.FromSlash(relative)), maxBytes)
}

func validateParentBaselineBinding(binding *parentBaselineBinding, expectedRuns int) error {
	if binding == nil || binding.RunCount != expectedRuns || expectedRuns <= 0 {
		return errors.New("invalid parent baseline binding")
	}
	for _, hash := range []string{
		binding.AuthoritySHA256, binding.OutputSHA256, binding.OrderedRunSHA256,
		binding.HistorySHA256, binding.HistoryTailRecordHash, binding.PrivacyManifestSHA256,
		binding.ValidatorRecordHash, binding.ValidatorPayloadSHA256,
	} {
		if !validSHA256(hash) || allZero(hash) {
			return errors.New("invalid parent baseline hash binding")
		}
	}
	return nil
}

func verifyParentFrozenEquality(candidate, parent controllerAuthorityV4) error {
	if parent.Schema != controllerAuthorityV4Schema || parent.Mode != "baseline" || parent.ParentBaseline != nil || parent.CandidateDelta != nil {
		return errors.New("parent authority is not a protocol-v7 baseline")
	}
	if candidate.CandidateDelta == nil || candidate.CandidateDelta.BaselineCommit != parent.Expected.Commit || candidate.CandidateDelta.BaselineTree != parent.Expected.Tree {
		return errors.New("candidate delta does not descend from its parent baseline")
	}
	if !reflect.DeepEqual(parent.ControllerSourceCapsule, parent.RuntimeSourceCapsule) || !reflect.DeepEqual(candidate.ControllerSourceCapsule, parent.ControllerSourceCapsule) {
		return errors.New("candidate controller capsule differs from its parent baseline")
	}
	if parent.Expected.Commit != parent.Expected.ControllerCommit || parent.Expected.Tree != parent.Expected.ControllerTree ||
		candidate.Expected.ControllerCommit != parent.Expected.Commit || candidate.Expected.ControllerTree != parent.Expected.Tree ||
		candidate.Expected.ControllerSourceCapsuleSHA256 != parent.Expected.SourceCapsuleSHA256 ||
		candidate.Expected.ControllerSHA256 != parent.Expected.RuntimeSHA256 || parent.Expected.ControllerSHA256 != parent.Expected.RuntimeSHA256 {
		return errors.New("candidate controller identity differs from its parent baseline")
	}
	if candidate.BuildReplay.ControllerSHA256 != parent.BuildReplay.ControllerSHA256 ||
		candidate.BuildReplay.RebuiltControllerSHA256 != parent.BuildReplay.RebuiltControllerSHA256 ||
		parent.BuildReplay.ControllerSHA256 != parent.BuildReplay.RuntimeSHA256 ||
		parent.BuildReplay.RebuiltControllerSHA256 != parent.BuildReplay.RebuiltRuntimeSHA256 {
		return errors.New("candidate controller binary differs from its parent baseline")
	}
	candidateFrozen := []any{
		candidate.Expected.FixtureSHA256, candidate.Expected.ScorerSHA256, candidate.Expected.ControllerSHA256,
		candidate.Expected.EnvironmentSHA256, candidate.Expected.ModuleGraphSHA256, candidate.Expected.BudgetSHA256,
		candidate.Expected.ActionEnvelopeSHA256, candidate.Expected.SemanticMode, candidate.Expected.ContextThresholdAuthoritySHA256,
		candidate.Expected.PrivacyPolicySHA256, candidate.Expected.PrivacyTermCount, candidate.Expected.PrivacyDetectorSHA256,
		candidate.Expected.GoExecutableSHA256, candidate.Expected.GoModVerifySHA256, candidate.Expected.ScanCapsSHA256, candidate.Expected.ProtocolTestSHA256,
		candidate.Expected.BenchmarkSHA256, candidate.BudgetLimits, candidate.ActionEnvelope,
		candidate.CanonicalRowBytesDefinition, candidate.ContextThresholdAuthorityRecord, candidate.ScanCaps,
		candidate.PrivacyPolicySHA256, candidate.PrivacyTermCount, candidate.PrivacyDetectorSHA256,
		candidate.GoExecutableSHA256, candidate.GoModVerifySHA256, candidate.ProtocolTestSHA256,
	}
	parentFrozen := []any{
		parent.Expected.FixtureSHA256, parent.Expected.ScorerSHA256, parent.Expected.RuntimeSHA256,
		parent.Expected.EnvironmentSHA256, parent.Expected.ModuleGraphSHA256, parent.Expected.BudgetSHA256,
		parent.Expected.ActionEnvelopeSHA256, parent.Expected.SemanticMode, parent.Expected.ContextThresholdAuthoritySHA256,
		parent.Expected.PrivacyPolicySHA256, parent.Expected.PrivacyTermCount, parent.Expected.PrivacyDetectorSHA256,
		parent.Expected.GoExecutableSHA256, parent.Expected.GoModVerifySHA256, parent.Expected.ScanCapsSHA256, parent.Expected.ProtocolTestSHA256,
		parent.Expected.BenchmarkSHA256, parent.BudgetLimits, parent.ActionEnvelope,
		parent.CanonicalRowBytesDefinition, parent.ContextThresholdAuthorityRecord, parent.ScanCaps,
		parent.PrivacyPolicySHA256, parent.PrivacyTermCount, parent.PrivacyDetectorSHA256,
		parent.GoExecutableSHA256, parent.GoModVerifySHA256, parent.ProtocolTestSHA256,
	}
	if !reflect.DeepEqual(candidateFrozen, parentFrozen) {
		return errors.New("candidate frozen inputs differ from its parent baseline")
	}
	return nil
}

func expectedParentRun(fixtures []fixtureCase, bench benchmarkConfig, index int) ([]fixtureCase, string, string, error) {
	if index < len(fixtures) {
		return []fixtureCase{fixtures[index]}, corpusIsolated, fixtures[index].CaseID, nil
	}
	if index == len(fixtures) {
		return fixtures, corpusCombined, "", nil
	}
	if index == len(fixtures)+1 {
		scaled := append([]fixtureCase(nil), fixtures...)
		for i := range scaled {
			cfg := bench.Scale
			cfg.Seed += i * 1000
			scaled[i].Corpus = generateScaleCorpus(scaled[i].Corpus, cfg)
		}
		return scaled, "scale", "", nil
	}
	return nil, "", "", errors.New("unexpected parent run index")
}

func verifyParentRunArtifacts(root string, output controllerOutput, history []historyRecord, parent controllerAuthorityV4, fixtures []fixtureCase, bench benchmarkConfig, scanner *privacyScanner) (map[string][]byte, error) {
	known := map[string][]byte{}
	seenIdentity := map[string]bool{}
	for index, run := range output.Runs {
		selected, mode, caseID, err := expectedParentRun(fixtures, bench, index)
		if err != nil || run.Mode != mode || run.CaseID != caseID {
			return nil, errors.New("parent ordered run plan mismatch")
		}
		identity := run.Mode + "\x00" + run.CaseID
		if seenIdentity[identity] {
			return nil, errors.New("duplicate parent run identity")
		}
		seenIdentity[identity] = true
		resultBytes, err := parentRelativeArtifact(root, run.ResultPath, parent.ScanCaps.SingleDurableArtifactBytesMax)
		if err != nil || shaString(string(resultBytes)) != run.ResultSHA256 {
			return nil, errors.New("parent result artifact mismatch")
		}
		reportBytes, err := parentRelativeArtifact(root, run.ReportPath, parent.ScanCaps.SingleDurableArtifactBytesMax)
		if err != nil || shaString(string(reportBytes)) != run.ReportSHA256 {
			return nil, errors.New("parent report artifact mismatch")
		}
		if run.ResultPath == run.ReportPath || known[run.ResultPath] != nil || known[run.ReportPath] != nil {
			return nil, errors.New("duplicate parent artifact path")
		}
		known[run.ResultPath], known[run.ReportPath] = resultBytes, reportBytes
		if err := scanner.Scan("parent_result", run.ResultPath, resultBytes); err != nil {
			return nil, err
		}
		if err := scanner.Scan("parent_report", run.ReportPath, reportBytes); err != nil {
			return nil, err
		}
		var response armResponse
		if err := decodeStrictBytes(resultBytes, &response); err != nil || response.Schema != armResponseSchema {
			return nil, errors.New("invalid parent arm response")
		}
		var report durableReport
		if err := decodeStrictBytes(reportBytes, &report); err != nil || report.Schema != durableReportV4Schema {
			return nil, errors.New("invalid parent durable report")
		}
		if err := validateLogicalDump(report.Database); err != nil {
			return nil, err
		}
		if err := verifyReportWithProbes(selected, response, report.Database, mode, report.DirectProbes, report.Report); err != nil {
			return nil, fmt.Errorf("parent report replay: %w", err)
		}
		if report.ResultPath != run.ResultPath || report.ResultSHA256 != run.ResultSHA256 || report.BudgetLimits != parent.BudgetLimits ||
			report.ActualBudget != run.ActualBudget || report.CanonicalRowBytesDefinition != parent.CanonicalRowBytesDefinition ||
			!reflect.DeepEqual(report.ContextThresholdAuthority, bench.ContextThresholdAuthority) ||
			report.ContextThresholdAuthorityLineSHA256 != shaString(parent.ContextThresholdAuthorityRecord) ||
			report.BudgetVerdict != map[bool]string{true: "within_limits", false: "over_limit"}[classifyBudget(parent.BudgetLimits, run.ActualBudget) == "score"] {
			return nil, errors.New("parent durable report authority mismatch")
		}
		historyRow := history[index]
		if historyRow.RecordHash != run.HistoryRecordHash || historyRow.ResultPath != run.ResultPath || historyRow.ResultSHA256 != run.ResultSHA256 ||
			historyRow.Budgets != run.ActualBudget || historyRow.ParentKeep != "GENESIS" || historyRow.ChangedVariable != "baseline" ||
			len(historyRow.ChangedPaths) != 0 || historyRow.Status != "invalid" || historyRow.Reason != "diagnostic_unadmitted" ||
			historyRow.ErrorClass != "" || historyRow.ErrorSHA256 != "" || run.CandidateDeltaSHA256 != "" || run.BundleSHA256 != parent.Expected.BundleSHA256 {
			return nil, errors.New("parent history row does not match its run")
		}
		wantEvidence := []string{run.ReportPath + "#sha256=" + run.ReportSHA256, bench.ContextThresholdAuthority.CheckinRef + "#sha256=" + bench.ContextThresholdAuthority.CheckinSHA256}
		if !reflect.DeepEqual(historyRow.Evidence, wantEvidence) || report.Report.Provenance == nil {
			return nil, errors.New("parent history evidence mismatch")
		}
		wantProvenance := parent.persistenceAuthority().Expected
		wantProvenance.ExperimentID = fmt.Sprintf("%s-%02d", parent.Expected.ExperimentID, index+1)
		wantProvenance.ArmID = mode
		if caseID != "" {
			wantProvenance.ArmID += ":" + caseID
		}
		wantProvenance.LogicalDumpSHA256 = report.Database.LogicalSHA256
		wantProvenance.CorpusMode = mode
		wantProvenance.ProjectDirSHA256 = historyRow.Provenance.ProjectDirSHA256
		wantProvenance.C3DirSHA256 = historyRow.Provenance.C3DirSHA256
		wantProvenance.ContextThresholdAuthorityRecord = []byte(parent.ContextThresholdAuthorityRecord)
		historyProvenance := historyRow.Provenance
		historyProvenance.ContextThresholdAuthorityRecord = []byte(parent.ContextThresholdAuthorityRecord)
		reportProvenance := *report.Report.Provenance
		reportProvenance.ContextThresholdAuthorityRecord = []byte(parent.ContextThresholdAuthorityRecord)
		if !reflect.DeepEqual(historyProvenance, wantProvenance) || !reflect.DeepEqual(reportProvenance, wantProvenance) {
			return nil, errors.New("parent run provenance mismatch")
		}
		if err := validateBaselineAdmission(bench, wantProvenance); err != nil {
			return nil, err
		}
	}
	return known, nil
}

func fullReachablePathArtifacts(sourceRoot, commit string, caps privacyScanCaps) (map[string]privacyArtifact, map[string][]byte, error) {
	commits, err := gitCommandBytes(sourceRoot, "rev-list", "--reverse", commit)
	if err != nil {
		return nil, nil, err
	}
	artifacts := map[string]privacyArtifact{}
	contents := map[string][]byte{}
	pathCount := 0
	for _, commitID := range strings.Fields(string(commits)) {
		if !validHexN(commitID, 40) {
			return nil, nil, errors.New("invalid source history commit identity")
		}
		raw, err := gitCommandBytes(sourceRoot, "ls-tree", "-r", "-z", "--name-only", commitID)
		if err != nil {
			return nil, nil, err
		}
		for index, path := range bytes.Split(raw, []byte{0}) {
			if len(path) == 0 {
				continue
			}
			pathBytes := append([]byte(nil), path...)
			pathCount++
			if !utf8.Valid(pathBytes) || len(pathBytes) > caps.SinglePathUTF8BytesMax || pathCount > caps.TreePathCountMax {
				return nil, nil, errors.New("source tree paths exceed protocol-v7 cap")
			}
			genericPath := fmt.Sprintf("commits/%s/paths/%06d", commitID, index+1)
			artifacts[genericPath] = privacyArtifact{Role: "source_tree_path", Path: genericPath, SHA256: shaString(string(pathBytes)), Bytes: len(pathBytes)}
			contents[genericPath] = pathBytes
		}
	}
	return artifacts, contents, nil
}

func parentReachableSourceArtifacts(sourceRoot, commit string, caps privacyScanCaps) (map[string]privacyArtifact, map[string][]byte, int, int64, error) {
	objects, err := gitCommandBytes(sourceRoot, "rev-list", "--objects", "--no-object-names", commit)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	artifacts := map[string]privacyArtifact{}
	contents := map[string][]byte{}
	objectCount := 0
	var objectBytes int64
	for _, objectID := range strings.Fields(string(objects)) {
		if !validHexN(objectID, 40) {
			return nil, nil, 0, 0, errors.New("invalid parent source object identity")
		}
		objectTypeBytes, err := gitCommandBytes(sourceRoot, "cat-file", "-t", objectID)
		if err != nil {
			return nil, nil, 0, 0, err
		}
		objectType := strings.TrimSpace(string(objectTypeBytes))
		if objectType != "blob" && objectType != "tree" && objectType != "commit" && objectType != "tag" {
			return nil, nil, 0, 0, errors.New("unsupported parent source object type")
		}
		sizeBytes, err := gitCommandBytes(sourceRoot, "cat-file", "-s", objectID)
		if err != nil {
			return nil, nil, 0, 0, err
		}
		size, err := strconv.ParseInt(strings.TrimSpace(string(sizeBytes)), 10, 64)
		if err != nil || size < 0 {
			return nil, nil, 0, 0, errors.New("invalid parent source object size")
		}
		content, err := gitCommandBytes(sourceRoot, "cat-file", "-p", objectID)
		if err != nil {
			return nil, nil, 0, 0, err
		}
		path := "objects/" + objectID
		artifacts[path] = privacyArtifact{Role: "source_object_" + objectType, Path: path, SHA256: shaString(string(content)), Bytes: len(content)}
		contents[path] = content
		objectCount++
		objectBytes += size
	}
	pathArtifacts, pathContents, err := fullReachablePathArtifacts(sourceRoot, commit, caps)
	if err != nil {
		return nil, nil, 0, 0, err
	}
	for genericPath, artifact := range pathArtifacts {
		artifacts[genericPath] = artifact
		contents[genericPath] = pathContents[genericPath]
	}
	return artifacts, contents, objectCount, objectBytes, nil
}

func verifyParentPrivacyManifest(root string, output controllerOutput, parent controllerAuthorityV4, authorityBytes []byte, known map[string][]byte, controllerSourceRoot, fixturesPath, benchmarkPath, scorerPath string, scanner *privacyScanner) error {
	manifestBytes, err := parentRelativeArtifact(root, output.PrivacyManifestPath, parent.ScanCaps.SingleDurableArtifactBytesMax)
	if err != nil || shaString(string(manifestBytes)) != output.PrivacyManifestSHA256 {
		return errors.New("parent privacy manifest hash mismatch")
	}
	if err := scanner.Scan("parent_privacy_manifest", output.PrivacyManifestPath, manifestBytes); err != nil {
		return err
	}
	var manifest privacyManifest
	if err := decodeStrictBytes(manifestBytes, &manifest); err != nil {
		return errors.New("invalid parent privacy manifest")
	}
	if manifest.Schema != "structural-retrieval-privacy-scan.v1" || manifest.PrivacyPolicySHA256 != parent.PrivacyPolicySHA256 ||
		manifest.PrivacyTermCount != parent.PrivacyTermCount || manifest.DetectorVersion != privacyDetectorVersion ||
		manifest.DetectorSHA256 != parent.PrivacyDetectorSHA256 || manifest.ScanCaps != parent.ScanCaps || manifest.Hits != 0 ||
		manifest.ArtifactCount != len(manifest.Artifacts) || manifest.ArtifactCount > parent.ScanCaps.DurableArtifactCountMax {
		return errors.New("parent privacy manifest authority mismatch")
	}
	expected := map[string]privacyArtifact{
		"authority\x00controller-authority.v4.json": {Role: "authority", Path: "controller-authority.v4.json", SHA256: shaString(string(authorityBytes)), Bytes: len(authorityBytes)},
	}
	content := map[string][]byte{"authority\x00controller-authority.v4.json": authorityBytes}
	for _, frozen := range []struct {
		role string
		path string
		file string
	}{
		{"frozen_fixture", "research/eval/structural-retrieval/fixtures.dev.v2.jsonl", fixturesPath},
		{"frozen_benchmark", "research/eval/structural-retrieval/benchmark.v2.json", benchmarkPath},
		{"frozen_scorer", "cli/tools/structural-search-eval-v2/main.go", scorerPath},
		{"frozen_protocol_test", "cli/tools/structural-search-eval-v2/main_test.go", filepath.Join(filepath.Dir(scorerPath), "main_test.go")},
	} {
		data, err := readBoundedRegularFileInside(controllerSourceRoot, frozen.file, parent.ScanCaps.SingleDurableArtifactBytesMax)
		if err != nil {
			return err
		}
		key := frozen.role + "\x00" + frozen.path
		expected[key] = privacyArtifact{Role: frozen.role, Path: frozen.path, SHA256: shaString(string(data)), Bytes: len(data)}
		content[key] = data
	}
	sourceArtifacts, sourceContents, sourceObjectCount, sourceBytes, err := parentReachableSourceArtifacts(controllerSourceRoot, parent.Expected.Commit, parent.ScanCaps)
	if err != nil {
		return err
	}
	for path, artifact := range sourceArtifacts {
		key := artifact.Role + "\x00" + path
		expected[key] = artifact
		content[key] = sourceContents[path]
	}
	for path, data := range known {
		role := ""
		switch {
		case path == output.HistoryPath:
			role = "history"
		case strings.HasPrefix(path, "results/"):
			role = "result"
		case strings.HasPrefix(path, "reports/"):
			role = "report"
		case strings.HasPrefix(path, "runtime/"):
			role = "runtime_stderr"
		default:
			return errors.New("unknown parent durable artifact role")
		}
		key := role + "\x00" + path
		expected[key] = privacyArtifact{Role: role, Path: path, SHA256: shaString(string(data)), Bytes: len(data)}
		content[key] = data
	}
	if manifest.SourceObjectCount != sourceObjectCount || manifest.SourceObjectBytes != sourceBytes || len(expected) != len(manifest.Artifacts) {
		return fmt.Errorf("parent privacy manifest coverage mismatch: artifacts=%d/%d source_objects=%d/%d source_bytes=%d/%d", len(manifest.Artifacts), len(expected), manifest.SourceObjectCount, sourceObjectCount, manifest.SourceObjectBytes, sourceBytes)
	}
	var total int64
	seen := map[string]bool{}
	for index, artifact := range manifest.Artifacts {
		if index > 0 && !privacyArtifactLess(manifest.Artifacts[index-1], artifact) {
			return errors.New("parent privacy manifest ordering mismatch")
		}
		key := artifact.Role + "\x00" + artifact.Path
		want, ok := expected[key]
		if !ok || seen[key] || artifact != want {
			return errors.New("parent privacy manifest artifact mismatch")
		}
		seen[key] = true
		total += int64(artifact.Bytes)
		if err := scanner.Scan("parent_replay_"+artifact.Role, artifact.Path, content[key]); err != nil {
			return err
		}
	}
	if total != manifest.TotalScannedBytes {
		return errors.New("parent privacy manifest byte count mismatch")
	}
	return nil
}

func parseValidatorRecordRef(ref string) (string, int, error) {
	const prefix = "workers/validator-baseline-protocol-v7/progress.jsonl#seq="
	if !strings.HasPrefix(ref, prefix) {
		return "", 0, errors.New("invalid parent validator record ref")
	}
	seq, err := strconv.Atoi(strings.TrimPrefix(ref, prefix))
	if err != nil || seq <= 0 || ref != prefix+strconv.Itoa(seq) {
		return "", 0, errors.New("invalid parent validator record sequence")
	}
	return strings.TrimSuffix(prefix, "#seq="), seq, nil
}

func verifyParentValidatorRecord(files parentBaselineFiles, binding parentBaselineBinding, acceptance baselineAcceptance, scorerPath, protocolTestPath string) error {
	relative, seq, err := parseValidatorRecordRef(binding.ValidatorRecordRef)
	if err != nil {
		return err
	}
	recordBytes, err := readBoundedRegularFileInside(files.ValidatorStore, filepath.Join(files.ValidatorStore, filepath.FromSlash(relative)), 4<<20)
	if err != nil {
		return err
	}
	payloadMap, err := verifyWorkerRecordPrefix(bytes.NewReader(recordBytes), seq, binding.ValidatorRecordHash, binding.ValidatorPayloadSHA256)
	if err != nil {
		return err
	}
	payloadBytes, err := json.Marshal(payloadMap)
	if err != nil {
		return err
	}
	var payload baselineValidatorPayload
	if err := decodeStrictBytes(payloadBytes, &payload); err != nil {
		return errors.New("invalid parent validator payload")
	}
	if payload.Event != "finish" || payload.WorkerID != "validator-baseline-protocol-v7" || payload.Role != "independent baseline validator" || payload.Status != "accepted" || payload.EffectClaim || !reflect.DeepEqual(payload.BaselineAcceptance, acceptance) {
		return errors.New("parent validator did not accept this baseline")
	}
	mainHash, err := fileSHA256(scorerPath)
	if err != nil {
		return err
	}
	testHash, err := fileSHA256(protocolTestPath)
	if err != nil {
		return err
	}
	if acceptance.Schema != "structural-retrieval-baseline-acceptance.v1" || acceptance.Verdict != "accepted" ||
		acceptance.ValidatedSourceMainSHA256 != mainHash || acceptance.ValidatedSourceTestSHA256 != testHash {
		return errors.New("parent validator source binding mismatch")
	}
	return nil
}

func verifyParentRootCoverage(files parentBaselineFiles, output controllerOutput) error {
	absoluteRoot, err := filepath.Abs(files.Root)
	if err != nil {
		return err
	}
	allowed := map[string]bool{output.HistoryPath: true, output.PrivacyManifestPath: true}
	for _, run := range output.Runs {
		allowed[run.ResultPath] = true
		allowed[run.ReportPath] = true
	}
	for index := range output.Runs {
		allowed[fmt.Sprintf("runtime/%02d.stderr", index+1)] = true
	}
	for _, administrative := range []string{files.Authority, files.Output} {
		absolute, err := filepath.Abs(administrative)
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(absoluteRoot, absolute)
		if err != nil {
			return err
		}
		allowed[filepath.ToSlash(relative)] = true
	}
	seen := map[string]bool{}
	err = filepath.WalkDir(absoluteRoot, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return errors.New("parent baseline tree contains a symlink")
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil || !info.Mode().IsRegular() {
			return errors.New("parent baseline tree contains a non-regular file")
		}
		relative, err := filepath.Rel(absoluteRoot, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if !allowed[relative] || seen[relative] {
			return errors.New("parent baseline tree has missing or extra artifacts")
		}
		seen[relative] = true
		return nil
	})
	if err != nil || len(seen) != len(allowed) {
		return errors.New("parent baseline tree coverage mismatch")
	}
	return nil
}

func verifyParentBaseline(a controllerAuthorityV4, files parentBaselineFiles, controllerSourceRoot, runtimeSourceRoot, fixturesPath, benchmarkPath, scorerPath string, bench benchmarkConfig, scanner *privacyScanner) error {
	expectedRuns := bench.FixtureCount + 2
	if err := validateParentBaselineBinding(a.ParentBaseline, expectedRuns); err != nil {
		return err
	}
	if files == (parentBaselineFiles{}) || pathsOverlap(files.Root, controllerSourceRoot) || pathsOverlap(files.Root, runtimeSourceRoot) || pathsOverlap(files.ValidatorStore, controllerSourceRoot) || pathsOverlap(files.ValidatorStore, runtimeSourceRoot) {
		return errors.New("parent baseline paths overlap governed source roots")
	}
	if a.Expected.ParentBaselineAuthoritySHA256 != a.ParentBaseline.AuthoritySHA256 ||
		a.Expected.ParentBaselineOutputSHA256 != a.ParentBaseline.OutputSHA256 ||
		a.Expected.ParentBaselineOrderedRunSHA256 != a.ParentBaseline.OrderedRunSHA256 ||
		a.Expected.ParentBaselineRunCount != a.ParentBaseline.RunCount ||
		a.Expected.ParentBaselineHistorySHA256 != a.ParentBaseline.HistorySHA256 ||
		a.Expected.ParentBaselineHistoryTailSHA256 != a.ParentBaseline.HistoryTailRecordHash ||
		a.Expected.ParentBaselinePrivacySHA256 != a.ParentBaseline.PrivacyManifestSHA256 ||
		a.Expected.ParentBaselineValidatorHash != a.ParentBaseline.ValidatorRecordHash ||
		a.Expected.ParentBaselineValidatorPayload != a.ParentBaseline.ValidatorPayloadSHA256 {
		return errors.New("candidate provenance does not bind its parent baseline")
	}
	authorityBytes, err := readBoundedRegularFileInside(files.Root, files.Authority, 4<<20)
	if err != nil || shaString(string(authorityBytes)) != a.ParentBaseline.AuthoritySHA256 {
		return errors.New("parent baseline authority hash mismatch")
	}
	outputBytes, err := readBoundedRegularFileInside(files.Root, files.Output, 4<<20)
	if err != nil || shaString(string(outputBytes)) != a.ParentBaseline.OutputSHA256 {
		return errors.New("parent baseline output hash mismatch")
	}
	if err := scanner.Scan("parent_authority", "parent/controller-authority.v4.json", authorityBytes); err != nil {
		return err
	}
	if err := scanner.Scan("parent_output", "parent/controller-output.v4.json", outputBytes); err != nil {
		return err
	}
	var parent controllerAuthorityV4
	if err := decodeStrictBytes(authorityBytes, &parent); err != nil {
		return errors.New("invalid parent baseline authority")
	}
	if err := verifyParentFrozenEquality(a, parent); err != nil {
		return err
	}
	var output controllerOutput
	if err := decodeStrictBytes(outputBytes, &output); err != nil {
		return errors.New("invalid parent baseline output")
	}
	if output.Schema != "structural-retrieval-controller-output.v4" || output.Admitted || output.Admission != "diagnostic_unadmitted" || output.Failure != nil ||
		len(output.Runs) != expectedRuns || output.OrderedRunManifestSHA256 != canonicalSHA256(output.Runs) || output.OrderedRunManifestSHA256 != a.ParentBaseline.OrderedRunSHA256 ||
		output.HistorySHA256 != a.ParentBaseline.HistorySHA256 || output.PrivacyManifestSHA256 != a.ParentBaseline.PrivacyManifestSHA256 {
		return errors.New("parent baseline final output binding mismatch")
	}
	historyBytes, err := parentRelativeArtifact(files.Root, output.HistoryPath, a.ScanCaps.SingleDurableArtifactBytesMax)
	if err != nil || shaString(string(historyBytes)) != output.HistorySHA256 {
		return errors.New("parent baseline history hash mismatch")
	}
	if err := scanner.Scan("parent_history", output.HistoryPath, historyBytes); err != nil {
		return err
	}
	history, err := decodeHistoryBytes(historyBytes)
	if err != nil || len(history) != expectedRuns {
		return errors.New("invalid parent baseline history")
	}
	if err := verifyHistorySchema(history, historyV4Schema); err != nil || history[len(history)-1].RecordHash != a.ParentBaseline.HistoryTailRecordHash {
		return errors.New("parent baseline history chain mismatch")
	}
	fixtures, fixtureHash, err := loadFixtures(fixturesPath)
	if err != nil || fixtureHash != parent.Expected.FixtureSHA256 || len(fixtures) != bench.FixtureCount {
		return errors.New("parent baseline fixture binding mismatch")
	}
	known, err := verifyParentRunArtifacts(files.Root, output, history, parent, fixtures, bench, scanner)
	if err != nil {
		return err
	}
	known[output.HistoryPath] = historyBytes
	for index := range output.Runs {
		runtimePath := fmt.Sprintf("runtime/%02d.stderr", index+1)
		data, err := parentRelativeArtifact(files.Root, runtimePath, parent.ScanCaps.SingleDurableArtifactBytesMax)
		if err != nil {
			return errors.New("parent runtime diagnostic artifact mismatch")
		}
		known[runtimePath] = data
	}
	if err := verifyParentPrivacyManifest(files.Root, output, parent, authorityBytes, known, controllerSourceRoot, fixturesPath, benchmarkPath, scorerPath, scanner); err != nil {
		return err
	}
	if err := verifyParentRootCoverage(files, output); err != nil {
		return err
	}
	acceptance := baselineAcceptance{
		Schema: "structural-retrieval-baseline-acceptance.v1", Verdict: "accepted",
		AuthoritySHA256: a.ParentBaseline.AuthoritySHA256, OutputSHA256: a.ParentBaseline.OutputSHA256,
		OrderedRunManifestSHA256: a.ParentBaseline.OrderedRunSHA256, RunCount: a.ParentBaseline.RunCount,
		HistorySHA256: a.ParentBaseline.HistorySHA256, HistoryTailRecordHash: a.ParentBaseline.HistoryTailRecordHash,
		PrivacyManifestSHA256: a.ParentBaseline.PrivacyManifestSHA256,
	}
	acceptance.ValidatedSourceMainSHA256, err = fileSHA256(scorerPath)
	if err != nil {
		return err
	}
	protocolTestPath := filepath.Join(filepath.Dir(scorerPath), "main_test.go")
	acceptance.ValidatedSourceTestSHA256, err = fileSHA256(protocolTestPath)
	if err != nil {
		return err
	}
	return verifyParentValidatorRecord(files, *a.ParentBaseline, acceptance, scorerPath, protocolTestPath)
}

func verifyControllerAuthorityV4(a controllerAuthorityV4, authorityBytes []byte, runtimePath, controllerPath, controllerSourceRoot, runtimeSourceRoot, bundlePath, privacyPolicyPath, fixturesPath, benchmarkPath, scorerPath string, bench benchmarkConfig, policy privacyPolicy, scanner *privacyScanner, parentFiles parentBaselineFiles) error {
	timeout := time.Duration(protocolV7ScanCaps().PreflightWallTimeMillis) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := withProtocolCommandContext(ctx, func() error {
		return verifyControllerAuthorityV4Preflight(a, authorityBytes, runtimePath, controllerPath, controllerSourceRoot, runtimeSourceRoot, bundlePath, privacyPolicyPath, fixturesPath, benchmarkPath, scorerPath, bench, policy, scanner, parentFiles)
	})
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return errors.New("protocol-v7 preflight deadline exceeded")
	}
	return err
}

func verifyControllerAuthorityV4Preflight(a controllerAuthorityV4, authorityBytes []byte, runtimePath, controllerPath, controllerSourceRoot, runtimeSourceRoot, bundlePath, privacyPolicyPath, fixturesPath, benchmarkPath, scorerPath string, bench benchmarkConfig, policy privacyPolicy, scanner *privacyScanner, parentFiles parentBaselineFiles) error {
	if a.Schema != controllerAuthorityV4Schema || (a.Mode != "baseline" && a.Mode != "candidate") || a.ScanCaps != protocolV7ScanCaps() {
		return errors.New("invalid controller authority v4")
	}
	if policy.SHA256 != a.PrivacyPolicySHA256 || len(policy.DenyTerms) != a.PrivacyTermCount || a.Expected.PrivacyPolicySHA256 != a.PrivacyPolicySHA256 || a.Expected.PrivacyTermCount != a.PrivacyTermCount {
		return errors.New("privacy policy authority mismatch")
	}
	if scanner == nil || scanner.detector.DefinitionSHA256 != a.PrivacyDetectorSHA256 || a.Expected.PrivacyDetectorSHA256 != a.PrivacyDetectorSHA256 {
		return errors.New("privacy detector authority mismatch")
	}
	if canonicalSHA256(a.ScanCaps) != a.Expected.ScanCapsSHA256 || a.SourceBundleHeadsSHA256 != a.Expected.SourceBundleHeadsSHA256 {
		return errors.New("protocol-v7 cap or bundle-head authority mismatch")
	}
	if !validSHA256(a.ProtocolTestSHA256) || allZero(a.ProtocolTestSHA256) || a.Expected.ProtocolTestSHA256 != a.ProtocolTestSHA256 {
		return errors.New("protocol-v7 test-source authority mismatch")
	}
	protocolTestPath := filepath.Join(filepath.Dir(scorerPath), "main_test.go")
	goHash, err := goExecutableSHA256()
	if err != nil || goHash != a.GoExecutableSHA256 || goHash != a.Expected.GoExecutableSHA256 {
		return errors.New("Go executable authority mismatch")
	}
	for _, exact := range []struct{ path, relative string }{
		{fixturesPath, "research/eval/structural-retrieval/fixtures.dev.v2.jsonl"},
		{benchmarkPath, "research/eval/structural-retrieval/benchmark.v2.json"},
		{scorerPath, "cli/tools/structural-search-eval-v2/main.go"},
		{protocolTestPath, "cli/tools/structural-search-eval-v2/main_test.go"},
	} {
		if !registeredPathInside(controllerSourceRoot, exact.path, exact.relative) {
			return errors.New("protocol-v7 frozen input is symlinked or at the wrong B path")
		}
	}
	for _, frozen := range []struct {
		role, relative, path string
	}{
		{"frozen_fixture", "research/eval/structural-retrieval/fixtures.dev.v2.jsonl", fixturesPath},
		{"frozen_benchmark", "research/eval/structural-retrieval/benchmark.v2.json", benchmarkPath},
		{"frozen_scorer", "cli/tools/structural-search-eval-v2/main.go", scorerPath},
		{"frozen_protocol_test", "cli/tools/structural-search-eval-v2/main_test.go", protocolTestPath},
	} {
		data, err := readBoundedRegularFileInside(controllerSourceRoot, frozen.path, a.ScanCaps.SingleDurableArtifactBytesMax)
		if err != nil {
			return err
		}
		if frozen.role == "frozen_protocol_test" && shaString(string(data)) != a.ProtocolTestSHA256 {
			return errors.New("protocol-v7 test-source file mismatch")
		}
		if err := scanner.Scan(frozen.role, frozen.relative, data); err != nil {
			return err
		}
	}
	policyInfo, err := os.Lstat(privacyPolicyPath)
	if err != nil || !policyInfo.Mode().IsRegular() || policyInfo.Mode()&os.ModeSymlink != 0 || policyInfo.Mode().Perm() != 0o600 || pathWithin(controllerSourceRoot, privacyPolicyPath) || pathWithin(runtimeSourceRoot, privacyPolicyPath) {
		return errors.New("privacy policy must be an external mode-0600 regular file")
	}
	if err := rejectReplaceRefs(controllerSourceRoot); err != nil {
		return err
	}
	if a.Mode == "candidate" {
		if err := rejectReplaceRefs(runtimeSourceRoot); err != nil {
			return err
		}
	}
	var exactAuthority controllerAuthorityV4
	if len(authorityBytes) == 0 || decodeStrictBytes(authorityBytes, &exactAuthority) != nil || !reflect.DeepEqual(exactAuthority, a) {
		return errors.New("exact protocol-v7 authority bytes mismatch")
	}
	if err := scanner.Scan("authority", "controller-authority.v4.json", authorityBytes); err != nil {
		return err
	}
	beforeController, err := goModVerifySHA256(controllerSourceRoot)
	if err != nil {
		return err
	}
	beforeRuntime, err := goModVerifySHA256(runtimeSourceRoot)
	if err != nil {
		return err
	}
	if beforeController != a.GoModVerifySHA256 || beforeRuntime != a.GoModVerifySHA256 || a.Expected.GoModVerifySHA256 != a.GoModVerifySHA256 {
		return errors.New("go mod verify authority mismatch")
	}
	v3 := controllerAuthorityV3{
		Schema: controllerAuthorityV3Schema, Mode: a.Mode, Expected: a.Expected,
		ControllerSourceCapsule: a.ControllerSourceCapsule, RuntimeSourceCapsule: a.RuntimeSourceCapsule,
		CandidateDelta: a.CandidateDelta, BuildReplay: a.BuildReplay, BudgetLimits: a.BudgetLimits,
		ActionEnvelope: a.ActionEnvelope, CanonicalRowBytesDefinition: a.CanonicalRowBytesDefinition,
		ContextThresholdAuthorityRecord: a.ContextThresholdAuthorityRecord,
	}
	if err := verifyControllerAuthorityV3(v3, runtimePath, controllerPath, controllerSourceRoot, runtimeSourceRoot, bundlePath, fixturesPath, benchmarkPath, scorerPath, bench); err != nil {
		return err
	}
	if a.Mode == "baseline" {
		if a.ParentBaseline != nil || parentFiles != (parentBaselineFiles{}) {
			return errors.New("baseline authority has a parent")
		}
	} else if a.ParentBaseline == nil {
		return errors.New("candidate authority is missing its accepted baseline parent")
	} else {
		if pathsOverlap(parentFiles.Root, bundlePath) || pathsOverlap(parentFiles.Root, privacyPolicyPath) || pathsOverlap(parentFiles.ValidatorStore, bundlePath) || pathsOverlap(parentFiles.ValidatorStore, privacyPolicyPath) {
			return errors.New("parent baseline paths overlap candidate governed inputs")
		}
		if err := verifyParentBaseline(a, parentFiles, controllerSourceRoot, runtimeSourceRoot, fixturesPath, benchmarkPath, scorerPath, bench, scanner); err != nil {
			return err
		}
	}
	if _, _, err := verifyAndScanSourceBundle(runtimeSourceRoot, bundlePath, a.Expected.Commit, a.Expected.BundleSHA256, a.SourceBundleHeadsSHA256, scanner, a.ScanCaps); err != nil {
		return err
	}
	afterController, err := goModVerifySHA256(controllerSourceRoot)
	if err != nil {
		return err
	}
	afterRuntime, err := goModVerifySHA256(runtimeSourceRoot)
	if err != nil {
		return err
	}
	if afterController != a.GoModVerifySHA256 || afterRuntime != a.GoModVerifySHA256 {
		return errors.New("module cache changed during protocol-v7 rebuild")
	}
	return nil
}

func verifyCleanSourceCapsuleClosure(sourceRoot string, want sourceCapsule) error {
	status, err := gitCommandBytes(sourceRoot, "status", "--porcelain=v1", "--untracked-files=all")
	if err != nil {
		return err
	}
	if len(status) != 0 {
		return errors.New("source root is dirty")
	}
	got, err := captureSourceCapsule(sourceRoot)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(got, want) {
		return errors.New("source capsule does not match exact clean build closure")
	}
	return validateCleanSourceCapsule(want)
}

func buildFrozenRuntime(sourceRoot, outputPath string) error {
	goPath, err := exec.LookPath("go")
	if err != nil {
		return err
	}
	home, moduleCache, buildCache, err := hermeticGoPaths()
	if err != nil {
		return err
	}
	command := protocolCommand(goPath, "build", "-mod=readonly", "-trimpath", "-buildvcs=false", "-o", outputPath, "./tools/structural-search-eval-v2")
	command.Dir = filepath.Join(sourceRoot, "cli")
	command.Env = hermeticGoEnvironment(home, moduleCache, buildCache)
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("go build: %w: %s", err, output)
	} else if len(output) != 0 {
		return errors.New("hermetic go build wrote output")
	}
	return nil
}

func verifyBaselineBundle(repoRoot, bundlePath, wantBundleHash, commit string) error {
	got, err := fileSHA256(bundlePath)
	if err != nil || got != wantBundleHash {
		return errors.New("baseline bundle SHA-256 mismatch")
	}
	if _, err := gitCommandBytes(repoRoot, "bundle", "verify", bundlePath); err != nil {
		return err
	}
	fresh, err := os.MkdirTemp("", "c3-v6-baseline-bundle-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(fresh)
	if _, err := gitCommandBytes(fresh, "init", "--bare", "repo.git"); err != nil {
		return err
	}
	freshGit := filepath.Join(fresh, "repo.git")
	if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "bundle", "unbundle", bundlePath); err != nil {
		return err
	}
	if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "cat-file", "-e", commit+"^{commit}"); err != nil {
		return errors.New("baseline bundle is missing its commit")
	}
	return nil
}

func rejectReplaceRefs(repoRoot string) error {
	refs, err := gitCommandBytes(repoRoot, "for-each-ref", "--format=%(refname)", "refs/replace/")
	if err != nil {
		return err
	}
	if len(refs) != 0 {
		return errors.New("repository contains replacement refs")
	}
	return nil
}

func verifyAndScanSourceBundle(repoRoot, bundlePath, expectedCommit, wantBundleHash, wantHeadsHash string, scanner *privacyScanner, caps privacyScanCaps) (int, int64, error) {
	if scanner == nil || caps != protocolV7ScanCaps() {
		return 0, 0, errors.New("invalid protocol-v7 scan authority")
	}
	info, err := os.Stat(bundlePath)
	if err != nil {
		return 0, 0, err
	}
	if !info.Mode().IsRegular() || info.Size() > caps.BundleFileBytesMax {
		return 0, 0, errors.New("source bundle exceeds protocol-v7 cap")
	}
	gotBundleHash, err := fileSHA256(bundlePath)
	if err != nil || gotBundleHash != wantBundleHash {
		return 0, 0, errors.New("source bundle SHA-256 mismatch")
	}
	headsHash, headsBytes, err := gitSHA256(repoRoot, "bundle", "list-heads", bundlePath)
	if err != nil || headsHash != wantHeadsHash {
		return 0, 0, errors.New("source bundle heads mismatch")
	}
	headLines := strings.Split(strings.TrimSuffix(string(headsBytes), "\n"), "\n")
	if len(headLines) != 1 {
		return 0, 0, errors.New("source bundle must have exactly one head")
	}
	head := strings.Fields(headLines[0])
	if len(head) != 2 || head[0] != expectedCommit || !strings.HasPrefix(head[1], "refs/c3-eval/commit-pool/") {
		return 0, 0, errors.New("source bundle head mismatch")
	}
	fresh, err := os.MkdirTemp("", "c3-v7-source-bundle-")
	if err != nil {
		return 0, 0, err
	}
	defer os.RemoveAll(fresh)
	if _, err := gitCommandBytes(fresh, "init", "--bare", "repo.git"); err != nil {
		return 0, 0, err
	}
	freshGit := filepath.Join(fresh, "repo.git")
	if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "bundle", "verify", bundlePath); err != nil {
		return 0, 0, err
	}
	if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "bundle", "unbundle", bundlePath); err != nil {
		return 0, 0, err
	}
	if _, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "update-ref", "refs/c3-eval/verified-source", expectedCommit); err != nil {
		return 0, 0, err
	}
	if err := rejectReplaceRefs(freshGit); err != nil {
		return 0, 0, err
	}
	objectLines, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "cat-file", "--batch-all-objects", "--batch-check=%(objectname) %(objecttype) %(objectsize)")
	if err != nil {
		return 0, 0, err
	}
	reachableBytes, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "rev-list", "--objects", "--no-object-names", expectedCommit)
	if err != nil {
		return 0, 0, err
	}
	reachable := map[string]bool{}
	for _, object := range strings.Fields(string(reachableBytes)) {
		reachable[object] = true
	}
	objectCount := 0
	var totalBytes int64
	seen := map[string]bool{}
	for _, line := range strings.Split(strings.TrimSuffix(string(objectLines), "\n"), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 3 || seen[fields[0]] || !reachable[fields[0]] {
			return 0, 0, errors.New("source bundle contains duplicate or unreachable objects")
		}
		size, err := strconv.ParseInt(fields[2], 10, 64)
		if err != nil || size < 0 {
			return 0, 0, errors.New("invalid source object size")
		}
		switch fields[1] {
		case "blob":
			if size > caps.SingleBlobBytesMax {
				return 0, 0, errors.New("source blob exceeds protocol-v7 cap")
			}
		case "commit", "tag":
			if size > caps.SingleCommitOrTagBytesMax {
				return 0, 0, errors.New("source commit or tag exceeds protocol-v7 cap")
			}
		case "tree":
		default:
			return 0, 0, errors.New("unsupported source object type")
		}
		objectCount++
		totalBytes += size
		if objectCount > caps.SourceObjectCountMax || totalBytes > caps.SourceObjectUncompressedBytesMax {
			return 0, 0, errors.New("source objects exceed protocol-v7 aggregate cap")
		}
		content, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "cat-file", "-p", fields[0])
		if err != nil {
			return 0, 0, err
		}
		if fields[1] != "tree" && int64(len(content)) != size {
			return 0, 0, errors.New("source object size changed during scan")
		}
		if err := scanner.Scan("source_object_"+fields[1], "objects/"+fields[0], content); err != nil {
			return 0, 0, err
		}
		seen[fields[0]] = true
	}
	pathArtifacts, pathContents, err := fullReachablePathArtifacts(freshGit, expectedCommit, caps)
	if err != nil {
		return 0, 0, err
	}
	paths := make([]string, 0, len(pathArtifacts))
	for path := range pathArtifacts {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := scanner.Scan("source_tree_path", path, pathContents[path]); err != nil {
			return 0, 0, err
		}
	}
	if len(seen) != len(reachable) {
		return 0, 0, errors.New("source bundle reachability set mismatch")
	}
	fsck, err := gitCommandBytes(fresh, "--git-dir="+freshGit, "fsck", "--full", "--strict", "--unreachable", "--no-reflogs")
	if err != nil || len(fsck) != 0 {
		return 0, 0, errors.New("source bundle failed standalone fsck or has unreachable objects")
	}
	scanner.sourceObjectCount = objectCount
	scanner.sourceObjectBytes = totalBytes
	return objectCount, totalBytes, nil
}

func verifyPortableTreePrivacy(repoRoot, commit string) error {
	paths, err := gitCommandBytes(repoRoot, "ls-tree", "-r", "--name-only", "-z", commit)
	if err != nil {
		return err
	}
	for _, rawPath := range bytes.Split(paths, []byte{0}) {
		if len(rawPath) == 0 {
			continue
		}
		path := filepath.ToSlash(string(rawPath))
		lowerPath := strings.ToLower(path)
		if filepath.IsAbs(path) || strings.HasPrefix(lowerPath, ".okra/") || strings.Contains(lowerPath, "private") {
			return fmt.Errorf("non-portable or private path in source bundle: %s", path)
		}
		content, err := gitCommandBytes(repoRoot, "show", commit+":"+path)
		if err != nil {
			return err
		}
		lower := strings.ToLower(string(content))
		for _, forbidden := range []string{"/" + "home/", "/" + "users/", "-----begin " + "private key", "gh" + "p_", "sk" + "-ant-"} {
			if strings.Contains(lower, forbidden) {
				return fmt.Errorf("non-portable or credential-like content in %s", path)
			}
		}
	}
	return nil
}

func pathWithin(root, path string) bool {
	absoluteRoot, err := resolvedAbsolutePath(root)
	if err != nil {
		return false
	}
	absolutePath, err := resolvedAbsolutePath(path)
	if err != nil {
		return false
	}
	relative, err := filepath.Rel(absoluteRoot, absolutePath)
	return err == nil && relative != "." && relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator))
}

func verifyPortableAuthorityPrivacy(authority controllerAuthorityV3) error {
	data, err := json.Marshal(authority)
	if err != nil {
		return err
	}
	lower := strings.ToLower(string(data))
	for _, forbidden := range []string{"/" + "home/", "/" + "users/", "-----begin " + "private key", "gh" + "p_", "sk" + "-ant-"} {
		if strings.Contains(lower, forbidden) {
			return errors.New("controller authority contains a non-portable path or credential-like value")
		}
	}
	return nil
}
func shaString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func validateCandidateSource(base, candidate string) error {
	baseFile, err := parser.ParseFile(token.NewFileSet(), "base.go", base, parser.AllErrors)
	if err != nil {
		return err
	}
	candidateFile, err := parser.ParseFile(token.NewFileSet(), "candidate.go", candidate, parser.AllErrors)
	if err != nil {
		return err
	}
	baseImports := importPaths(baseFile)
	candidateImports := importPaths(candidateFile)
	if !reflect.DeepEqual(baseImports, candidateImports) {
		return errors.New("candidate import set changed")
	}
	for _, bad := range []string{"func init(", "os.Getenv", "os.ReadFile", "net.Dial", "exec.Command", "fmt.Print", "expected answer", "fixture"} {
		if strings.Contains(candidate, bad) {
			return fmt.Errorf("candidate contains forbidden construct %q", bad)
		}
	}
	for _, decl := range candidateFile.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Recv == nil && fn.Name.Name == "init" {
			return errors.New("candidate adds init")
		}
	}
	if !strings.Contains(candidate, "fts5(title, goal, parent_id)") || !strings.Contains(candidate, "migrateParentKeywordFTS()") || strings.Contains(candidate, "parent_title") {
		return errors.New("candidate does not match registered parent-keyword SQL behavior")
	}
	return nil
}
func importPaths(f *ast.File) []string {
	var out []string
	for _, in := range f.Imports {
		value, err := strconv.Unquote(in.Path.Value)
		if err == nil {
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}

type familyResults struct {
	DefaultIDNaturalLanguageLift float64
	ExplicitIDLift               float64
	SemanticIDLift               float64
}

func familyDecision(r familyResults) string {
	if r.DefaultIDNaturalLanguageLift <= 0 && (r.ExplicitIDLift > 0 || r.SemanticIDLift > 0) {
		return decisionDiscard
	}
	if r.DefaultIDNaturalLanguageLift > 0 {
		return decisionKeep
	}
	return decisionDiscard
}

type transferRecord struct {
	SourceSemanticMode      string
	DestinationSemanticMode string
	NumericComparison       bool
	Status                  string
}

func validateTransfer(r transferRecord) error {
	if r.SourceSemanticMode != semanticDisabled || r.DestinationSemanticMode != semanticDefault || r.NumericComparison || r.Status != "candidate" {
		return errors.New("invalid cross-DKR transfer")
	}
	return nil
}

type holdoutPolicy struct {
	ContentVisibleToWriter    bool
	ContentVisibleToCandidate bool
	AllowedPurposes           []string
	MaxScoreReads             int
}

func validateHoldoutPolicy(p holdoutPolicy) error {
	if p.ContentVisibleToWriter || p.ContentVisibleToCandidate || p.MaxScoreReads != 1 || !reflect.DeepEqual(p.AllowedPurposes, []string{"post-selection-confirmation", "audit"}) {
		return errors.New("invalid holdout policy")
	}
	return nil
}

type legacySchemaState struct {
	FTSColumns        []string
	LogicalRowsSHA256 string
}

func simulateRegisteredMigration(s legacySchemaState) (legacySchemaState, error) {
	if !validSHA256(s.LogicalRowsSHA256) {
		return s, errors.New("invalid logical rows hash")
	}
	if reflect.DeepEqual(s.FTSColumns, []string{"title", "goal", "parent_id"}) {
		return s, nil
	}
	if !reflect.DeepEqual(s.FTSColumns, []string{"title", "goal"}) {
		return s, errors.New("unknown FTS schema")
	}
	s.FTSColumns = []string{"title", "goal", "parent_id"}
	return s, nil
}

type confinementSpec struct {
	RuntimePath    string
	Root           string
	UnitName       string
	SystemdRunPath string
	BwrapPath      string
	PrlimitPath    string
	Environment    map[string]string
	ReadOnlyBinds  []string
	Command        []string
	BudgetLimits   resourceBudget
}

var confinementUnitSequence atomic.Uint64

func newConfinementSpec(runtimePath, root string, limits resourceBudget) (confinementSpec, error) {
	absRuntime, err := filepath.Abs(runtimePath)
	if err != nil {
		return confinementSpec{}, err
	}
	runtimeMaxSeconds, err := runtimeMaxSeconds(limits)
	if err != nil {
		return confinementSpec{}, err
	}
	info, err := os.Stat(absRuntime)
	if err != nil || info.IsDir() || info.Mode()&0o111 == 0 {
		return confinementSpec{}, fmt.Errorf("runtime is not executable: %w", err)
	}
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return confinementSpec{}, err
	}
	bwrap, err := exec.LookPath("bwrap")
	if err != nil {
		return confinementSpec{}, errors.New("bubblewrap unavailable")
	}
	prlimit, err := exec.LookPath("prlimit")
	if err != nil {
		return confinementSpec{}, errors.New("prlimit unavailable")
	}
	systemdRun, err := exec.LookPath("systemd-run")
	if err != nil {
		return confinementSpec{}, errors.New("systemd-run unavailable")
	}
	unitName := fmt.Sprintf("c3-eval-v2-%d-%d", os.Getpid(), confinementUnitSequence.Add(1))
	spec := confinementSpec{RuntimePath: absRuntime, Root: absRoot, UnitName: unitName, SystemdRunPath: systemdRun, BwrapPath: bwrap, PrlimitPath: prlimit, Environment: map[string]string{"HOME": "/home", "TZ": "UTC", "LANG": "C.UTF-8"}, ReadOnlyBinds: []string{"/usr"}, BudgetLimits: limits}
	// The trusted outer shell leaves the cgroup alive briefly after the arm exits.
	// The controller reads kernel-owned cpu.stat, memory.peak, and pids.peak in
	// that window; ProcessState for the systemd-run client is never used.
	spec.Command = []string{systemdRun, "--user", "--wait", "--quiet", "--pipe", "--unit", unitName, "-p", "TasksMax=16", "-p", "MemoryMax=536870912", "-p", fmt.Sprintf("RuntimeMaxSec=%ds", runtimeMaxSeconds), "--", "/bin/sh", "-c", `"$@"; rc=$?; /bin/sleep 0.25; exit "$rc"`, "sh", prlimit, "--nofile=64:64", "--cpu=10:10", "--", bwrap, "--unshare-all", "--unshare-net", "--new-session", "--die-with-parent", "--ro-bind", "/usr", "/usr", "--symlink", "usr/bin", "/bin", "--symlink", "usr/lib", "/lib", "--symlink", "usr/lib64", "/lib64", "--ro-bind", absRuntime, "/runtime", "--bind", absRoot, "/work", "--proc", "/proc", "--dev", "/dev", "--tmpfs", "/tmp", "--dir", "/home", "--clearenv", "--setenv", "HOME", "/home", "--setenv", "TZ", "UTC", "--setenv", "LANG", "C.UTF-8", "/runtime"}
	if err := validateConfinementSpec(spec); err != nil {
		return confinementSpec{}, err
	}
	return spec, nil
}
func validateConfinementSpec(s confinementSpec) error {
	if s.RuntimePath == "" || s.Root == "" || s.UnitName == "" || s.SystemdRunPath == "" || s.BwrapPath == "" || s.PrlimitPath == "" || !filepath.IsAbs(s.RuntimePath) || !filepath.IsAbs(s.Root) {
		return errors.New("incomplete confinement spec")
	}
	runtimeMaxSeconds, err := runtimeMaxSeconds(s.BudgetLimits)
	if err != nil {
		return err
	}
	joined := strings.Join(s.Command, "\x00")
	for _, want := range []string{"TasksMax=16", "MemoryMax=536870912", fmt.Sprintf("RuntimeMaxSec=%ds", runtimeMaxSeconds), "--unshare-all", "--unshare-net", "--clearenv", "--nofile=64:64", "--cpu=10:10", "--ro-bind\x00" + s.RuntimePath + "\x00/runtime", "--bind\x00" + s.Root + "\x00/work"} {
		if !strings.Contains(joined, want) {
			return fmt.Errorf("confinement command omits %s", want)
		}
	}
	if s.Environment["HOME"] != "/home" || s.Environment["TZ"] != "UTC" {
		return errors.New("environment not frozen")
	}
	for _, bind := range s.ReadOnlyBinds {
		if strings.HasPrefix(filepath.Clean(bind), filepath.Clean(s.Root)) {
			return errors.New("controller root exposed read-only")
		}
	}
	return nil
}
func proveConfinementBackend(ctx context.Context, root string, limits resourceBudget) error {
	if err := os.MkdirAll(root, 0o700); err != nil {
		return err
	}
	secret := filepath.Join(root, "oracle-secret")
	if err := os.WriteFile(secret, []byte("secret"), 0o600); err != nil {
		return err
	}
	armRoot := filepath.Join(root, "arm")
	if err := os.Mkdir(armRoot, 0o700); err != nil {
		return err
	}
	spec, err := newConfinementSpec("/bin/sh", armRoot, limits)
	if err != nil {
		return err
	}
	repositoryPath, err := os.Getwd()
	if err != nil {
		return err
	}
	script := `test ! -e "$1" && test ! -e "$2" && ! /usr/bin/curl -fsS --connect-timeout 1 http://1.1.1.1 >/dev/null 2>&1`
	args := append([]string(nil), spec.Command...)
	args = append(args, "-c", script, "sh", secret, repositoryPath)
	cmdline := exec.CommandContext(ctx, args[0], args[1:]...)
	cmdline.Env = systemdTransportEnv()
	out, err := cmdline.CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemd cgroup plus bubblewrap and prlimit capability proof failed: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func systemdTransportEnv() []string {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	if runtimeDir == "" {
		runtimeDir = "/run/user/" + strconv.Itoa(os.Getuid())
	}
	bus := os.Getenv("DBUS_SESSION_BUS_ADDRESS")
	if bus == "" {
		bus = "unix:path=" + runtimeDir + "/bus"
	}
	return []string{"XDG_RUNTIME_DIR=" + runtimeDir, "DBUS_SESSION_BUS_ADDRESS=" + bus}
}

type limitedBuffer struct {
	Buffer bytes.Buffer
	Limit  int
}

func (w *limitedBuffer) Write(p []byte) (int, error) {
	if w.Limit <= 0 || w.Buffer.Len()+len(p) > w.Limit {
		return 0, errors.New("confined arm output limit exceeded")
	}
	return w.Buffer.Write(p)
}

type completionStatus struct {
	RC int
}

func (s completionStatus) validate() error {
	if s.RC < 0 || s.RC > 255 {
		return errors.New("completion rc is out of range")
	}
	return nil
}

type completionEndpoint struct {
	Dir       string
	Path      string
	Nonce     string
	OwnerUID  int
	CreatedAt time.Time
}

type completionMarker struct {
	Status  completionStatus
	ModTime time.Time
}

func newCompletionEndpoint(parent, armRoot string, now func() time.Time) (completionEndpoint, error) {
	if now == nil {
		return completionEndpoint{}, errors.New("completion clock is required")
	}
	absParent, err := filepath.Abs(parent)
	if err != nil {
		return completionEndpoint{}, err
	}
	absArmRoot, err := filepath.Abs(armRoot)
	if err != nil {
		return completionEndpoint{}, err
	}
	if pathsOverlap(absParent, absArmRoot) && absParent != filepath.Dir(absArmRoot) {
		return completionEndpoint{}, errors.New("completion parent overlaps confined arm root")
	}
	nonceBytes := make([]byte, 32)
	if _, err := rand.Read(nonceBytes); err != nil {
		return completionEndpoint{}, fmt.Errorf("create completion nonce: %w", err)
	}
	dir, err := os.MkdirTemp(absParent, ".c3-completion-")
	if err != nil {
		return completionEndpoint{}, fmt.Errorf("create private completion directory: %w", err)
	}
	endpoint := completionEndpoint{
		Dir:       dir,
		Path:      filepath.Join(dir, "complete"),
		Nonce:     hex.EncodeToString(nonceBytes),
		OwnerUID:  os.Geteuid(),
		CreatedAt: now(),
	}
	if pathsOverlap(endpoint.Dir, absArmRoot) {
		_ = os.RemoveAll(endpoint.Dir)
		return completionEndpoint{}, errors.New("completion directory is visible inside arm root")
	}
	info, err := os.Stat(endpoint.Dir)
	if err != nil || !info.IsDir() || info.Mode().Perm() != 0o700 {
		_ = os.RemoveAll(endpoint.Dir)
		return completionEndpoint{}, errors.New("completion directory is not private mode 0700")
	}
	return endpoint, nil
}

func cleanupCompletionEndpoint(endpoint completionEndpoint) error {
	if endpoint.Dir == "" {
		return nil
	}
	return os.RemoveAll(endpoint.Dir)
}

func guardCompletionEndpointCleanup(endpoint completionEndpoint, cleanup func(completionEndpoint) error) func() {
	return func() { _ = cleanup(endpoint) }
}

func encodeCompletionMarker(nonce string, status completionStatus) string {
	return fmt.Sprintf("nonce=%s\nrc=%d\n", nonce, status.RC)
}

func publishCompletionMarker(endpoint completionEndpoint, status completionStatus) error {
	if err := status.validate(); err != nil {
		return err
	}
	if _, err := os.Lstat(endpoint.Path); err == nil {
		return errors.New("stale completion marker already exists")
	} else if !os.IsNotExist(err) {
		return err
	}
	tempPath := filepath.Join(endpoint.Dir, ".complete."+endpoint.Nonce+".tmp")
	file, err := os.OpenFile(tempPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	removeTemp := true
	defer func() {
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()
	data := []byte(encodeCompletionMarker(endpoint.Nonce, status))
	if _, err := file.Write(data); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Sync(); err != nil {
		_ = file.Close()
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tempPath, 0o600); err != nil {
		return err
	}
	if err := os.Rename(tempPath, endpoint.Path); err != nil {
		return err
	}
	removeTemp = false
	return nil
}

func readCompletionMarker(endpoint completionEndpoint) (completionMarker, error) {
	info, err := os.Lstat(endpoint.Path)
	if err != nil {
		return completionMarker{}, err
	}
	if !info.Mode().IsRegular() {
		return completionMarker{}, errors.New("completion marker is not a regular file")
	}
	if info.Mode().Perm() != 0o600 {
		return completionMarker{}, errors.New("completion marker mode is not 0600")
	}
	if info.Size() <= 0 || info.Size() > completionMarkerMaxBytes {
		return completionMarker{}, errors.New("completion marker size is invalid")
	}
	if info.ModTime().Add(time.Second).Before(endpoint.CreatedAt) {
		return completionMarker{}, errors.New("completion marker is stale")
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || int(stat.Uid) != endpoint.OwnerUID {
		return completionMarker{}, errors.New("completion marker owner is invalid")
	}
	file, err := os.Open(endpoint.Path)
	if err != nil {
		return completionMarker{}, err
	}
	defer file.Close()
	openedInfo, err := file.Stat()
	if err != nil {
		return completionMarker{}, err
	}
	openedStat, ok := openedInfo.Sys().(*syscall.Stat_t)
	if !openedInfo.Mode().IsRegular() || openedInfo.Mode().Perm() != 0o600 || openedInfo.Size() != info.Size() || !ok || int(openedStat.Uid) != endpoint.OwnerUID {
		return completionMarker{}, errors.New("completion marker changed during authentication")
	}
	data, err := io.ReadAll(io.LimitReader(file, completionMarkerMaxBytes+1))
	if err != nil {
		return completionMarker{}, err
	}
	if len(data) > completionMarkerMaxBytes {
		return completionMarker{}, errors.New("completion marker exceeds size limit")
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) != 3 || lines[2] != "" || !strings.HasPrefix(lines[0], "nonce=") || !strings.HasPrefix(lines[1], "rc=") {
		return completionMarker{}, errors.New("completion marker shape is invalid")
	}
	if strings.TrimPrefix(lines[0], "nonce=") != endpoint.Nonce {
		return completionMarker{}, errors.New("completion marker nonce is invalid")
	}
	rc, err := strconv.Atoi(strings.TrimPrefix(lines[1], "rc="))
	if err != nil {
		return completionMarker{}, errors.New("completion marker rc is invalid")
	}
	status := completionStatus{RC: rc}
	if err := status.validate(); err != nil {
		return completionMarker{}, err
	}
	return completionMarker{Status: status, ModTime: info.ModTime()}, nil
}

type completionShellTools struct {
	MovePath   string
	RemovePath string
}

func defaultCompletionShellTools() completionShellTools {
	return completionShellTools{MovePath: "/bin/mv", RemovePath: "/bin/rm"}
}

const completionShellTemplate = `marker_dir=$1; nonce=$2; move_bin=$3; remove_bin=$4; shift 4; "$@"; rc=$?; tmp="$marker_dir/.complete.$nonce.tmp"; final="$marker_dir/complete"; publish_rc=0; umask 077; printf 'nonce=%s\nrc=%s\n' "$nonce" "$rc" >"$tmp" && chmod 600 "$tmp" && "$move_bin" -- "$tmp" "$final" || publish_rc=$?; /bin/sleep 0.25; if [ "$publish_rc" -ne 0 ]; then "$remove_bin" -f -- "$tmp"; exit 125; fi; exit "$rc"`

func completionShellScript(tools completionShellTools) (string, error) {
	if !filepath.IsAbs(tools.MovePath) || !filepath.IsAbs(tools.RemovePath) {
		return "", errors.New("completion shell tools must be absolute")
	}
	return completionShellTemplate, nil
}

func completionProtocolCommand(spec confinementSpec, endpoint completionEndpoint) ([]string, error) {
	return completionProtocolCommandWithTools(spec, endpoint, defaultCompletionShellTools())
}

func completionProtocolCommandWithTools(spec confinementSpec, endpoint completionEndpoint, tools completionShellTools) ([]string, error) {
	if err := validateCompletionThreatScope(false); err != nil {
		return nil, err
	}
	if endpoint.Dir == "" || endpoint.Path != filepath.Join(endpoint.Dir, "complete") || endpoint.Nonce == "" || pathsOverlap(endpoint.Dir, spec.Root) {
		return nil, errors.New("completion endpoint is invalid or exposed to bwrap")
	}
	command := append([]string(nil), spec.Command...)
	scriptAt := -1
	for i := 0; i+3 < len(command); i++ {
		if command[i] == "/bin/sh" && command[i+1] == "-c" {
			scriptAt = i + 2
			break
		}
	}
	if scriptAt < 0 || command[scriptAt+1] != "sh" {
		return nil, errors.New("trusted outer shell shape is invalid")
	}
	script, err := completionShellScript(tools)
	if err != nil {
		return nil, err
	}
	command[scriptAt] = script
	insertAt := scriptAt + 2
	command = append(command[:insertAt], append([]string{endpoint.Dir, endpoint.Nonce, tools.MovePath, tools.RemovePath}, command[insertAt:]...)...)
	if strings.Count(strings.Join(command, "\x00"), "/bin/sleep 0.25") != 1 {
		return nil, errors.New("completion protocol must preserve exactly one 250ms retention")
	}
	return command, nil
}

func validateCompletionThreatScope(hostileSameUID bool) error {
	if hostileSameUID {
		return errors.New("pathname completion markers do not protect against hostile same-UID host processes")
	}
	return nil
}

type commandWaitResult struct {
	Err    error
	At     time.Time
	Status completionStatus
}

type commandWaitHandle struct {
	done   chan struct{}
	result commandWaitResult
}

func startCommandWait(wait func() error, now func() time.Time) *commandWaitHandle {
	handle := &commandWaitHandle{done: make(chan struct{})}
	go func() {
		err := wait()
		handle.result = commandWaitResult{Err: err, At: now(), Status: completionStatusFromWaitError(err)}
		close(handle.done)
	}()
	return handle
}

func (h *commandWaitHandle) Completed() bool {
	select {
	case <-h.done:
		return true
	default:
		return false
	}
}

func (h *commandWaitHandle) Join(ctx context.Context) (commandWaitResult, error) {
	select {
	case <-h.done:
		return h.result, nil
	case <-ctx.Done():
		return commandWaitResult{}, fmt.Errorf("join command Wait: %w", ctx.Err())
	}
}

func completionStatusFromWaitError(err error) completionStatus {
	if err == nil {
		return completionStatus{RC: 0}
	}
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ProcessState == nil {
		return completionStatus{RC: -1}
	}
	rc := exitErr.ExitCode()
	if rc >= 0 {
		return completionStatus{RC: rc}
	}
	if status, ok := exitErr.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		return completionStatus{RC: 128 + int(status.Signal())}
	}
	return completionStatus{RC: -1}
}

func reconcileCompletionStatus(marker, waited completionStatus) error {
	if err := marker.validate(); err != nil {
		return fmt.Errorf("completion marker status: %w", err)
	}
	if err := waited.validate(); err != nil {
		return fmt.Errorf("command Wait status: %w", err)
	}
	if marker != waited {
		return fmt.Errorf("completion marker status %#v does not match command Wait status %#v", marker, waited)
	}
	return nil
}

func reconcileAndCleanupCompletion(marker, waited completionStatus, endpoint completionEndpoint, cleanup func(completionEndpoint) error, diagnostics *cgroupAccountingDiagnostics) error {
	statusErr := reconcileCompletionStatus(marker, waited)
	cleanupErr := cleanup(endpoint)
	if cleanupErr != nil && diagnostics != nil {
		diagnostics.SampleErrorCount++
		diagnostics.Errors = append(diagnostics.Errors, cgroupSamplerError{Stage: "completion_cleanup", ErrorSHA256: shaString(cleanupErr.Error())})
	}
	return errors.Join(statusErr, cleanupErr)
}

func validateCompletionOrdering(targetExitedAt, markerAt, waitAt time.Time) error {
	if markerAt.Before(targetExitedAt) {
		return errors.New("completion marker precedes target exit")
	}
	if waitAt.Before(markerAt) {
		return errors.New("command Wait precedes completion marker")
	}
	return nil
}

func waitForAuthenticatedCompletion(ctx context.Context, endpoint completionEndpoint, handle *commandWaitHandle, now func() time.Time) (completionMarker, time.Time, error) {
	ticker := time.NewTicker(completionMarkerPollInterval)
	defer ticker.Stop()
	for {
		marker, err := readCompletionMarker(endpoint)
		if err == nil {
			observedAt := now()
			if handle.Completed() && handle.result.At.Before(observedAt) {
				return completionMarker{}, time.Time{}, errors.New("command Wait completed before marker authentication")
			}
			return marker, observedAt, nil
		}
		if !os.IsNotExist(err) {
			return completionMarker{}, time.Time{}, fmt.Errorf("authenticate completion marker: %w", err)
		}
		select {
		case <-ctx.Done():
			return completionMarker{}, time.Time{}, fmt.Errorf("wait for completion marker: %w", ctx.Err())
		case <-handle.done:
			return completionMarker{}, time.Time{}, errors.New("command Wait completed before authenticated completion marker")
		case <-ticker.C:
		}
	}
}

func snapshotCompletedOutput(handle *commandWaitHandle, stdout, stderr *limitedBuffer) ([]byte, []byte, error) {
	if !handle.Completed() {
		return nil, nil, errors.New("cannot inspect output before command Wait joins")
	}
	return append([]byte(nil), stdout.Buffer.Bytes()...), append([]byte(nil), stderr.Buffer.Bytes()...), nil
}

var errControllerFatalLiveWait = errors.New("controller-fatal live command Wait after hard reap")

func abortAndJoinCompletionPhased(endpoint completionEndpoint, handle *commandWaitHandle, stopTimeout, joinTimeout time.Duration, stop, hardStop func(context.Context) error, cleanup func(completionEndpoint) error) error {
	if handle == nil || stop == nil || hardStop == nil || cleanup == nil || stopTimeout <= 0 || joinTimeout <= 0 {
		return errors.New("completion abort dependencies and deadlines are required")
	}
	stopCtx, stopCancel := context.WithTimeout(context.Background(), stopTimeout)
	stopErr := stop(stopCtx)
	stopCancel()

	joinCtx, joinCancel := context.WithTimeout(context.Background(), joinTimeout)
	_, joinErr := handle.Join(joinCtx)
	joinCancel()

	var hardStopErr, finalJoinErr error
	if joinErr != nil && !handle.Completed() {
		hardStopCtx, hardStopCancel := context.WithTimeout(context.Background(), stopTimeout)
		hardStopErr = hardStop(hardStopCtx)
		hardStopCancel()

		finalJoinCtx, finalJoinCancel := context.WithTimeout(context.Background(), joinTimeout)
		_, finalJoinErr = handle.Join(finalJoinCtx)
		finalJoinCancel()
	}

	cleanupErr := cleanup(endpoint)
	if finalJoinErr != nil && !handle.Completed() {
		fatalErr := fmt.Errorf("%w: initial_join=%v hard_stop=%v final_join=%v", errControllerFatalLiveWait, joinErr, hardStopErr, finalJoinErr)
		return errors.Join(stopErr, fatalErr, cleanupErr)
	}
	return errors.Join(stopErr, joinErr, hardStopErr, finalJoinErr, cleanupErr)
}

type systemctlRunner interface {
	Run(context.Context, ...string) ([]byte, error)
}

type execSystemctlRunner struct {
	Path string
	Env  []string
}

func (r execSystemctlRunner) Run(ctx context.Context, args ...string) ([]byte, error) {
	command := exec.CommandContext(ctx, r.Path, append([]string{"--user"}, args...)...)
	command.Env = append([]string(nil), r.Env...)
	return command.CombinedOutput()
}

func stopAndCleanTransientUnit(ctx context.Context, unitName string, env []string) error {
	systemctl, err := exec.LookPath("systemctl")
	if err != nil {
		return errors.New("systemctl unavailable for transient unit cleanup")
	}
	return stopAndCleanTransientUnitWithRunner(ctx, unitName, execSystemctlRunner{Path: systemctl, Env: env})
}

func stopAndCleanTransientUnitWithRunner(ctx context.Context, unitName string, runner systemctlRunner) error {
	if runner == nil {
		return errors.New("systemctl runner is required")
	}
	unit := unitName + ".service"
	run := func(args ...string) error {
		output, commandErr := runner.Run(ctx, args...)
		if commandErr != nil {
			if explicitUnitNotFound(unit, output) {
				return nil
			}
			return fmt.Errorf("systemctl %s: %w: %s", strings.Join(args, " "), commandErr, strings.TrimSpace(string(output)))
		}
		return nil
	}
	stopErr := run("stop", unit)
	resetErr := run("reset-failed", unit)
	if stopErr != nil || resetErr != nil {
		return errors.Join(stopErr, resetErr)
	}
	for {
		output, showErr := runner.Run(ctx, "show", unit, "--property", "LoadState", "--value")
		if showErr != nil {
			if explicitUnitNotFound(unit, output) {
				return nil
			}
			return fmt.Errorf("systemctl show %s: %w: %s", unit, showErr, strings.TrimSpace(string(output)))
		}
		state := strings.TrimSpace(string(output))
		if state == "not-found" || state == "" {
			return nil
		}
		timer := time.NewTimer(cgroupResolveRetryInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return fmt.Errorf("clean transient unit %s: %w", unit, ctx.Err())
		case <-timer.C:
		}
	}
}

func explicitUnitNotFound(unit string, output []byte) bool {
	message := strings.ToLower(strings.TrimSpace(string(output)))
	unit = strings.ToLower(unit)
	return strings.Contains(message, "unit "+unit+" could not be found") || strings.Contains(message, "unit "+unit+" not found")
}

type armInvocation struct {
	Response              armResponse
	RawStdout             []byte
	RawStderr             []byte
	ActualBudget          resourceBudget
	AccountingDiagnostics cgroupAccountingDiagnostics
}

type cgroupStats struct {
	CPUUsageMicros int64
	MemoryPeak     int64
	PIDsPeak       int
}

type cgroupSamplerConfig struct {
	ResolveDeadline       time.Duration
	ResolveAttemptTimeout time.Duration
	ResolveRetryInterval  time.Duration
	SampleInterval        time.Duration
	SampleTimeout         time.Duration
	JoinTimeout           time.Duration
}

func defaultCgroupSamplerConfig() cgroupSamplerConfig {
	return cgroupSamplerConfig{
		ResolveDeadline:       cgroupResolveDeadline,
		ResolveAttemptTimeout: cgroupResolveAttemptTimeout,
		ResolveRetryInterval:  cgroupResolveRetryInterval,
		SampleInterval:        cgroupSampleInterval,
		SampleTimeout:         cgroupSampleTimeout,
		JoinTimeout:           cgroupMonitorJoinTimeout,
	}
}

func (c cgroupSamplerConfig) validate() error {
	if c.ResolveDeadline <= 0 || c.ResolveAttemptTimeout <= 0 || c.ResolveRetryInterval <= 0 || c.SampleInterval <= 0 || c.SampleTimeout <= 0 || c.JoinTimeout <= 0 {
		return errors.New("cgroup sampler deadlines must be positive")
	}
	if c.ResolveAttemptTimeout > c.ResolveDeadline || c.SampleTimeout > c.JoinTimeout {
		return errors.New("cgroup sampler deadline ordering is invalid")
	}
	return nil
}

type cgroupSamplerError struct {
	Stage       string `json:"stage"`
	ErrorSHA256 string `json:"error_sha256"`
}

type cgroupAccountingDiagnostics struct {
	ResolveAttempts        int                  `json:"resolve_attempts"`
	ResolveLatencyMicros   int64                `json:"resolve_latency_micros"`
	SampleAttempts         int                  `json:"sample_attempts"`
	SuccessfulSamples      int                  `json:"successful_samples"`
	SuccessfulCPUReads     int                  `json:"successful_cpu_reads"`
	SampleErrorCount       int                  `json:"sample_error_count"`
	SampleLatencyMicros    int64                `json:"sample_latency_micros"`
	MaxSampleLatencyMicros int64                `json:"max_sample_latency_micros"`
	FinalReadLatencyMicros int64                `json:"final_read_latency_micros"`
	CommandWaitWallMillis  int64                `json:"command_wait_wall_millis"`
	MonitorJoinTailMillis  int64                `json:"monitor_join_tail_millis"`
	Errors                 []cgroupSamplerError `json:"errors"`
}

type cgroupAccountingResult struct {
	PeriodicPeak cgroupStats
	Final        cgroupStats
	Diagnostics  cgroupAccountingDiagnostics
}

type resolvedControlGroup struct {
	Root     string
	Attempts int
	Latency  time.Duration
}

type controlGroupResolver interface {
	Resolve(context.Context, string) (resolvedControlGroup, error)
}

type cgroupStatsReader interface {
	ReadFile(context.Context, string) ([]byte, error)
}

type contextCgroupFileReader struct{}

func (contextCgroupFileReader) ReadFile(ctx context.Context, path string) ([]byte, error) {
	type readResult struct {
		data []byte
		err  error
	}
	result := make(chan readResult, 1)
	go func() {
		data, err := os.ReadFile(path)
		result <- readResult{data: data, err: err}
	}()
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("read %s: %w", filepath.Base(path), ctx.Err())
	case read := <-result:
		if read.err != nil {
			return nil, fmt.Errorf("read %s: %w", filepath.Base(path), read.err)
		}
		return read.data, nil
	}
}

type systemdControlGroupResolver struct {
	Config cgroupSamplerConfig
	Env    []string
}

func (r systemdControlGroupResolver) Resolve(ctx context.Context, unitName string) (resolvedControlGroup, error) {
	started := time.Now()
	resolved := resolvedControlGroup{}
	if err := r.Config.validate(); err != nil {
		return resolved, err
	}
	resolveCtx, cancel := context.WithTimeout(ctx, r.Config.ResolveDeadline)
	defer cancel()
	var lastErr error
	for {
		resolved.Attempts++
		attemptCtx, attemptCancel := context.WithTimeout(resolveCtx, r.Config.ResolveAttemptTimeout)
		show := exec.CommandContext(attemptCtx, "systemctl", "--user", "show", unitName+".service", "--property", "ControlGroup", "--value")
		show.Env = append([]string(nil), r.Env...)
		out, err := show.Output()
		attemptCancel()
		if err == nil {
			root, rootErr := validatedCgroupRoot(strings.TrimSpace(string(out)))
			if rootErr == nil {
				resolved.Root = root
				resolved.Latency = time.Since(started)
				return resolved, nil
			}
			lastErr = rootErr
		} else {
			lastErr = err
		}
		timer := time.NewTimer(r.Config.ResolveRetryInterval)
		select {
		case <-resolveCtx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			resolved.Latency = time.Since(started)
			if lastErr == nil {
				lastErr = resolveCtx.Err()
			}
			return resolved, fmt.Errorf("resolve ControlGroup after %d attempt(s): %w", resolved.Attempts, lastErr)
		case <-timer.C:
		}
	}
}

func validatedCgroupRoot(controlGroup string) (string, error) {
	if controlGroup == "" || !strings.HasPrefix(controlGroup, "/") || strings.Contains(controlGroup, "..") {
		return "", errors.New("invalid transient ControlGroup")
	}
	base := filepath.Clean("/sys/fs/cgroup")
	root := filepath.Join(base, filepath.FromSlash(strings.TrimPrefix(controlGroup, "/")))
	if root == base || !strings.HasPrefix(root, base+string(os.PathSeparator)) {
		return "", errors.New("transient ControlGroup escapes cgroup root")
	}
	info, err := os.Stat(root)
	if err != nil {
		return "", fmt.Errorf("stat transient ControlGroup: %w", err)
	}
	if !info.IsDir() {
		return "", errors.New("transient ControlGroup is not a directory")
	}
	return root, nil
}

type cgroupSample struct {
	Stats   cgroupStats
	CPURead bool
	Elapsed time.Duration
}

func readCgroupStatsFromResolvedRoot(ctx context.Context, root string, reader cgroupStatsReader) (cgroupStats, bool, error) {
	var stats cgroupStats
	cpuData, err := reader.ReadFile(ctx, filepath.Join(root, "cpu.stat"))
	if err != nil {
		return stats, false, err
	}
	foundCPU := false
	for _, line := range strings.Split(string(cpuData), "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 || fields[0] != "usage_usec" {
			continue
		}
		stats.CPUUsageMicros, err = strconv.ParseInt(fields[1], 10, 64)
		if err != nil || stats.CPUUsageMicros < 0 {
			return stats, false, fmt.Errorf("parse cpu.stat usage_usec: %w", err)
		}
		foundCPU = true
		break
	}
	if !foundCPU {
		return stats, false, errors.New("cpu.stat omits usage_usec")
	}
	readPeak := func(name string) (int64, error) {
		data, readErr := reader.ReadFile(ctx, filepath.Join(root, name))
		if readErr != nil {
			return 0, readErr
		}
		value, parseErr := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
		if parseErr != nil {
			return 0, fmt.Errorf("parse %s: %w", name, parseErr)
		}
		if value <= 0 {
			return 0, fmt.Errorf("%s must be positive", name)
		}
		return value, nil
	}
	stats.MemoryPeak, err = readPeak("memory.peak")
	if err != nil {
		return stats, true, err
	}
	pidsPeak, err := readPeak("pids.peak")
	if err != nil {
		return stats, true, err
	}
	stats.PIDsPeak = int(pidsPeak)
	if int64(stats.PIDsPeak) != pidsPeak {
		return stats, true, errors.New("pids.peak overflows int")
	}
	return stats, true, nil
}

func sampleCgroupStats(ctx context.Context, root string, reader cgroupStatsReader, timeout time.Duration) (cgroupStats, bool, time.Duration, error) {
	if timeout <= 0 {
		return cgroupStats{}, false, 0, errors.New("cgroup sample timeout must be positive")
	}
	sampleCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	started := time.Now()
	stats, cpuRead, err := readCgroupStatsFromResolvedRoot(sampleCtx, root, reader)
	return stats, cpuRead, time.Since(started), err
}

type cgroupSampler struct {
	root        string
	reader      cgroupStatsReader
	config      cgroupSamplerConfig
	diagnostics cgroupAccountingDiagnostics
}

func newCgroupSampler(ctx context.Context, unitName string, resolver controlGroupResolver, reader cgroupStatsReader, config cgroupSamplerConfig) (*cgroupSampler, error) {
	if resolver == nil || reader == nil {
		return nil, errors.New("cgroup sampler dependencies are required")
	}
	if err := config.validate(); err != nil {
		return nil, err
	}
	resolved, err := resolver.Resolve(ctx, unitName)
	sampler := &cgroupSampler{root: resolved.Root, reader: reader, config: config, diagnostics: cgroupAccountingDiagnostics{ResolveAttempts: resolved.Attempts, ResolveLatencyMicros: durationMicrosCeil(resolved.Latency)}}
	if err != nil {
		sampler.recordError(&sampler.diagnostics, "resolve", err)
		return sampler, fmt.Errorf("resolve cgroup once: %w", err)
	}
	if strings.TrimSpace(resolved.Root) == "" {
		err := errors.New("resolved cgroup root is empty")
		sampler.recordError(&sampler.diagnostics, "resolve", err)
		return sampler, err
	}
	return sampler, nil
}

func (s *cgroupSampler) sampleOnce(ctx context.Context, final bool) (cgroupSample, error) {
	stats, cpuRead, elapsed, err := sampleCgroupStats(ctx, s.root, s.reader, s.config.SampleTimeout)
	return cgroupSample{Stats: stats, CPURead: cpuRead, Elapsed: elapsed}, err
}

func (s *cgroupSampler) recordError(diagnostics *cgroupAccountingDiagnostics, stage string, err error) {
	diagnostics.SampleErrorCount++
	diagnostics.Errors = append(diagnostics.Errors, cgroupSamplerError{Stage: stage, ErrorSHA256: shaString(err.Error())})
}

type runningCgroupSampler struct {
	stop   chan struct{}
	result chan cgroupSamplerRunResult
	once   sync.Once
}

type cgroupSamplerRunResult struct {
	accounting cgroupAccountingResult
	err        error
}

func (s *cgroupSampler) start(ctx context.Context) *runningCgroupSampler {
	running := &runningCgroupSampler{stop: make(chan struct{}), result: make(chan cgroupSamplerRunResult, 1)}
	go func() {
		diagnostics := s.diagnostics
		var periodicPeak cgroupStats
		take := func(stage string, final bool) (cgroupSample, error) {
			diagnostics.SampleAttempts++
			sample, err := s.sampleOnce(ctx, final)
			micros := durationMicrosCeil(sample.Elapsed)
			diagnostics.SampleLatencyMicros += micros
			if micros > diagnostics.MaxSampleLatencyMicros {
				diagnostics.MaxSampleLatencyMicros = micros
			}
			if final {
				diagnostics.FinalReadLatencyMicros = micros
			}
			if sample.CPURead {
				diagnostics.SuccessfulCPUReads++
			}
			if err != nil {
				s.recordError(&diagnostics, stage, err)
				return sample, err
			}
			diagnostics.SuccessfulSamples++
			return sample, nil
		}
		if sample, err := take("sample", false); err == nil {
			periodicPeak = maxCgroupStats(periodicPeak, sample.Stats)
		}
		ticker := time.NewTicker(s.config.SampleInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				err := fmt.Errorf("cgroup sampler context: %w", ctx.Err())
				s.recordError(&diagnostics, "context", err)
				running.result <- cgroupSamplerRunResult{accounting: cgroupAccountingResult{PeriodicPeak: periodicPeak, Diagnostics: diagnostics}, err: err}
				return
			case <-running.stop:
				final, err := take("final", true)
				accounting := cgroupAccountingResult{PeriodicPeak: periodicPeak, Final: final.Stats, Diagnostics: diagnostics}
				if err != nil {
					err = fmt.Errorf("final post-Wait cgroup read: %w", err)
				}
				running.result <- cgroupSamplerRunResult{accounting: accounting, err: err}
				return
			case <-ticker.C:
				if sample, err := take("sample", false); err == nil {
					periodicPeak = maxCgroupStats(periodicPeak, sample.Stats)
				}
			}
		}
	}()
	return running
}

func (s *runningCgroupSampler) finish(ctx context.Context) (cgroupAccountingResult, error) {
	s.once.Do(func() { close(s.stop) })
	select {
	case result := <-s.result:
		return result.accounting, result.err
	case <-ctx.Done():
		return cgroupAccountingResult{}, fmt.Errorf("join cgroup sampler: %w", ctx.Err())
	}
}

func takeCompletionAccountingSnapshot(ctx context.Context, sampler *cgroupSampler) (cgroupAccountingResult, error) {
	if sampler == nil {
		return cgroupAccountingResult{}, errors.New("cgroup sampler is required")
	}
	diagnostics := sampler.diagnostics
	diagnostics.SampleAttempts++
	sample, err := sampler.sampleOnce(ctx, true)
	micros := durationMicrosCeil(sample.Elapsed)
	diagnostics.SampleLatencyMicros = micros
	diagnostics.MaxSampleLatencyMicros = micros
	diagnostics.FinalReadLatencyMicros = micros
	if sample.CPURead {
		diagnostics.SuccessfulCPUReads++
	}
	if err != nil {
		sampler.recordError(&diagnostics, "final", err)
		return cgroupAccountingResult{Final: sample.Stats, Diagnostics: diagnostics}, fmt.Errorf("first post-completion cgroup read: %w", err)
	}
	diagnostics.SuccessfulSamples++
	return cgroupAccountingResult{Final: sample.Stats, Diagnostics: diagnostics}, nil
}

func maxCgroupStats(left, right cgroupStats) cgroupStats {
	if right.CPUUsageMicros > left.CPUUsageMicros {
		left.CPUUsageMicros = right.CPUUsageMicros
	}
	if right.MemoryPeak > left.MemoryPeak {
		left.MemoryPeak = right.MemoryPeak
	}
	if right.PIDsPeak > left.PIDsPeak {
		left.PIDsPeak = right.PIDsPeak
	}
	return left
}

func durationMicrosCeil(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	return int64((duration + time.Microsecond - 1) / time.Microsecond)
}

func durationMillisCeil(duration time.Duration) int64 {
	if duration <= 0 {
		return 0
	}
	return int64((duration + time.Millisecond - 1) / time.Millisecond)
}

func finalizeArmAccounting(started, waitedAt, joinedAt time.Time, accounting cgroupAccountingResult, stdoutBytes, stderrBytes, caseCount int) (resourceBudget, cgroupAccountingDiagnostics, error) {
	return finalizeArmAccountingAtCompletion(started, waitedAt, joinedAt, accounting, stdoutBytes, stderrBytes, caseCount)
}

func finalizeArmAccountingAtCompletion(started, completedAt, joinedAt time.Time, accounting cgroupAccountingResult, stdoutBytes, stderrBytes, caseCount int) (resourceBudget, cgroupAccountingDiagnostics, error) {
	diagnostics := accounting.Diagnostics
	diagnostics.CommandWaitWallMillis = durationMillisCeil(completedAt.Sub(started))
	diagnostics.MonitorJoinTailMillis = durationMillisCeil(joinedAt.Sub(completedAt))
	if diagnostics.SuccessfulCPUReads == 0 {
		return resourceBudget{}, diagnostics, errors.New("zero successful CPU reads is invalid")
	}
	if accounting.Final.CPUUsageMicros < 0 || accounting.Final.MemoryPeak <= 0 || accounting.Final.PIDsPeak <= 0 {
		return resourceBudget{}, diagnostics, errors.New("final post-Wait cgroup stats are incomplete")
	}
	actual := resourceBudget{
		WallTimeMillis: diagnostics.CommandWaitWallMillis,
		CPUTimeMillis:  (accounting.Final.CPUUsageMicros + 999) / 1000,
		MaxRSSBytes:    accounting.Final.MemoryPeak,
		ProcessCount:   accounting.Final.PIDsPeak,
		StdoutBytes:    stdoutBytes,
		StderrBytes:    stderrBytes,
		CaseCount:      caseCount,
	}
	return actual, diagnostics, nil
}

func runConfinedArm(ctx context.Context, runtimePath string, req armRequest, controllerRoot string, limits resourceBudget) (controllerResult, error) {
	result, _, _, _, err := runConfinedArmMeasured(ctx, runtimePath, req, controllerRoot, limits)
	return result, err
}

func runConfinedArmMeasured(ctx context.Context, runtimePath string, req armRequest, controllerRoot string, limits resourceBudget) (controllerResult, armInvocation, string, string, error) {
	if err := os.MkdirAll(controllerRoot, 0o700); err != nil {
		return controllerResult{}, armInvocation{}, "", "", err
	}
	armRoot := filepath.Join(controllerRoot, "arm")
	if err := os.Mkdir(armRoot, 0o700); err != nil {
		return controllerResult{}, armInvocation{}, "", "", fmt.Errorf("create isolated arm root: %w", err)
	}
	invocation, err := invokeConfinedArmMeasured(ctx, runtimePath, req, armRoot, limits)
	if err != nil {
		return controllerResult{}, invocation, "", "", err
	}
	result, err := inspectArmResult(req, invocation.Response, filepath.Join(armRoot, "db", "c3.db"))
	if err != nil {
		return controllerResult{}, invocation, "", "", fmt.Errorf("inspect arm result: %w", err)
	}
	dumpBytes, err := json.Marshal(result.Database)
	if err != nil {
		return controllerResult{}, armInvocation{}, "", "", err
	}
	invocation.ActualBudget.SQLiteRowCount = result.Database.SQLiteRowCount
	invocation.ActualBudget.LogicalDumpBytes = len(dumpBytes)
	projectHash, err := directoryTreeSHA256(filepath.Join(armRoot, "project"))
	if err != nil {
		return result, invocation, "", "", err
	}
	c3Hash, err := directoryTreeSHA256(filepath.Join(armRoot, "c3"))
	if err != nil {
		return result, invocation, projectHash, "", err
	}
	if !actualBudgetComplete(invocation.ActualBudget) {
		return controllerResult{}, armInvocation{}, "", "", errors.New("kernel/resource actuals are incomplete")
	}
	return result, invocation, projectHash, c3Hash, nil
}

func invokeConfinedArm(ctx context.Context, runtimePath string, req armRequest, armRoot string, limits resourceBudget) (armResponse, error) {
	invocation, err := invokeConfinedArmMeasured(ctx, runtimePath, req, armRoot, limits)
	return invocation.Response, err
}

func invokeConfinedArmMeasured(ctx context.Context, runtimePath string, req armRequest, armRoot string, limits resourceBudget) (armInvocation, error) {
	if req.Schema != armRequestSchema {
		return armInvocation{}, errors.New("arm request schema mismatch")
	}
	endpoint, err := newCompletionEndpoint(filepath.Dir(armRoot), armRoot, time.Now)
	if err != nil {
		return armInvocation{}, err
	}
	defer guardCompletionEndpointCleanup(endpoint, cleanupCompletionEndpoint)()
	spec, err := newConfinementSpec(runtimePath, armRoot, limits)
	if err != nil {
		_ = cleanupCompletionEndpoint(endpoint)
		return armInvocation{}, err
	}
	input, err := json.Marshal(req)
	if err != nil {
		_ = cleanupCompletionEndpoint(endpoint)
		return armInvocation{}, err
	}
	args, err := completionProtocolCommand(spec, endpoint)
	if err != nil {
		_ = cleanupCompletionEndpoint(endpoint)
		return armInvocation{}, err
	}
	args = append(args, "--arm", "--db", "/work/db/c3.db", "--project", "/work/project", "--c3", "/work/c3")
	command := exec.CommandContext(ctx, args[0], args[1:]...)
	command.Env = systemdTransportEnv()
	command.Stdin = bytes.NewReader(input)
	stdout := &limitedBuffer{Limit: 16 << 20}
	stderr := &limitedBuffer{Limit: 1 << 20}
	command.Stdout = stdout
	command.Stderr = stderr
	started := time.Now()
	if err := command.Start(); err != nil {
		_ = cleanupCompletionEndpoint(endpoint)
		return armInvocation{}, fmt.Errorf("start confined arm: %w", err)
	}
	waitHandle := startCommandWait(command.Wait, time.Now)
	config := defaultCgroupSamplerConfig()
	resolver := systemdControlGroupResolver{Config: config, Env: systemdTransportEnv()}
	sampler, resolveErr := newCgroupSampler(ctx, spec.UnitName, resolver, contextCgroupFileReader{}, config)
	failAfterStart := func(cause error, diagnostics cgroupAccountingDiagnostics) (armInvocation, error) {
		cleanupErr := abortAndJoinCompletionPhased(endpoint, waitHandle, config.ResolveDeadline, config.ResolveDeadline, func(stopCtx context.Context) error {
			return stopAndCleanTransientUnit(stopCtx, spec.UnitName, systemdTransportEnv())
		}, func(hardStopCtx context.Context) error {
			select {
			case <-hardStopCtx.Done():
				return hardStopCtx.Err()
			default:
			}
			if command.Process == nil {
				return errors.New("confined arm process is unavailable for hard stop")
			}
			err := command.Process.Kill()
			if errors.Is(err, os.ErrProcessDone) {
				return nil
			}
			return err
		}, cleanupCompletionEndpoint)
		if errors.Is(cleanupErr, errControllerFatalLiveWait) {
			return armInvocation{}, errors.Join(cause, cleanupErr)
		}
		rawStdout, rawStderr, outputErr := snapshotCompletedOutput(waitHandle, stdout, stderr)
		waitedAt := time.Now()
		if waitHandle.Completed() {
			waitedAt = waitHandle.result.At
		}
		diagnostics.CommandWaitWallMillis = durationMillisCeil(waitedAt.Sub(started))
		invocation := armInvocation{
			RawStdout: rawStdout, RawStderr: rawStderr,
			ActualBudget: resourceBudget{
				WallTimeMillis: diagnostics.CommandWaitWallMillis,
				StdoutBytes:    len(rawStdout), StderrBytes: len(rawStderr), CaseCount: len(req.Queries),
			},
			AccountingDiagnostics: diagnostics,
		}
		return invocation, errors.Join(cause, cleanupErr, outputErr)
	}
	if resolveErr != nil {
		diagnostics := cgroupAccountingDiagnostics{}
		if sampler != nil {
			diagnostics = sampler.diagnostics
		}
		invocation, failure := failAfterStart(fmt.Errorf("confined arm cgroup resolution failed: %w", resolveErr), diagnostics)
		return invocation, failure
	}
	markerCtx, markerCancel := context.WithTimeout(ctx, time.Duration(limits.WallTimeMillis)*time.Millisecond)
	marker, markerAt, markerErr := waitForAuthenticatedCompletion(markerCtx, endpoint, waitHandle, time.Now)
	markerCancel()
	if markerErr != nil {
		invocation, failure := failAfterStart(fmt.Errorf("confined arm completion failed: %w", markerErr), sampler.diagnostics)
		return invocation, failure
	}
	accounting, accountingErr := takeCompletionAccountingSnapshot(ctx, sampler)
	if accountingErr != nil {
		invocation, failure := failAfterStart(fmt.Errorf("confined arm kernel accounting failed: %w", accountingErr), accounting.Diagnostics)
		return invocation, failure
	}
	joinCtx, joinCancel := context.WithTimeout(context.Background(), config.ResolveDeadline)
	waitResult, joinErr := waitHandle.Join(joinCtx)
	joinCancel()
	if joinErr != nil {
		invocation, failure := failAfterStart(joinErr, accounting.Diagnostics)
		return invocation, failure
	}
	if err := validateCompletionOrdering(markerAt, markerAt, waitResult.At); err != nil {
		invocation, failure := failAfterStart(err, accounting.Diagnostics)
		return invocation, failure
	}
	completionErr := reconcileAndCleanupCompletion(marker.Status, waitResult.Status, endpoint, cleanupCompletionEndpoint, &accounting.Diagnostics)
	rawStdout, rawStderr, outputErr := snapshotCompletedOutput(waitHandle, stdout, stderr)
	actual, diagnostics, finalizeErr := finalizeArmAccountingAtCompletion(started, markerAt, waitResult.At, accounting, len(rawStdout), len(rawStderr), len(req.Queries))
	invocation := armInvocation{RawStdout: rawStdout, RawStderr: rawStderr, ActualBudget: actual, AccountingDiagnostics: diagnostics}
	if accountingCause := errors.Join(completionErr, outputErr, finalizeErr); accountingCause != nil {
		return invocation, fmt.Errorf("confined arm completion accounting failed: %w", accountingCause)
	}
	if waitResult.Err != nil {
		return invocation, fmt.Errorf("confined arm failed: %w: %s", waitResult.Err, strings.TrimSpace(string(rawStderr)))
	}
	response, err := decodeArmResponse(bytes.NewReader(rawStdout), 16<<20)
	if err != nil {
		return invocation, fmt.Errorf("strict confined arm response: %w", err)
	}
	invocation.Response = response
	return invocation, nil
}

func directoryTreeSHA256(root string) (string, error) {
	h := sha256.New()
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s\x00%s\x00%o\x00", filepath.ToSlash(rel), entry.Type().String(), info.Mode().Perm())
		if entry.Type().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			_, copyErr := io.Copy(h, file)
			closeErr := file.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		}
		h.Write([]byte{0})
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func loadFixtures(path string) ([]fixtureCase, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	sum := sha256.Sum256(data)
	hash := hex.EncodeToString(sum[:])
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024), 16<<20)
	var fixtures []fixtureCase
	seen := map[string]bool{}
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		var f fixtureCase
		if err := decodeStrictBytes(line, &f); err != nil {
			return nil, "", err
		}
		if err := validateFixture(f); err != nil {
			return nil, "", err
		}
		if seen[f.CaseID] {
			return nil, "", fmt.Errorf("duplicate fixture %s", f.CaseID)
		}
		seen[f.CaseID] = true
		fixtures = append(fixtures, f)
	}
	if err := scanner.Err(); err != nil {
		return nil, "", err
	}
	if len(fixtures) == 0 {
		return nil, "", errors.New("fixture file is empty")
	}
	return fixtures, hash, nil
}
func decodeBenchmark(r io.Reader) (benchmarkConfig, error) {
	var b benchmarkConfig
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&b); err != nil {
		return b, err
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return b, errors.New("trailing benchmark JSON")
		}
		return b, err
	}
	if err := validateAdapterImplementationConfig(b); err != nil {
		return b, err
	}
	return b, nil
}
func loadBenchmark(path string) (benchmarkConfig, error) {
	f, err := os.Open(path)
	if err != nil {
		return benchmarkConfig{}, err
	}
	defer f.Close()
	return decodeBenchmark(f)
}
