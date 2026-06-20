package cmd

import (
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
}

// EvalReport is the output of a c3 eval run.
type EvalReport struct {
	Total          int            `json:"total"`
	Holds          int            `json:"holds"`
	Drift          int            `json:"drift"`
	NeedsJudgement int            `json:"needs_judgement"`
	Verdicts       []eval.Verdict `json:"verdicts"`
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
	eng := &eval.Engine{ProjectDir: opts.ProjectDir, CodeBindings: EvalBindings(specs), FactIDs: ids}

	rep := EvalReport{Verdicts: []eval.Verdict{}}
	for _, sp := range specs {
		if opts.Only != "" && sp.Fact != opts.Only {
			continue
		}
		v := eng.Run(sp)
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
	return WriteObjectOutput(w, rep, ResolveFormat(opts.JSON, isAgentMode()), nil)
}
