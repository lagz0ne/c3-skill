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

func TestComponentSchemaHasGovernanceAndNoUpCap(t *testing.T) {
	sections := ForType("component")
	foundGovernance := false
	for _, s := range sections {
		if s.Name == "Governance" {
			foundGovernance = true
		}
		if s.Name == "Up Cap" {
			t.Error("component schema should not include 'Up Cap'")
		}
	}
	if !foundGovernance {
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

func TestADRDefinitionOwnsCurrentSectionsAndRejectRules(t *testing.T) {
	def, ok := DefinitionFor("adr")
	if !ok {
		t.Fatal("adr definition should exist")
	}
	if def.ID != "adr" {
		t.Fatalf("definition id = %q, want adr", def.ID)
	}
	if len(def.Sections) == 0 {
		t.Fatal("adr definition should expose sections")
	}
	if len(def.Reject.Bullets) == 0 {
		t.Fatal("adr definition should expose reject rules")
	}
	if got, want := def.Reject, RejectFor("adr"); len(got.Bullets) != len(want.Bullets) || got.Workorder != want.Workorder {
		t.Fatalf("adr definition reject rules should match reject registry")
	}

	registrySections := ForType("adr")
	if len(def.Sections) != len(registrySections) {
		t.Fatalf("definition sections = %d, registry sections = %d", len(def.Sections), len(registrySections))
	}
	for i := range registrySections {
		if def.Sections[i].Name != registrySections[i].Name {
			t.Fatalf("section %d = %q, want %q", i, def.Sections[i].Name, registrySections[i].Name)
		}
	}
}
