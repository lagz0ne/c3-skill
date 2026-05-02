package schema

import "testing"

func TestRegistry_HasAllTypes(t *testing.T) {
	for _, typ := range []string{"component", "container", "context", "ref", "rule", "adr"} {
		if ForType(typ) == nil {
			t.Errorf("Registry missing type %q", typ)
		}
	}
}

func TestRegistry_UnknownType(t *testing.T) {
	if ForType("bogus") != nil {
		t.Error("expected nil for unknown type")
	}
}

func TestPurposeOf(t *testing.T) {
	p := PurposeOf("component", "Goal")
	if p == "" {
		t.Error("expected non-empty purpose for component/Goal")
	}
}

func TestPurposeOf_Unknown(t *testing.T) {
	if PurposeOf("component", "NonExistent") != "" {
		t.Error("expected empty purpose for unknown section")
	}
}

func TestRuleSchemaExists(t *testing.T) {
	sections := ForType("rule")
	if sections == nil {
		t.Fatal("no schema for 'rule'")
	}
	required := map[string]bool{}
	for _, s := range sections {
		if s.Required {
			required[s.Name] = true
		}
	}
	for _, name := range []string{"Goal", "Rule", "Golden Example"} {
		if !required[name] {
			t.Errorf("expected required section %q", name)
		}
	}
}

func TestComponentSchemaHasGovernance(t *testing.T) {
	sections := ForType("component")
	found := false
	for _, s := range sections {
		if s.Name == "Governance" {
			found = true
		}
	}
	if !found {
		t.Error("component schema should have 'Governance' section")
	}
}

func TestRegistry_ComponentHasPurpose(t *testing.T) {
	sections := ForType("component")
	for _, s := range sections {
		if s.Purpose == "" {
			t.Errorf("component section %q has no purpose", s.Name)
		}
	}
}

func TestRefRequiredSectionsHaveFillAndFailure(t *testing.T) {
	sections := ForType("ref")
	if sections == nil {
		t.Fatal("no schema for 'ref'")
	}
	for _, s := range sections {
		if !s.Required {
			continue
		}
		if s.Fill == "" {
			t.Errorf("ref required section %q has empty Fill (rejection contract is incomplete)", s.Name)
		}
		if s.Failure == "" {
			t.Errorf("ref required section %q has empty Failure (cannot reject thin drafts)", s.Name)
		}
	}
}

func TestRuleRequiredSectionsHaveFillAndFailure(t *testing.T) {
	sections := ForType("rule")
	if sections == nil {
		t.Fatal("no schema for 'rule'")
	}
	for _, s := range sections {
		if !s.Required {
			continue
		}
		if s.Fill == "" {
			t.Errorf("rule required section %q has empty Fill (rejection contract is incomplete)", s.Name)
		}
		if s.Failure == "" {
			t.Errorf("rule required section %q has empty Failure (cannot reject thin drafts)", s.Name)
		}
	}
}
