// Package eval runs conformance pipelines: a frozen fact's claim checked against
// an uncontrolled external, expressed as a composition of five ops —
// gather · filter · transform · eval · loop. A run produces a one-off Verdict
// stamped with the external state it measured; it is never a gate.
package eval

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
)

// Spec is one fact's eval pipeline, loaded from .c3/eval/<fact>.yaml.
type Spec struct {
	Fact  string   `yaml:"fact"`
	Claim string   `yaml:"claim"`
	Code  []string `yaml:"code"` // the fact→code binding (lookup + a default resolve check)
	Pipeline []Step `yaml:"pipeline"`
}

// Step is exactly one of the five ops (the others nil).
type Step struct {
	Gather    *Gather    `yaml:"gather"`
	Filter    *Filter    `yaml:"filter"`
	Transform *Transform `yaml:"transform"`
	Eval      *Eval      `yaml:"eval"`
	Loop      *Loop      `yaml:"loop"`
}

// Gather acquires data from a source — raw (read) or mechanical (exec).
type Gather struct {
	File    string   `yaml:"file"`    // raw: read a file → one value (its content)
	Command string   `yaml:"command"` // mechanical: exec sh -c at repo root → stdout lines
	Files   string   `yaml:"files"`   // glob → matching file paths
	Facts   string   `yaml:"facts"`   // fact-id glob → ids (for loop.over)
	Code    string   `yaml:"code"`    // a fact-id (or $item) → its declared code globs → files
	Literal []string `yaml:"literal"` // these exact values as the stream (for loop.over)
	Each    []Gather `yaml:"each"`    // several sub-gathers, concatenated
}

// Filter keeps the values matching a predicate.
type Filter struct {
	Contains string `yaml:"contains"`
	Matches  string `yaml:"matches"`
}

// Transform reshapes each value.
type Transform struct {
	Trim  bool `yaml:"trim"`
	First bool `yaml:"first"`
	Lines bool `yaml:"lines"`
}

// Eval asserts on the stream and produces the verdict (the terminal op).
type Eval struct {
	Exists      bool     `yaml:"exists"`       // every value is an existing file path
	Equals      string   `yaml:"equals"`       // every value equals this literal
	AllEqual    bool     `yaml:"all_equal"`    // all values equal each other
	ContainsAll []string `yaml:"contains_all"` // the joined stream contains every substring
	Contains    string   `yaml:"contains"`     // the joined stream contains this substring
	Count       string   `yaml:"count"`        // value count satisfies e.g. "> 0", "== 6"
	Judgement   string   `yaml:"judgement"`    // a rubric question → needs-judgement
}

// Loop fans a sub-pipeline over each item of Over, binding $item.
type Loop struct {
	Over Gather `yaml:"over"`
	Do   []Step `yaml:"do"`
}

// Verdict is the stamped result of one eval run.
type Verdict struct {
	Fact          string   `json:"fact"`
	Claim         string   `json:"claim,omitempty"`
	Verdict       string   `json:"verdict"` // holds | drift | needs-judgement
	ExternalState string   `json:"external_state"`
	Evidence      []string `json:"evidence,omitempty"`
}

// Engine carries the run context shared by every spec.
type Engine struct {
	ProjectDir   string          // repo root (the dir holding .c3) — gather cwd
	CodeBindings codemap.CodeMap // fact-id → declared code globs, from the eval-specs
	FactIDs      []string        // every entity id, for gather: facts
}

// Run evaluates one spec to a Verdict. A spec with `code:` and no pipeline gets a
// default per-glob resolve check (every declared glob must resolve to ≥1 file).
func (e *Engine) Run(spec Spec) Verdict {
	pipeline := spec.Pipeline
	if len(pipeline) == 0 && len(spec.Code) > 0 {
		pipeline = resolvePipeline(spec.Code)
	}
	return e.run(spec, pipeline, "")
}

func resolvePipeline(globs []string) []Step {
	return []Step{{Loop: &Loop{
		Over: Gather{Literal: globs},
		Do:   []Step{{Gather: &Gather{Files: "$item"}}, {Eval: &Eval{Exists: true}}},
	}}}
}

func (e *Engine) run(spec Spec, pipeline []Step, item string) Verdict {
	var frame []string
	for _, st := range pipeline {
		switch {
		case st.Gather != nil:
			vals, err := e.gather(*st.Gather, item)
			if err != nil {
				return drift(spec, "gather: "+err.Error())
			}
			frame = vals
		case st.Filter != nil:
			frame = applyFilter(frame, *st.Filter)
		case st.Transform != nil:
			frame = applyTransform(frame, *st.Transform)
		case st.Eval != nil:
			return e.assert(spec, frame, *st.Eval, item)
		case st.Loop != nil:
			return e.loop(spec, *st.Loop)
		}
	}
	return drift(spec, "pipeline has no terminal eval/loop")
}

func (e *Engine) loop(spec Spec, l Loop) Verdict {
	items, err := e.gather(l.Over, "")
	if err != nil {
		return drift(spec, "loop.over: "+err.Error())
	}
	verdict, evidence := "holds", []string{}
	for _, item := range items {
		v := e.run(spec, l.Do, item)
		evidence = append(evidence, fmt.Sprintf("%s → %s", item, v.Verdict))
		switch v.Verdict {
		case "drift":
			verdict = "drift"
		case "needs-judgement":
			if verdict == "holds" {
				verdict = "needs-judgement"
			}
		}
	}
	if len(items) == 0 {
		return drift(spec, "loop.over produced no items")
	}
	return Verdict{Fact: spec.Fact, Claim: spec.Claim, Verdict: verdict, ExternalState: stamp(items), Evidence: evidence}
}

func (e *Engine) gather(g Gather, item string) ([]string, error) {
	switch {
	case len(g.Each) > 0:
		var out []string
		for _, sub := range g.Each {
			vals, err := e.gather(sub, item)
			if err != nil {
				return nil, err
			}
			out = append(out, vals...)
		}
		return out, nil
	case g.File != "":
		b, err := os.ReadFile(filepath.Join(e.ProjectDir, subst(g.File, item)))
		if err != nil {
			return nil, err
		}
		return []string{strings.TrimRight(string(b), "\n")}, nil
	case g.Command != "":
		return e.exec(subst(g.Command, item))
	case g.Files != "":
		return codemap.GlobFiles(os.DirFS(e.ProjectDir), subst(g.Files, item))
	case g.Facts != "":
		return matchIDs(e.FactIDs, subst(g.Facts, item)), nil
	case len(g.Literal) > 0:
		out := make([]string, len(g.Literal))
		for i, v := range g.Literal {
			out[i] = subst(v, item)
		}
		return out, nil
	case g.Code != "":
		id := subst(g.Code, item)
		var files []string
		for _, glob := range e.CodeBindings[id] {
			fs, err := codemap.GlobFiles(os.DirFS(e.ProjectDir), glob)
			if err != nil {
				return nil, err
			}
			files = append(files, fs...)
		}
		sort.Strings(files)
		return files, nil
	}
	return nil, fmt.Errorf("empty gather")
}

func (e *Engine) exec(command string) ([]string, error) {
	c := exec.Command("sh", "-c", command)
	c.Dir = e.ProjectDir
	out, err := c.Output()
	if err != nil {
		// A non-zero exit is not a failure — grep/find/jq routinely exit 1 on
		// "no match", which is a legitimately empty gather. Only a real exec
		// failure (command can't start) is an error; otherwise use the stdout.
		if _, isExit := err.(*exec.ExitError); !isExit {
			return nil, fmt.Errorf("command %q: %w", command, err)
		}
	}
	return splitLines(string(out)), nil
}

func (e *Engine) assert(spec Spec, frame []string, ev Eval, item string) Verdict {
	v := func(verdict string, evidence ...string) Verdict {
		return Verdict{Fact: spec.Fact, Claim: spec.Claim, Verdict: verdict, ExternalState: stamp(frame), Evidence: evidence}
	}
	switch {
	case ev.Judgement != "":
		return v("needs-judgement", "judge: "+subst(ev.Judgement, item), fmt.Sprintf("%d value(s) gathered", len(frame)))
	case ev.Exists:
		var missing []string
		for _, p := range frame {
			if _, err := os.Stat(filepath.Join(e.ProjectDir, p)); err != nil {
				missing = append(missing, p+" ✗")
			}
		}
		if len(frame) == 0 {
			return v("drift", "nothing to check exists")
		}
		if len(missing) > 0 {
			return v("drift", missing...)
		}
		return v("holds", fmt.Sprintf("%d path(s) exist", len(frame)))
	case ev.AllEqual:
		if len(uniq(frame)) <= 1 {
			return v("holds", "all equal: "+joinShort(frame))
		}
		return v("drift", "distinct: "+strings.Join(uniq(frame), " | "))
	case ev.Equals != "":
		for _, x := range frame {
			if strings.TrimSpace(x) != ev.Equals {
				return v("drift", fmt.Sprintf("%q ≠ %q", x, ev.Equals))
			}
		}
		return v("holds", "all == "+ev.Equals)
	case len(ev.ContainsAll) > 0:
		joined := strings.Join(frame, "\n")
		var miss []string
		for _, sub := range ev.ContainsAll {
			if !strings.Contains(joined, sub) {
				miss = append(miss, sub+" ✗")
			}
		}
		if len(miss) > 0 {
			return v("drift", miss...)
		}
		return v("holds", "contains all "+strconv.Itoa(len(ev.ContainsAll)))
	case ev.Contains != "":
		if strings.Contains(strings.Join(frame, "\n"), ev.Contains) {
			return v("holds", "contains "+ev.Contains)
		}
		return v("drift", "missing "+ev.Contains)
	case ev.Count != "":
		ok, desc := countOK(len(frame), ev.Count)
		if ok {
			return v("holds", desc)
		}
		return v("drift", desc)
	}
	return drift(spec, "eval has no predicate")
}

// --- pure helpers ---

func applyFilter(frame []string, f Filter) []string {
	var re *regexp.Regexp
	if f.Matches != "" {
		re = regexp.MustCompile(f.Matches)
	}
	out := frame[:0:0]
	for _, x := range frame {
		if f.Contains != "" && !strings.Contains(x, f.Contains) {
			continue
		}
		if re != nil && !re.MatchString(x) {
			continue
		}
		out = append(out, x)
	}
	return out
}

func applyTransform(frame []string, t Transform) []string {
	if t.Lines {
		var out []string
		for _, x := range frame {
			out = append(out, splitLines(x)...)
		}
		frame = out
	}
	if t.Trim {
		for i := range frame {
			frame[i] = strings.TrimSpace(frame[i])
		}
	}
	if t.First && len(frame) > 1 {
		frame = frame[:1]
	}
	return frame
}

func drift(spec Spec, why string) Verdict {
	return Verdict{Fact: spec.Fact, Claim: spec.Claim, Verdict: "drift", Evidence: []string{why}}
}

func subst(s, item string) string { return strings.ReplaceAll(s, "$item", item) }

func splitLines(s string) []string {
	var out []string
	for _, ln := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		if strings.TrimSpace(ln) != "" {
			out = append(out, ln)
		}
	}
	return out
}

func uniq(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, x := range in {
		k := strings.TrimSpace(x)
		if !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	return out
}

func matchIDs(ids []string, glob string) []string {
	var out []string
	for _, id := range ids {
		if ok, _ := filepath.Match(glob, id); ok {
			out = append(out, id)
		}
	}
	sort.Strings(out)
	return out
}

func countOK(n int, expr string) (bool, string) {
	f := strings.Fields(expr)
	if len(f) != 2 {
		return false, "bad count expr: " + expr
	}
	want, err := strconv.Atoi(f[1])
	if err != nil {
		return false, "bad count number: " + expr
	}
	got := false
	switch f[0] {
	case ">":
		got = n > want
	case ">=":
		got = n >= want
	case "<":
		got = n < want
	case "<=":
		got = n <= want
	case "==":
		got = n == want
	case "!=":
		got = n != want
	}
	return got, fmt.Sprintf("count %d %s", n, expr)
}

func joinShort(frame []string) string {
	if len(frame) == 0 {
		return "(empty)"
	}
	return strings.TrimSpace(frame[0])
}

func stamp(frame []string) string {
	h := sha256.Sum256([]byte(strings.Join(frame, "\x00")))
	return hex.EncodeToString(h[:])[:12]
}
