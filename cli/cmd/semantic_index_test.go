package cmd

import (
	"bytes"
	"testing"

	"github.com/lagz0ne/c3-design/cli/internal/store"
)

func TestRunSemanticIndex_AgentTOONIncludesHelpHints(t *testing.T) {
	provider := &searchSemanticProvider{}
	restore := store.SetSemanticProviderForTest(provider)
	defer restore()
	s := createDBFixture(t)
	t.Setenv("C3X_MODE", "agent")

	var buf bytes.Buffer
	if err := RunSemanticIndex(SemanticIndexOptions{Store: s, JSON: true}, &buf); err != nil {
		t.Fatal(err)
	}

	requireAll(t, buf.String(),
		"help[",
		"c3x search <query>",
		"c3x index",
	)
}
