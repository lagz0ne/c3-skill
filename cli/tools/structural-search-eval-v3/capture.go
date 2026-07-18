package main

// capture.go is the unchanged-C3 B-v3 capture boundary. It deliberately has
// no candidate, capability, activation, or controller-authority inputs. The
// caller supplies only frozen generic fixtures and receives per-case row
// sizes, ids, and hashes from the real cmd.RunSearch seam.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/lagz0ne/c3-design/cli/cmd"
	"github.com/lagz0ne/c3-design/cli/internal/content"
	"github.com/lagz0ne/c3-design/cli/internal/store"
)

const baselineSchemaV3 = "structural-retrieval-baseline.v3"

type V3BaselineCase struct {
	CaseID            string   `json:"case_id"`
	RowIDs            []string `json:"row_ids"`
	CanonicalRowBytes int      `json:"canonical_row_bytes"`
	ResultSHA256      string   `json:"result_sha256"`
	EntityFTSIDs      []string `json:"entity_fts_ids,omitempty"`
	ContentFTSIDs     []string `json:"content_fts_ids,omitempty"`
}

type V3BaselineArtifact struct {
	Schema                string           `json:"$schema"`
	Status                string           `json:"status"`
	Controller            string           `json:"controller"`
	FixtureSHA256         string           `json:"fixture_sha256"`
	BenchmarkSHA256       string           `json:"benchmark_sha256"`
	ScorerSHA256          string           `json:"scorer_sha256"`
	BenchmarkBaselineFile string           `json:"benchmark_baseline_file"`
	FixtureCount          int              `json:"fixture_count"`
	Cases                 []V3BaselineCase `json:"cases"`
	Privacy               V3PrivacyBinding `json:"privacy"`
	GenericOnly           bool             `json:"generic_only"`
	RawRowsRetained       bool             `json:"raw_rows_retained"`
	CandidateSource       bool             `json:"candidate_source"`
}

type v3ArmCapture struct {
	Rows   []cmd.SearchResultRow
	Probes V3DirectFTSMissWitness
}

// V3CandidateArtifact is the generic-only transport envelope for an explicit
// candidate response.  Keeping the source marker outside V3Response leaves
// the frozen scorer schema and ordinary response bytes unchanged.
type V3CandidateArtifact struct {
	Schema          string     `json:"$schema"`
	Status          string     `json:"status"`
	Response        V3Response `json:"response"`
	GenericOnly     bool       `json:"generic_only"`
	RawRowsRetained bool       `json:"raw_rows_retained"`
	CandidateSource bool       `json:"candidate_source"`
}

// CaptureCandidateV3 is the explicit, read-only candidate adapter.  It is not
// called by the baseline CLI and has no activation or effect authority.  Each
// fixture gets a disposable store and the reviewed projection capability is
// enabled directly at the cmd.RunSearch seam.  Only typed rows and route
// witnesses are returned; fixture/oracle labels and raw markdown stay outside
// the candidate envelope.
func CaptureCandidateV3(fixtures []V3Fixture) (V3Response, error) {
	if len(fixtures) == 0 {
		return V3Response{}, errors.New("no v3 fixtures")
	}
	seen := map[string]bool{}
	out := V3Response{Schema: ResponseSchemaV3, Cases: make([]V3CaseResponse, 0, len(fixtures))}
	for _, fixture := range fixtures {
		if err := ValidateFixture(fixture); err != nil {
			return V3Response{}, fmt.Errorf("%s: %w", fixture.CaseID, err)
		}
		if seen[fixture.CaseID] {
			return V3Response{}, fmt.Errorf("duplicate fixture %s", fixture.CaseID)
		}
		seen[fixture.CaseID] = true
		capture, err := captureCandidateV3Case(fixture)
		if err != nil {
			// The scorer explicitly permits omit/flagged for no-target cases.  A
			// projection that found no owner target is therefore represented as a
			// generic flagged case, while every other error remains fatal.
			if !OwnerMetricsApplicable(fixture) && isNoTargetProjectionError(err) {
				out.Cases = append(out.Cases, V3CaseResponse{CaseID: fixture.CaseID, Disposition: "flagged", GenericReason: "structural projection found no target"})
				continue
			}
			return V3Response{}, fmt.Errorf("candidate capture %s: %w", fixture.CaseID, err)
		}
		out.Cases = append(out.Cases, capture)
	}
	return out, nil
}

// CaptureCandidateV3Artifact returns the same response as CaptureCandidateV3
// with the explicit candidate-source and privacy markers required by retained
// generic run artifacts.
func CaptureCandidateV3Artifact(fixtures []V3Fixture) (V3CandidateArtifact, error) {
	response, err := CaptureCandidateV3(fixtures)
	if err != nil {
		return V3CandidateArtifact{}, err
	}
	return V3CandidateArtifact{Schema: "structural-retrieval-candidate.v3", Status: "candidate", Response: response, GenericOnly: true, RawRowsRetained: false, CandidateSource: true}, nil
}

func isNoTargetProjectionError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "no structurally witnessed owner target")
}

func captureCandidateV3Case(f V3Fixture) (V3CaseResponse, error) {
	root, err := os.MkdirTemp("", "c3-v3-candidate-")
	if err != nil {
		return V3CaseResponse{}, err
	}
	defer os.RemoveAll(root)
	dbPath := filepath.Join(root, "db", "c3.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return V3CaseResponse{}, err
	}
	s, err := store.Open(dbPath)
	if err != nil {
		return V3CaseResponse{}, err
	}
	defer s.Close()
	if err := s.WithTx(func(ts *store.Store) error {
		if err := insertV3Entities(ts, f.Entities); err != nil {
			return err
		}
		for _, entity := range f.Entities {
			if err := content.WriteEntity(ts, entity.ID, syntheticMarkdown(entity, f)); err != nil {
				return err
			}
		}
		for _, rel := range f.Relationships {
			if err := ts.AddRelationship(&store.Relationship{FromID: rel.FromID, ToID: rel.ToID, RelType: rel.RelType}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return V3CaseResponse{}, err
	}
	var raw bytes.Buffer
	if err := runCandidateSearch(s, f.Query, filepath.Join(root, "project"), filepath.Join(root, "c3"), &raw); err != nil {
		return V3CaseResponse{}, fmt.Errorf("RunSearch %s: %w", f.CaseID, err)
	}
	var output cmd.SearchOutput
	dec := json.NewDecoder(bytes.NewReader(raw.Bytes()))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&output); err != nil {
		return V3CaseResponse{}, fmt.Errorf("decode RunSearch %s: %w", f.CaseID, err)
	}
	return candidateCaseFromOutput(f.CaseID, output)
}

func candidateCaseFromOutput(caseID string, output cmd.SearchOutput) (V3CaseResponse, error) {
	rows, err := redactCandidateRows(output.Results)
	if err != nil {
		return V3CaseResponse{}, err
	}
	witnesses, err := convertCandidateRouteWitnesses(output.RouteWitnesses)
	if err != nil {
		return V3CaseResponse{}, err
	}
	return V3CaseResponse{CaseID: caseID, Rows: rows, Disposition: "omit", RouteWitnesses: witnesses}, nil
}

func redactCandidateRows(input []cmd.SearchResultRow) ([]cmd.SearchResultRow, error) {
	rows := make([]cmd.SearchResultRow, 0, len(input))
	seen := make(map[string]bool, len(input))
	for _, row := range input {
		if strings.TrimSpace(row.ID) == "" {
			return nil, errors.New("candidate response contains an empty row id")
		}
		if seen[row.ID] {
			return nil, fmt.Errorf("candidate response contains duplicate row id %s", row.ID)
		}
		seen[row.ID] = true
		// Search titles/snippets/context can contain project markdown.  The v3
		// scorer uses IDs, match sources, and route fields only, so clear those
		// fields at the privacy boundary instead of retaining raw content.
		row.Title = ""
		row.Snippet = ""
		row.Context = cmd.SearchContext{}
		row.MatchSources = append([]string(nil), row.MatchSources...)
		rows = append(rows, row)
	}
	return rows, nil
}

func convertCandidateRouteWitnesses(input []cmd.SearchRouteWitness) ([]V3RouteResponseWitness, error) {
	if len(input) == 0 {
		return nil, nil
	}
	out := make([]V3RouteResponseWitness, 0, len(input))
	seen := make(map[string]bool, len(input))
	for _, witness := range input {
		if witness.EntityID == "" || seen[witness.EntityID] {
			if witness.EntityID == "" {
				return nil, errors.New("candidate route witness has empty entity id")
			}
			return nil, fmt.Errorf("candidate route witness has duplicate entity id %s", witness.EntityID)
		}
		if len(witness.EntityContentIDs) != 1 || len(witness.DirectFTSContentMissIDs) != 1 {
			return nil, fmt.Errorf("candidate route witness %s is not one-to-one", witness.EntityID)
		}
		seen[witness.EntityID] = true
		out = append(out, V3RouteResponseWitness{
			EntityID: witness.EntityID, EntityContentID: witness.EntityContentIDs[0], MatchSource: witness.MatchSource,
			GraphFromID: witness.GraphFromID, GraphRelType: witness.GraphRelType, GraphToID: witness.GraphToID,
			DirectFTSEntityMissID: witness.DirectFTSEntityMissID, DirectFTSContentMissID: witness.DirectFTSContentMissIDs[0],
			RouteFieldValues: candidateRouteFieldValues(witness.RouteFieldValues),
		})
	}
	return out, nil
}

func candidateRouteFieldValues(route cmd.RouteEnrichment) map[string]any {
	values := make(map[string]any)
	if len(route.Facts) > 0 {
		values["facts"] = append([]string(nil), route.Facts...)
	}
	if len(route.Graph) > 0 {
		values["graph"] = append([]string(nil), route.Graph...)
	}
	if len(route.Anchors) > 0 {
		values["anchors"] = append([]string(nil), route.Anchors...)
	}
	if len(route.Lanes) > 0 {
		values["lanes"] = append([]string(nil), route.Lanes...)
	}
	if len(route.Drift) > 0 {
		values["drift"] = append([]string(nil), route.Drift...)
	}
	if route.Hash != "" {
		values["hash"] = route.Hash
	}
	if route.HashBasis != "" {
		values["hash_basis"] = route.HashBasis
	}
	return values
}

var captureFormatMu sync.Mutex

// CaptureFreshBaseline runs each fixture in an isolated disposable C3 store.
// It does not modify the fixture or any frozen v2 path. The generated entity
// text is synthetic only and is derived from the fixture query/ids when the
// fixture intentionally omits a body.
func CaptureFreshBaseline(fixtures []V3Fixture) (V3BaselineArtifact, error) {
	if len(fixtures) == 0 {
		return V3BaselineArtifact{}, errors.New("no v3 fixtures")
	}
	seen := map[string]bool{}
	out := V3BaselineArtifact{Schema: baselineSchemaV3, Status: "fresh", Controller: "unchanged-c3-cmd.RunSearch", FixtureCount: len(fixtures), GenericOnly: true, RawRowsRetained: false, CandidateSource: false}
	for _, fixture := range fixtures {
		if err := ValidateFixture(fixture); err != nil {
			return V3BaselineArtifact{}, fmt.Errorf("%s: %w", fixture.CaseID, err)
		}
		if seen[fixture.CaseID] {
			return V3BaselineArtifact{}, fmt.Errorf("duplicate fixture %s", fixture.CaseID)
		}
		seen[fixture.CaseID] = true
		capture, err := captureV3Case(fixture)
		if err != nil {
			return V3BaselineArtifact{}, err
		}
		if len(capture.Rows) == 0 || canonicalBytes(capture.Rows) <= 0 {
			return V3BaselineArtifact{}, fmt.Errorf("%s: unchanged C3 returned no usable rows", fixture.CaseID)
		}
		resultBytes, err := json.Marshal(struct {
			CaseID string                `json:"case_id"`
			Rows   []cmd.SearchResultRow `json:"rows"`
		}{fixture.CaseID, capture.Rows})
		if err != nil {
			return V3BaselineArtifact{}, err
		}
		out.Cases = append(out.Cases, V3BaselineCase{
			CaseID: fixture.CaseID, RowIDs: rowIDs(capture.Rows), CanonicalRowBytes: canonicalBytes(capture.Rows),
			ResultSHA256: sha256Bytes(resultBytes), EntityFTSIDs: append([]string(nil), capture.Probes.ProbeEntityHitIDs...), ContentFTSIDs: append([]string(nil), capture.Probes.ProbeContentHitIDs...),
		})
	}
	sort.Slice(out.Cases, func(i, j int) bool { return out.Cases[i].CaseID < out.Cases[j].CaseID })
	return out, nil
}

func captureV3Case(f V3Fixture) (v3ArmCapture, error) {
	root, err := os.MkdirTemp("", "c3-v3-b-baseline-")
	if err != nil {
		return v3ArmCapture{}, err
	}
	defer os.RemoveAll(root)
	dbPath := filepath.Join(root, "db", "c3.db")
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o700); err != nil {
		return v3ArmCapture{}, err
	}
	s, err := store.Open(dbPath)
	if err != nil {
		return v3ArmCapture{}, err
	}
	defer s.Close()
	if err := s.WithTx(func(ts *store.Store) error {
		if err := insertV3Entities(ts, f.Entities); err != nil {
			return err
		}
		for _, e := range f.Entities {
			if err := content.WriteEntity(ts, e.ID, syntheticMarkdown(e, f)); err != nil {
				return err
			}
		}
		for _, rel := range f.Relationships {
			if err := ts.AddRelationship(&store.Relationship{FromID: rel.FromID, ToID: rel.ToID, RelType: rel.RelType}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return v3ArmCapture{}, err
	}
	var raw bytes.Buffer
	if err := runCaptureSearch(s, f.Query, filepath.Join(root, "project"), filepath.Join(root, "c3"), &raw); err != nil {
		return v3ArmCapture{}, fmt.Errorf("RunSearch %s: %w", f.CaseID, err)
	}
	var output cmd.SearchOutput
	dec := json.NewDecoder(bytes.NewReader(raw.Bytes()))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&output); err != nil {
		return v3ArmCapture{}, fmt.Errorf("decode RunSearch %s: %w", f.CaseID, err)
	}
	entityHits, err := s.SearchWithLimit(f.Query, "", 20)
	if err != nil {
		return v3ArmCapture{}, err
	}
	contentHits, err := s.SearchContent(f.Query, 20)
	if err != nil {
		return v3ArmCapture{}, err
	}
	probes := V3DirectFTSMissWitness{}
	for _, hit := range entityHits {
		probes.ProbeEntityHitIDs = append(probes.ProbeEntityHitIDs, hit.ID)
	}
	for _, hit := range contentHits {
		probes.ProbeContentHitIDs = append(probes.ProbeContentHitIDs, hit.ID)
	}
	return v3ArmCapture{Rows: output.Results, Probes: probes}, nil
}

func runCaptureSearch(s *store.Store, query, projectDir, c3Dir string, raw *bytes.Buffer) error {
	return runJSONSearch(s, query, projectDir, c3Dir, raw, false, false)
}

func runCandidateSearch(s *store.Store, query, projectDir, c3Dir string, raw *bytes.Buffer) error {
	return runJSONSearch(s, query, projectDir, c3Dir, raw, true, true)
}

func runJSONSearch(s *store.Store, query, projectDir, c3Dir string, raw *bytes.Buffer, structuralProjection, captureProvenance bool) error {
	// The capture contract is JSON even when the caller launched the surrounding
	// process with C3X_MODE=agent. The command's public agent mode intentionally
	// emits TOON, so isolate and restore the ambient mode around this internal
	// controller call.
	captureFormatMu.Lock()
	defer captureFormatMu.Unlock()
	old, had := os.LookupEnv("C3X_MODE")
	_ = os.Unsetenv("C3X_MODE")
	defer func() {
		if had {
			_ = os.Setenv("C3X_MODE", old)
		} else {
			_ = os.Unsetenv("C3X_MODE")
		}
	}()
	return cmd.RunSearch(cmd.SearchOptions{Store: s, Query: query, JSON: true, Limit: 5, NoSemantic: true, ProjectDir: projectDir, C3Dir: c3Dir, StructuralProjection: structuralProjection, CaptureProvenance: captureProvenance}, raw)
}

func insertV3Entities(s *store.Store, entities []V3Entity) error {
	pending := append([]V3Entity(nil), entities...)
	inserted := map[string]bool{}
	for len(pending) > 0 {
		next := pending[:0]
		progress := false
		for _, e := range pending {
			if parent := parentID(e); parent != "" && !inserted[parent] {
				next = append(next, e)
				continue
			}
			id := e.ID
			title := e.Title
			if title == "" {
				title = id
			}
			slug := e.Slug
			if slug == "" {
				slug = strings.ToLower(strings.ReplaceAll(id, " ", "-"))
			}
			metadata := e.Metadata
			if metadata == "" {
				metadata = "{}"
			}
			if err := s.InsertEntity(&store.Entity{ID: id, Type: e.Type, Title: title, Slug: slug, ParentID: parentID(e), Status: "active", Metadata: metadata}); err != nil {
				return err
			}
			inserted[id] = true
			progress = true
		}
		if !progress {
			return errors.New("v3 entity parent graph has a cycle or missing parent")
		}
		pending = append([]V3Entity(nil), next...)
	}
	return nil
}

// parentID reads only the optional generic containment key from entity
// metadata.  The v3 fixture contract keeps metadata as "{}", so this is a
// no-op for every frozen v3 artifact while allowing a disposable v4 fixture to
// express one-hop containment without adding a scorer field or project
// semantics. Malformed/other metadata is deliberately ignored here; the
// ordinary entity row still carries its original metadata unchanged.
func parentID(e V3Entity) string {
	if strings.TrimSpace(e.Metadata) == "" || strings.TrimSpace(e.Metadata) == "{}" {
		return ""
	}
	var metadata struct {
		ParentID string `json:"parent_id"`
	}
	if err := json.Unmarshal([]byte(e.Metadata), &metadata); err != nil {
		return ""
	}
	return strings.TrimSpace(metadata.ParentID)
}

func syntheticMarkdown(e V3Entity, f V3Fixture) string {
	if strings.TrimSpace(e.Markdown) != "" {
		return e.Markdown
	}
	var facts []string
	for _, fact := range f.Facts {
		if fact.EntityID == e.ID {
			facts = append(facts, fact.ID)
		}
	}
	return fmt.Sprintf("# %s\n\n%s\n\n%s", e.ID, f.Query, strings.Join(facts, " "))
}

func rowIDs(rows []cmd.SearchResultRow) []string {
	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	return ids
}
func sha256Bytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func LoadV3FixtureFile(path string) ([]V3Fixture, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	var fixtures []V3Fixture
	if err := json.Unmarshal(data, &fixtures); err != nil {
		return nil, "", err
	}
	for _, f := range fixtures {
		if err := ValidateFixture(f); err != nil {
			return nil, "", fmt.Errorf("%s: %w", f.CaseID, err)
		}
	}
	return fixtures, sha256Bytes(data), nil
}

func LoadV3BenchmarkFile(path string) (V3Benchmark, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return V3Benchmark{}, "", err
	}
	var b V3Benchmark
	if err := json.Unmarshal(data, &b); err != nil {
		return V3Benchmark{}, "", err
	}
	if err := ValidateBenchmark(b); err != nil {
		return V3Benchmark{}, "", err
	}
	return b, sha256Bytes(data), nil
}

func WriteV3Baseline(path string, artifact V3BaselineArtifact) error {
	if artifact.Status != "fresh" || !artifact.GenericOnly || artifact.RawRowsRetained || artifact.CandidateSource {
		return errors.New("baseline artifact is not fresh generic B-v3")
	}
	if !validSHA256(artifact.FixtureSHA256) || !validSHA256(artifact.BenchmarkSHA256) || !validSHA256(artifact.ScorerSHA256) {
		return errors.New("baseline artifact source bindings are incomplete")
	}
	if !validSHA256(artifact.Privacy.PolicySHA256) || !validSHA256(artifact.Privacy.ScannerSourceSHA256) || !validSHA256(artifact.Privacy.DetectorDefinitionSHA256) || len(artifact.Privacy.ScanScope) == 0 {
		return errors.New("baseline artifact privacy binding is incomplete")
	}
	if len(artifact.Cases) != artifact.FixtureCount {
		return errors.New("baseline case count mismatch")
	}
	seen := map[string]bool{}
	for _, c := range artifact.Cases {
		if c.CaseID == "" || seen[c.CaseID] || c.CanonicalRowBytes <= 0 || !validSHA256(c.ResultSHA256) {
			return fmt.Errorf("invalid baseline case %s", c.CaseID)
		}
		seen[c.CaseID] = true
	}
	data, err := json.MarshalIndent(artifact, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	if _, err := file.Write(data); err != nil {
		return err
	}
	return file.Close()
}
