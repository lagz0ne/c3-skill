package cmd

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/eval"
	"github.com/lagz0ne/c3-design/cli/internal/store"
	"gopkg.in/yaml.v3"
)

// EvalOptions holds parameters for the eval command.
type EvalOptions struct {
	Store      *store.Store
	ProjectDir string
	C3Dir      string
	JSON       bool
	Only       string // optional: run a single fact's spec
	Policy     bool   // report cache trust policy instead of running verdicts
}

// EvalReport is the output of a c3 eval run.
type EvalReport struct {
	Total          int            `json:"total"`
	Holds          int            `json:"holds"`
	Drift          int            `json:"drift"`
	NeedsJudgement int            `json:"needs_judgement"`
	Verdicts       []eval.Verdict `json:"verdicts"`
}

type compactEvalReport struct {
	Total          int            `json:"total"`
	Holds          int            `json:"holds"`
	Drift          int            `json:"drift"`
	NeedsJudgement int            `json:"needs_judgement"`
	Verdicts       []eval.Verdict `json:"verdicts,omitempty"`
}

// EvalPolicyReport is an explicit diagnostic for cache trust coverage. Routine
// eval output must not carry this data; ask for it with c3x eval --policy.
type EvalPolicyReport struct {
	Total       int              `json:"total"`
	Reusable    int              `json:"reusable"`
	Cacheable   int              `json:"cacheable"`
	Live        int              `json:"live"`
	Judgement   int              `json:"judgement"`
	Unsupported int              `json:"unsupported"`
	Specs       []EvalPolicySpec `json:"specs,omitempty"`
}

type EvalPolicySpec struct {
	Fact      string `json:"fact"`
	Class     string `json:"class"`
	Cacheable bool   `json:"cacheable"`
	Reason    string `json:"reason,omitempty"`
}

// LoadEvalSpecs reads every .c3/eval/<fact>.yaml pipeline. The fact→code binding
// lives in the specs (the `code:` field), so this is also the lookup index.
func LoadEvalSpecs(c3Dir string) ([]eval.Spec, error) {
	dir := filepath.Join(c3Dir, "eval")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read eval dir: %w", err)
	}
	var specs []eval.Spec
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".yaml") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, ent.Name()))
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", ent.Name(), err)
		}
		var sp eval.Spec
		if err := yaml.Unmarshal(b, &sp); err != nil {
			return nil, fmt.Errorf("parse %s: %w", ent.Name(), err)
		}
		sp.SpecHash = shortHash(b)
		if sp.Fact == "" {
			sp.Fact = strings.TrimSuffix(ent.Name(), ".yaml")
		}
		specs = append(specs, sp)
	}
	sort.Slice(specs, func(i, j int) bool { return specs[i].Fact < specs[j].Fact })
	return specs, nil
}

// EvalBindings is the fact→code map derived from the specs (replaces code-map.yaml).
func EvalBindings(specs []eval.Spec) codemap.CodeMap {
	b := codemap.CodeMap{}
	for _, sp := range specs {
		if len(sp.Code) > 0 {
			b[sp.Fact] = sp.Code
		}
	}
	return b
}

// RunEval runs every .c3/eval/<fact>.yaml conformance pipeline against the live
// external and reports a one-off verdict per fact. It is never a gate.
func RunEval(opts EvalOptions, w io.Writer) error {
	specs, err := LoadEvalSpecs(opts.C3Dir)
	if err != nil {
		return err
	}
	if len(specs) == 0 {
		return fmt.Errorf("error: no eval specs at %s\nhint: author .c3/eval/<fact>.yaml pipelines", filepath.Join(opts.C3Dir, "eval"))
	}

	ents, err := opts.Store.AllEntities()
	if err != nil {
		return fmt.Errorf("list entities: %w", err)
	}
	ids := make([]string, 0, len(ents))
	for _, e := range ents {
		ids = append(ids, e.ID)
	}
	roots := factRootsByID(ents)
	eng := &eval.Engine{ProjectDir: opts.ProjectDir, CodeBindings: EvalBindings(specs), FactIDs: ids}
	if opts.Policy {
		return writeEvalPolicy(opts, specs, eng, w)
	}

	rep := EvalReport{Verdicts: []eval.Verdict{}}
	records := make([]store.EvalMatchRecord, 0, len(specs))
	for _, sp := range specs {
		if opts.Only != "" && sp.Fact != opts.Only {
			continue
		}
		probe, cacheable := eng.ReusableManifest(sp)
		v, reused := cachedEvalVerdict(opts.Store, sp, roots[sp.Fact], probe, cacheable)
		if !reused {
			v = eng.Run(sp)
		}
		records = append(records, evalMatchRecord(sp, v, roots[sp.Fact], probe, cacheable))
		rep.Verdicts = append(rep.Verdicts, v)
		rep.Total++
		switch v.Verdict {
		case "holds":
			rep.Holds++
		case "drift":
			rep.Drift++
		case "needs-judgement":
			rep.NeedsJudgement++
		}
	}
	if err := saveEvalMatches(opts.Store, records); err != nil {
		return err
	}
	format := ResolveFormat(opts.JSON, isAgentMode())
	if format == FormatJSON {
		return WriteObjectOutput(w, rep, format, evalHelpHints())
	}
	compact := compactEvalForAgent(rep, opts.Only != "")
	var hints []HelpHint
	if compact.Drift > 0 || compact.NeedsJudgement > 0 {
		hints = evalHelpHints()
	}
	return WriteObjectOutput(w, compact, format, hints)
}

func writeEvalPolicy(opts EvalOptions, specs []eval.Spec, eng *eval.Engine, w io.Writer) error {
	report := EvalPolicyReport{}
	focused := opts.Only != ""
	for _, sp := range specs {
		if focused && sp.Fact != opts.Only {
			continue
		}
		policy := eval.ReusePolicyForSpec(sp)
		cacheable := false
		if policy.Class == eval.ReuseDeterministic {
			_, cacheable = eng.ReusableManifest(sp)
		}
		report.Total++
		if cacheable {
			report.Cacheable++
		}
		switch policy.Class {
		case eval.ReuseDeterministic:
			report.Reusable++
		case eval.ReuseLivePolicy:
			report.Live++
		case eval.ReuseJudgement:
			report.Judgement++
		default:
			report.Unsupported++
		}
		if focused {
			report.Specs = append(report.Specs, EvalPolicySpec{
				Fact:      sp.Fact,
				Class:     string(policy.Class),
				Cacheable: cacheable,
				Reason:    policy.Reason,
			})
		}
	}
	return WriteObjectOutput(w, report, ResolveFormat(opts.JSON, isAgentMode()), nil)
}

func cachedEvalVerdict(s *store.Store, spec eval.Spec, factRoot string, probe eval.ManifestProbe, cacheable bool) (eval.Verdict, bool) {
	if !cacheable {
		return eval.Verdict{}, false
	}
	cached, err := s.EvalMatch(spec.Fact)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return eval.Verdict{}, false
		}
		return eval.Verdict{}, false
	}
	if cached.FactRoot != factRoot || cached.EvalSpecHash != spec.SpecHash || cached.Verdict == "" {
		return eval.Verdict{}, false
	}
	if !sameEvalUnits(cached.CacheUnits, probe.Manifest) {
		return eval.Verdict{}, false
	}
	return eval.Verdict{
		Fact:          spec.Fact,
		Claim:         spec.Claim,
		Verdict:       cached.Verdict,
		ExternalState: cached.ExternalState,
		Evidence:      cached.Evidence,
		Manifest:      probe.Manifest,
	}, true
}

func sameEvalUnits(cached []store.EvalMatchUnit, current []eval.Unit) bool {
	if len(cached) != len(current) {
		return false
	}
	for i := range cached {
		if cached[i].Kind != current[i].Kind ||
			cached[i].Key != current[i].Key ||
			cached[i].Digest != current[i].Digest ||
			cached[i].Bytes != current[i].Bytes {
			return false
		}
	}
	return true
}

func saveEvalMatches(s *store.Store, records []store.EvalMatchRecord) error {
	if len(records) == 0 {
		return nil
	}
	return s.WithTx(func(ts *store.Store) error {
		for _, record := range records {
			if err := ts.SaveEvalMatch(record); err != nil {
				return fmt.Errorf("save eval match %s: %w", record.Fact, err)
			}
		}
		return nil
	})
}

func shortHash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])[:12]
}

func factRootsByID(entities []*store.Entity) map[string]string {
	roots := make(map[string]string, len(entities))
	for _, entity := range entities {
		roots[entity.ID] = entity.RootMerkle
	}
	return roots
}

func evalMatchRecord(spec eval.Spec, verdict eval.Verdict, factRoot string, probe eval.ManifestProbe, cacheable bool) store.EvalMatchRecord {
	units := make([]store.EvalMatchUnit, len(verdict.Manifest))
	for i, unit := range verdict.Manifest {
		units[i] = store.EvalMatchUnit{
			Kind:   unit.Kind,
			Key:    unit.Key,
			Digest: unit.Digest,
			Bytes:  unit.Bytes,
		}
	}
	cacheUnits := units
	if cacheable {
		cacheUnits = make([]store.EvalMatchUnit, len(probe.Manifest))
		for i, unit := range probe.Manifest {
			cacheUnits[i] = store.EvalMatchUnit{
				Kind:   unit.Kind,
				Key:    unit.Key,
				Digest: unit.Digest,
				Bytes:  unit.Bytes,
			}
		}
	}
	return store.EvalMatchRecord{
		Fact:          verdict.Fact,
		Claim:         verdict.Claim,
		FactRoot:      factRoot,
		EvalSpecHash:  spec.SpecHash,
		ExternalState: verdict.ExternalState,
		Verdict:       verdict.Verdict,
		Evidence:      verdict.Evidence,
		Units:         units,
		CacheUnits:    cacheUnits,
	}
}

func evalHelpHints() []HelpHint {
	return []HelpHint{
		{Command: "c3x eval <fact-id>", Description: "rerun a single conformance spec after inspecting a drift or judgement row"},
		{Command: "c3x lookup <file-or-glob>", Description: "check which fact owns a code path through the same eval-spec bindings"},
		{Command: "c3x read <fact-id>", Description: "read the frozen claim before changing its mutable eval lens"},
	}
}

func compactEvalForAgent(rep EvalReport, includeHolds bool) compactEvalReport {
	out := compactEvalReport{
		Total:          rep.Total,
		Holds:          rep.Holds,
		Drift:          rep.Drift,
		NeedsJudgement: rep.NeedsJudgement,
	}
	for _, verdict := range rep.Verdicts {
		if includeHolds || verdict.Verdict != "holds" {
			out.Verdicts = append(out.Verdicts, verdict)
		}
	}
	return out
}
