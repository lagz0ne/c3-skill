// Package eval runs conformance pipelines: a frozen fact's claim checked against
// an uncontrolled external, expressed as a composition of five ops —
// gather · filter · transform · eval · loop. A run produces a one-off Verdict
// stamped with the external state it measured; it is never a gate.
package eval

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/lagz0ne/c3-design/cli/internal/codemap"
	"github.com/lagz0ne/c3-design/cli/internal/content"
)

// Spec is one fact's eval pipeline, loaded from .c3/eval/<fact>.yaml.
type Spec struct {
	Fact     string   `yaml:"fact"`
	Claim    string   `yaml:"claim"`
	Code     []string `yaml:"code"` // the fact→code binding (lookup + a default resolve check)
	Pipeline []Step   `yaml:"pipeline"`
	SpecHash string   `yaml:"-"`
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
	File    string         `yaml:"file"`    // raw: read a file → one value (its content)
	Command string         `yaml:"command"` // mechanical: exec sh -c at repo root → stdout lines
	Inputs  []string       `yaml:"inputs"`  // command contract: file/glob inputs that fully determine stdout
	Files   string         `yaml:"files"`   // glob → matching file paths
	Facts   string         `yaml:"facts"`   // fact-id glob → ids (for loop.over)
	Code    string         `yaml:"code"`    // a fact-id (or $item) → its declared code globs → files
	Literal []string       `yaml:"literal"` // these exact values as the stream (for loop.over)
	Each    []Gather       `yaml:"each"`    // several sub-gathers, concatenated
	Outline *OutlineGather `yaml:"outline"` // ast-grep outline → structural code units
}

// OutlineGather extracts source structure through ast-grep outline. It is for
// claim-bearing units such as exported functions, structs, classes, and direct
// members, not full-body text.
type OutlineGather struct {
	Paths                 []string `yaml:"paths"`
	Lang                  string   `yaml:"lang"`
	Items                 string   `yaml:"items"`
	View                  string   `yaml:"view"`
	SymbolTypes           string   `yaml:"type"`
	Match                 string   `yaml:"match"`
	PubMembers            bool     `yaml:"pub_members"`
	Globs                 []string `yaml:"globs"`
	OutlineRules          []string `yaml:"outline_rules"`
	NoDefaultOutlineRules bool     `yaml:"no_default_outline_rules"`
}

// Filter keeps the values matching a predicate.
type Filter struct {
	Contains string `yaml:"contains"`
	Matches  string `yaml:"matches"`
}

// Transform reshapes each value.
type Transform struct {
	Trim       bool `yaml:"trim"`
	First      bool `yaml:"first"`
	Lines      bool `yaml:"lines"`
	CodeBlocks bool `yaml:"code_blocks"`
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
	Manifest      []Unit   `json:"-"`
}

// Unit is one refined matched surface after gather/transform. It is intentionally
// generic: transforms can refine a file to lines, code blocks, table rows, command
// rows, or any future typed unit without changing the verdict contract.
type Unit struct {
	Kind   string
	Key    string
	Digest string
	Bytes  int
}

// ManifestProbe is the current matched surface for a spec when it can be
// recomputed without running unsafe terminal work. It supports cache reuse:
// if this manifest equals the persisted manifest for the same fact root and
// spec hash, the previous deterministic verdict can be reused.
type ManifestProbe struct {
	ExternalState string
	Manifest      []Unit
}

// ReuseClass explains the cache trust class for a spec. It is structural: it
// says whether a result surface can be mechanically recomputed and trusted if
// the recomputed identity matches the persisted identity.
type ReuseClass string

const (
	ReuseDeterministic ReuseClass = "deterministic-reusable"
	ReuseLivePolicy    ReuseClass = "deterministic-live-by-policy"
	ReuseJudgement     ReuseClass = "non-deterministic-judgement"
	ReuseUnsupported   ReuseClass = "unsupported"
)

// ReusePolicy is the structural cache policy for one spec.
type ReusePolicy struct {
	Class  ReuseClass
	Reason string
}

type frameItem struct {
	kind  string
	key   string
	value string
}

const minimumAstGrepVersion = "0.44.0"

// Engine carries the run context shared by every spec.
type Engine struct {
	ProjectDir   string          // repo root (the dir holding .c3) — gather cwd
	CodeBindings codemap.CodeMap // fact-id → declared code globs, from the eval-specs
	FactIDs      []string        // every entity id, for gather: facts
}

// Run evaluates one spec to a Verdict. The code: binding is a precondition guard:
// it must resolve before any pipeline is trusted, so a vanished code surface drifts
// rather than letting a vacuous pipeline (e.g. count == 0 over nothing) read as holds.
func (e *Engine) Run(spec Spec) Verdict {
	// Guard: the declared code surface must resolve. If it does not, the fact has
	// drifted — and a downstream pipeline must not be trusted to pass vacuously: an
	// empty gather over a renamed/deleted path would otherwise read as "holds".
	if len(spec.Code) > 0 {
		guard := e.run(spec, resolvePipeline(spec.Code), "")
		if guard.Verdict != "holds" {
			return guard
		}
		if len(spec.Pipeline) == 0 {
			return guard
		}
	}
	if len(spec.Pipeline) == 0 {
		return drift(spec, "spec declares neither a code: binding nor a pipeline")
	}
	return e.run(spec, spec.Pipeline, "")
}

// ReusableManifest returns the current manifest only for deterministic specs
// whose terminal verdict may be reused after fact-root, spec-hash, and manifest
// identity have all matched. Command gathers and judgement terminals are
// intentionally excluded: they either have external state that cannot be skipped
// safely or require human/model judgement rather than a cached deterministic
// predicate. Loops are reusable only when their over-gather and child pipeline
// are also deterministic.
func (e *Engine) ReusableManifest(spec Spec) (ManifestProbe, bool) {
	if ReusePolicyForSpec(spec).Class != ReuseDeterministic {
		return ManifestProbe{}, false
	}
	if len(spec.Code) > 0 {
		guard, ok := e.codeGuardManifest(spec.Code)
		if !ok {
			return ManifestProbe{}, false
		}
		if len(spec.Pipeline) == 0 {
			return guard, true
		}
	}
	if len(spec.Pipeline) == 0 || !reusablePipeline(spec.Pipeline) {
		return ManifestProbe{}, false
	}
	if pipelineHasCommand(spec.Pipeline) {
		return e.dependencyManifest(spec.Pipeline, "")
	}
	return e.pipelineManifest(spec.Pipeline, "")
}

// ReusePolicyForSpec classifies the spec without looking at current files or
// command output. Missing files still force a live run because ReusableManifest
// cannot compute a matching manifest, but the structural trust class remains
// deterministic when the source type itself is mechanically bounded.
func ReusePolicyForSpec(spec Spec) ReusePolicy {
	if len(spec.Code) == 0 && len(spec.Pipeline) == 0 {
		return ReusePolicy{Class: ReuseUnsupported, Reason: "spec has no code binding or pipeline"}
	}
	if len(spec.Pipeline) == 0 {
		return ReusePolicy{Class: ReuseDeterministic, Reason: "code binding resolves to a deterministic path manifest"}
	}
	return reusePolicyForPipeline(spec.Pipeline)
}

func resolvePipeline(globs []string) []Step {
	return []Step{{Loop: &Loop{
		Over: Gather{Literal: globs},
		Do:   []Step{{Gather: &Gather{Files: "$item"}}, {Eval: &Eval{Exists: true}}},
	}}}
}

func (e *Engine) codeGuardManifest(globs []string) (ManifestProbe, bool) {
	var childStates []string
	childManifest := make([]Unit, 0, len(globs))
	for _, glob := range globs {
		files, err := codemap.GlobFiles(os.DirFS(e.ProjectDir), glob)
		if err != nil || len(files) == 0 {
			return ManifestProbe{}, false
		}
		pathUnits := unitManifest(frameItems("path", glob, files))
		childState := stampUnits(pathUnits)
		childStates = append(childStates, glob+"\x00"+childState)
		childManifest = append(childManifest, Unit{
			Kind:   "child",
			Key:    glob,
			Digest: digestString(childState),
			Bytes:  len(childState),
		})
	}
	return ManifestProbe{ExternalState: stampStrings(childStates), Manifest: childManifest}, true
}

func (e *Engine) pipelineManifest(pipeline []Step, item string) (ManifestProbe, bool) {
	var frame []frameItem
	for _, st := range pipeline {
		switch {
		case st.Gather != nil:
			vals, err := e.gather(*st.Gather, item)
			if err != nil {
				return ManifestProbe{}, false
			}
			frame = vals
		case st.Filter != nil:
			frame = applyFilter(frame, *st.Filter)
		case st.Transform != nil:
			frame = applyTransform(frame, *st.Transform)
		case st.Eval != nil:
			manifest := unitManifest(frame)
			return ManifestProbe{ExternalState: stampUnits(manifest), Manifest: manifest}, true
		case st.Loop != nil:
			return e.loopManifest(*st.Loop)
		}
	}
	return ManifestProbe{}, false
}

func (e *Engine) loopManifest(l Loop) (ManifestProbe, bool) {
	if !reusableLoop(l) {
		return ManifestProbe{}, false
	}
	items, err := e.gather(l.Over, "")
	if err != nil || len(items) == 0 {
		return ManifestProbe{}, false
	}
	var childStates []string
	var childManifest []Unit
	for _, item := range items {
		loopItem := item.value
		probe, ok := e.pipelineManifestForReuse(l.Do, loopItem)
		if !ok {
			return ManifestProbe{}, false
		}
		childStates = append(childStates, loopItem+"\x00"+probe.ExternalState)
		childManifest = append(childManifest, Unit{
			Kind:   "child",
			Key:    loopItem,
			Digest: digestString(probe.ExternalState),
			Bytes:  len(probe.ExternalState),
		})
	}
	return ManifestProbe{ExternalState: stampStrings(childStates), Manifest: childManifest}, true
}

func (e *Engine) pipelineManifestForReuse(pipeline []Step, item string) (ManifestProbe, bool) {
	if pipelineHasCommand(pipeline) {
		return e.dependencyManifest(pipeline, item)
	}
	return e.pipelineManifest(pipeline, item)
}

func (e *Engine) dependencyManifest(pipeline []Step, item string) (ManifestProbe, bool) {
	var units []Unit
	for _, st := range pipeline {
		switch {
		case st.Gather != nil:
			gatherUnits, ok := e.gatherDependencyUnits(*st.Gather, item)
			if !ok {
				return ManifestProbe{}, false
			}
			units = append(units, gatherUnits...)
		case st.Eval != nil:
			return ManifestProbe{ExternalState: stampUnits(units), Manifest: units}, true
		case st.Loop != nil:
			probe, ok := e.loopManifest(*st.Loop)
			if !ok {
				return ManifestProbe{}, false
			}
			return probe, true
		}
	}
	return ManifestProbe{}, false
}

func (e *Engine) gatherDependencyUnits(g Gather, item string) ([]Unit, bool) {
	if len(g.Each) > 0 {
		var out []Unit
		for _, sub := range g.Each {
			units, ok := e.gatherDependencyUnits(sub, item)
			if !ok {
				return nil, false
			}
			out = append(out, units...)
		}
		return out, true
	}
	if g.Command != "" {
		return e.commandInputUnits(g, item)
	}
	frame, err := e.gather(g, item)
	if err != nil {
		return nil, false
	}
	return unitManifest(frame), true
}

func (e *Engine) commandInputUnits(g Gather, item string) ([]Unit, bool) {
	command := subst(g.Command, item)
	if len(g.Inputs) == 0 {
		return nil, false
	}
	units := []Unit{{
		Kind:   "command",
		Key:    command,
		Digest: digestString(command),
		Bytes:  len(command),
	}}
	covered := map[string]bool{}
	for _, input := range g.Inputs {
		glob := subst(input, item)
		files, err := codemap.GlobFiles(os.DirFS(e.ProjectDir), glob)
		if err != nil || len(files) == 0 {
			return nil, false
		}
		sort.Strings(files)
		for _, file := range files {
			covered[normalizeProjectPath(file)] = true
			b, err := os.ReadFile(filepath.Join(e.ProjectDir, file))
			if err != nil {
				return nil, false
			}
			value := strings.TrimRight(string(b), "\n")
			units = append(units, Unit{
				Kind:   "command_input",
				Key:    command + "#input#" + glob + "#" + file,
				Digest: digestString(value),
				Bytes:  len([]byte(value)),
			})
		}
	}
	if !e.commandPathTokensCovered(command, covered) {
		return nil, false
	}
	return units, true
}

func (e *Engine) commandPathTokensCovered(command string, covered map[string]bool) bool {
	for _, token := range commandPathTokens(command) {
		info, err := os.Stat(filepath.Join(e.ProjectDir, token))
		if err != nil || info.IsDir() {
			continue
		}
		if !covered[normalizeProjectPath(token)] {
			return false
		}
	}
	return true
}

func commandPathTokens(command string) []string {
	var out []string
	seen := map[string]bool{}
	fields := strings.Fields(command)
	for i, field := range fields {
		if i > 0 && (fields[i-1] == "-v" || fields[i-1] == "--exclude") {
			continue
		}
		if strings.ContainsAny(field, "()") {
			continue
		}
		if strings.HasPrefix(field, "'") || strings.HasPrefix(field, `"`) {
			if !strings.Contains(field, "/") {
				continue
			}
		}
		token := strings.Trim(field, `"'(){}[];,|&<>`)
		token = strings.TrimPrefix(token, "--include=")
		token = strings.Trim(token, `"'`)
		if token == "" || strings.HasPrefix(token, "-") || strings.Contains(token, "*") {
			continue
		}
		if !strings.ContainsAny(token, "/.") {
			continue
		}
		token = normalizeProjectPath(token)
		if token == "." || strings.HasPrefix(token, "..") || seen[token] {
			continue
		}
		seen[token] = true
		out = append(out, token)
	}
	sort.Strings(out)
	return out
}

func normalizeProjectPath(path string) string {
	path = filepath.ToSlash(filepath.Clean(path))
	return strings.TrimPrefix(path, "./")
}

func reusablePipeline(pipeline []Step) bool {
	return reusePolicyForPipeline(pipeline).Class == ReuseDeterministic
}

func reusePolicyForPipeline(pipeline []Step) ReusePolicy {
	if len(pipeline) == 0 {
		return ReusePolicy{Class: ReuseUnsupported, Reason: "empty pipeline"}
	}
	for _, st := range pipeline {
		switch {
		case st.Gather != nil:
			if policy := reusePolicyForGather(*st.Gather); policy.Class != ReuseDeterministic {
				return policy
			}
		case st.Eval != nil:
			if st.Eval.Judgement != "" {
				return ReusePolicy{Class: ReuseJudgement, Reason: "judgement terminal requires non-mechanical review"}
			}
			return ReusePolicy{Class: ReuseDeterministic, Reason: "all gather and terminal steps are mechanically bounded"}
		case st.Loop != nil:
			if policy := reusePolicyForLoop(*st.Loop); policy.Class != ReuseDeterministic {
				return policy
			}
			return ReusePolicy{Class: ReuseDeterministic, Reason: "all gather and terminal steps are mechanically bounded"}
		}
	}
	return ReusePolicy{Class: ReuseUnsupported, Reason: "pipeline has no terminal eval or loop"}
}

func reusableLoop(l Loop) bool {
	return reusePolicyForLoop(l).Class == ReuseDeterministic
}

func reusePolicyForLoop(l Loop) ReusePolicy {
	if policy := reusePolicyForGather(l.Over); policy.Class != ReuseDeterministic {
		return policy
	}
	return reusePolicyForPipeline(l.Do)
}

func reusableGather(g Gather) bool {
	return reusePolicyForGather(g).Class == ReuseDeterministic
}

func reusePolicyForGather(g Gather) ReusePolicy {
	if g.Command != "" {
		if len(g.Inputs) == 0 {
			return ReusePolicy{Class: ReuseLivePolicy, Reason: "command gather output is not mechanically bounded by C3"}
		}
		return ReusePolicy{Class: ReuseDeterministic, Reason: "command gather declares deterministic input surfaces"}
	}
	for _, sub := range g.Each {
		if policy := reusePolicyForGather(sub); policy.Class != ReuseDeterministic {
			return policy
		}
	}
	return ReusePolicy{Class: ReuseDeterministic, Reason: "gather source is mechanically bounded"}
}

func pipelineHasCommand(pipeline []Step) bool {
	for _, st := range pipeline {
		switch {
		case st.Gather != nil:
			if gatherHasCommand(*st.Gather) {
				return true
			}
		case st.Loop != nil:
			if gatherHasCommand(st.Loop.Over) || pipelineHasCommand(st.Loop.Do) {
				return true
			}
		}
	}
	return false
}

func gatherHasCommand(g Gather) bool {
	if g.Command != "" {
		return true
	}
	for _, sub := range g.Each {
		if gatherHasCommand(sub) {
			return true
		}
	}
	return false
}

func (e *Engine) run(spec Spec, pipeline []Step, item string) Verdict {
	var frame []frameItem
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
	var childStates []string
	var childManifest []Unit
	for _, item := range items {
		loopItem := item.value
		v := e.run(spec, l.Do, loopItem)
		evidence = append(evidence, fmt.Sprintf("%s → %s", loopItem, v.Verdict))
		childState := childMatchedState(v)
		childStates = append(childStates, loopItem+"\x00"+childState)
		childManifest = append(childManifest, Unit{
			Kind:   "child",
			Key:    loopItem,
			Digest: digestString(childState),
			Bytes:  len(childState),
		})
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
	return Verdict{Fact: spec.Fact, Claim: spec.Claim, Verdict: verdict, ExternalState: stampStrings(childStates), Evidence: evidence, Manifest: childManifest}
}

func childMatchedState(v Verdict) string {
	if v.ExternalState != "" {
		return v.ExternalState
	}
	return "error:" + digestString(strings.Join(v.Evidence, "\x00"))
}

func (e *Engine) gather(g Gather, item string) ([]frameItem, error) {
	switch {
	case len(g.Each) > 0:
		var out []frameItem
		for _, sub := range g.Each {
			vals, err := e.gather(sub, item)
			if err != nil {
				return nil, err
			}
			out = append(out, vals...)
		}
		return out, nil
	case g.File != "":
		path := subst(g.File, item)
		b, err := os.ReadFile(filepath.Join(e.ProjectDir, path))
		if err != nil {
			return nil, err
		}
		return []frameItem{{kind: "file", key: path, value: strings.TrimRight(string(b), "\n")}}, nil
	case g.Command != "":
		return e.exec(subst(g.Command, item))
	case g.Files != "":
		glob := subst(g.Files, item)
		files, err := codemap.GlobFiles(os.DirFS(e.ProjectDir), glob)
		if err != nil {
			return nil, err
		}
		return frameItems("path", glob, files), nil
	case g.Facts != "":
		glob := subst(g.Facts, item)
		return frameItems("fact", glob, matchIDs(e.FactIDs, glob)), nil
	case len(g.Literal) > 0:
		out := make([]frameItem, len(g.Literal))
		for i, v := range g.Literal {
			out[i] = frameItem{kind: "literal", key: fmt.Sprintf("literal[%d]", i), value: subst(v, item)}
		}
		return out, nil
	case g.Outline != nil:
		return e.outline(*g.Outline, item)
	case g.Code != "":
		id := subst(g.Code, item)
		globs, ok := e.CodeBindings[id]
		if !ok || len(globs) == 0 {
			return nil, fmt.Errorf("code %s has no binding", id)
		}
		var files []string
		var missing []string
		for _, glob := range globs {
			fs, err := codemap.GlobFiles(os.DirFS(e.ProjectDir), glob)
			if err != nil {
				return nil, err
			}
			if len(fs) == 0 {
				missing = append(missing, glob)
				continue
			}
			for _, file := range fs {
				files = append(files, glob+"\x00"+file)
			}
		}
		if len(missing) > 0 {
			return nil, fmt.Errorf("code %s unmatched glob(s): %s", id, strings.Join(missing, ", "))
		}
		sort.Strings(files)
		out := make([]frameItem, 0, len(files))
		for _, packed := range files {
			parts := strings.SplitN(packed, "\x00", 2)
			out = append(out, frameItem{kind: "code_path", key: id + "#" + parts[0] + "#" + parts[1], value: parts[1]})
		}
		return out, nil
	}
	return nil, fmt.Errorf("empty gather")
}

func (e *Engine) exec(command string) ([]frameItem, error) {
	c := exec.Command("sh", "-c", command)
	c.Dir = e.ProjectDir
	out, err := c.Output()
	if err != nil {
		// Exit 1 is a common "no match" signal from grep-like tools and is a
		// legitimately empty gather. Exit 2+ is a command failure; treating it
		// as empty lets broken paths and bad regexes pass count == 0 checks.
		if exitErr, isExit := err.(*exec.ExitError); isExit {
			if exitErr.ExitCode() == 1 {
				return frameItems("command_line", command, splitLines(string(out))), nil
			}
			stderr := strings.TrimSpace(string(exitErr.Stderr))
			if stderr != "" {
				return nil, fmt.Errorf("command %q exited %d: %s", command, exitErr.ExitCode(), stderr)
			}
			return nil, fmt.Errorf("command %q exited %d", command, exitErr.ExitCode())
		} else {
			return nil, fmt.Errorf("command %q: %w", command, err)
		}
	}
	return frameItems("command_line", command, splitLines(string(out))), nil
}

func (e *Engine) outline(g OutlineGather, item string) ([]frameItem, error) {
	exe, err := astGrepExecutable()
	if err != nil {
		return nil, err
	}
	c := exec.Command(exe, outlineArgs(g, item)...)
	c.Dir = e.ProjectDir
	var stderr bytes.Buffer
	c.Stderr = &stderr
	out, err := c.Output()
	if err != nil {
		if exitErr, isExit := err.(*exec.ExitError); isExit {
			msg := strings.TrimSpace(stderr.String())
			if msg != "" {
				return nil, fmt.Errorf("ast-grep outline exited %d: %s", exitErr.ExitCode(), msg)
			}
			return nil, fmt.Errorf("ast-grep outline exited %d", exitErr.ExitCode())
		}
		return nil, fmt.Errorf("ast-grep outline: %w", err)
	}
	return outlineFrameItems(out)
}

func astGrepExecutable() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("C3_AST_GREP")); configured != "" {
		return configured, nil
	}
	if bundled := bundledAstGrepExecutable(); bundled != "" {
		return bundled, nil
	}
	path, err := exec.LookPath("ast-grep")
	if err != nil {
		return "", fmt.Errorf("ast-grep executable not found; install ast-grep %s+ for gather.outline", minimumAstGrepVersion)
	}
	return path, nil
}

func bundledAstGrepExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	pattern := filepath.Join(filepath.Dir(exe), fmt.Sprintf("ast-grep-*-%s-%s", runtime.GOOS, runtime.GOARCH))
	matches, err := filepath.Glob(pattern)
	if err != nil || len(matches) == 0 {
		return ""
	}
	sort.Strings(matches)
	for i := len(matches) - 1; i >= 0; i-- {
		if info, err := os.Stat(matches[i]); err == nil && !info.IsDir() {
			return matches[i]
		}
	}
	return ""
}

func outlineArgs(g OutlineGather, item string) []string {
	view := strings.TrimSpace(subst(g.View, item))
	if view == "" {
		view = "digest"
	}
	args := []string{"outline", "--json=stream", "--view", view, "--color", "never"}
	if g.Lang != "" {
		args = append(args, "--lang", subst(g.Lang, item))
	}
	if g.Items != "" {
		args = append(args, "--items", subst(g.Items, item))
	}
	if g.SymbolTypes != "" {
		args = append(args, "--type", subst(g.SymbolTypes, item))
	}
	if g.Match != "" {
		args = append(args, "--match", subst(g.Match, item))
	}
	if g.PubMembers {
		args = append(args, "--pub-members")
	}
	for _, glob := range g.Globs {
		args = append(args, "--globs", subst(glob, item))
	}
	for _, rule := range g.OutlineRules {
		args = append(args, "--outline-rules", subst(rule, item))
	}
	if g.NoDefaultOutlineRules {
		args = append(args, "--no-default-outline-rules")
	}
	paths := g.Paths
	if len(paths) == 0 {
		paths = []string{"."}
	}
	for _, path := range paths {
		args = append(args, subst(path, item))
	}
	return args
}

type outlineFile struct {
	Path     string        `json:"path"`
	Language string        `json:"language"`
	Items    []outlineItem `json:"items"`
}

type outlineItem struct {
	Role       string          `json:"role"`
	SymbolType string          `json:"symbolType"`
	Name       string          `json:"name"`
	Signature  string          `json:"signature"`
	ASTKind    string          `json:"astKind"`
	IsImport   bool            `json:"isImport"`
	IsExported bool            `json:"isExported"`
	Members    []outlineMember `json:"members"`
}

type outlineMember struct {
	Role       string `json:"role"`
	SymbolType string `json:"symbolType"`
	Name       string `json:"name"`
	Signature  string `json:"signature"`
	ASTKind    string `json:"astKind"`
	IsPublic   bool   `json:"isPublic"`
}

func outlineFrameItems(data []byte) ([]frameItem, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var files []outlineFile
	for {
		var file outlineFile
		if err := dec.Decode(&file); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, fmt.Errorf("parse ast-grep outline JSON: %w", err)
		}
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	var out []frameItem
	for _, file := range files {
		path := normalizeProjectPath(file.Path)
		for i, item := range file.Items {
			itemKey := fmt.Sprintf("%s#item[%d]#%s#%s", path, i, item.SymbolType, item.Name)
			out = append(out, frameItem{kind: "outline", key: itemKey, value: outlineItemValue(file, item)})
			for j, member := range item.Members {
				memberKey := fmt.Sprintf("%s#item[%d]#member[%d]#%s#%s", path, i, j, member.SymbolType, member.Name)
				out = append(out, frameItem{kind: "outline", key: memberKey, value: outlineMemberValue(file, item, member)})
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key < out[j].key })
	return out, nil
}

func outlineItemValue(file outlineFile, item outlineItem) string {
	return strings.Join([]string{
		"language=" + file.Language,
		"role=" + item.Role,
		"symbolType=" + item.SymbolType,
		"name=" + item.Name,
		"signature=" + item.Signature,
		"astKind=" + item.ASTKind,
		"isImport=" + strconv.FormatBool(item.IsImport),
		"isExported=" + strconv.FormatBool(item.IsExported),
	}, "\x00")
}

func outlineMemberValue(file outlineFile, parent outlineItem, member outlineMember) string {
	return strings.Join([]string{
		"language=" + file.Language,
		"parentRole=" + parent.Role,
		"parentSymbolType=" + parent.SymbolType,
		"parentName=" + parent.Name,
		"role=" + member.Role,
		"symbolType=" + member.SymbolType,
		"name=" + member.Name,
		"signature=" + member.Signature,
		"astKind=" + member.ASTKind,
		"isPublic=" + strconv.FormatBool(member.IsPublic),
	}, "\x00")
}

func (e *Engine) assert(spec Spec, frame []frameItem, ev Eval, item string) Verdict {
	v := func(verdict string, evidence ...string) Verdict {
		manifest := unitManifest(frame)
		return Verdict{Fact: spec.Fact, Claim: spec.Claim, Verdict: verdict, ExternalState: stampUnits(manifest), Evidence: evidence, Manifest: manifest}
	}
	values := frameValues(frame)
	switch {
	case ev.Judgement != "":
		return v("needs-judgement", "judge: "+subst(ev.Judgement, item), fmt.Sprintf("%d value(s) gathered", len(frame)))
	case ev.Exists:
		var missing []string
		for _, p := range values {
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
		if len(uniq(values)) <= 1 {
			return v("holds", "all equal: "+joinShort(values))
		}
		return v("drift", "distinct: "+strings.Join(uniq(values), " | "))
	case ev.Equals != "":
		for _, x := range values {
			if strings.TrimSpace(x) != ev.Equals {
				return v("drift", fmt.Sprintf("%q ≠ %q", x, ev.Equals))
			}
		}
		return v("holds", "all == "+ev.Equals)
	case len(ev.ContainsAll) > 0:
		joined := strings.Join(values, "\n")
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
		if strings.Contains(strings.Join(values, "\n"), ev.Contains) {
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

func applyFilter(frame []frameItem, f Filter) []frameItem {
	var re *regexp.Regexp
	if f.Matches != "" {
		re = regexp.MustCompile(f.Matches)
	}
	out := frame[:0:0]
	for _, x := range frame {
		if f.Contains != "" && !strings.Contains(x.value, f.Contains) {
			continue
		}
		if re != nil && !re.MatchString(x.value) {
			continue
		}
		out = append(out, x)
	}
	return out
}

func applyTransform(frame []frameItem, t Transform) []frameItem {
	if t.Lines {
		var out []frameItem
		for _, x := range frame {
			lines := splitLines(x.value)
			for i, line := range lines {
				out = append(out, frameItem{kind: "line", key: fmt.Sprintf("%s#line[%d]", x.key, i), value: line})
			}
		}
		frame = out
	}
	if t.CodeBlocks {
		var out []frameItem
		for _, x := range frame {
			for i, block := range codeBlockValues(x.value) {
				out = append(out, frameItem{kind: "code_block", key: fmt.Sprintf("%s#code_block[%d]", x.key, i), value: block})
			}
		}
		frame = out
	}
	if t.Trim {
		for i := range frame {
			frame[i].value = strings.TrimSpace(frame[i].value)
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

func frameItems(kind, source string, values []string) []frameItem {
	out := make([]frameItem, len(values))
	for i, value := range values {
		out[i] = frameItem{kind: kind, key: fmt.Sprintf("%s#%s[%d]", source, kind, i), value: value}
	}
	return out
}

func frameValues(frame []frameItem) []string {
	out := make([]string, len(frame))
	for i, item := range frame {
		out[i] = item.value
	}
	return out
}

func unitManifest(frame []frameItem) []Unit {
	out := make([]Unit, len(frame))
	for i, item := range frame {
		out[i] = Unit{
			Kind:   item.kind,
			Key:    item.key,
			Digest: digestString(item.value),
			Bytes:  len([]byte(item.value)),
		}
	}
	return out
}

func codeBlockValues(markdown string) []string {
	tree := content.ParseMarkdown("eval", markdown)
	var out []string
	for _, node := range tree.Nodes {
		if node.Type == "code_block" {
			out = append(out, node.Content)
		}
	}
	return out
}

func stampUnits(units []Unit) string {
	parts := make([]string, len(units))
	for i, unit := range units {
		parts[i] = unit.Kind + "\x00" + unit.Key + "\x00" + unit.Digest + "\x00" + strconv.Itoa(unit.Bytes)
	}
	return stampStrings(parts)
}

func stampStrings(parts []string) string {
	h := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(h[:])[:12]
}

func digestString(value string) string {
	h := sha256.Sum256([]byte(value))
	return hex.EncodeToString(h[:])[:12]
}
