package schema

import "testing"

func TestRegistry_HasAllTypes(t *testing.T) {
	for _, typ := range []string{"component", "container", "context", "ref", "adr"} {
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

func TestRegistry_ComponentHasPurpose(t *testing.T) {
	sections := ForType("component")
	for _, s := range sections {
		if s.Purpose == "" {
			t.Errorf("component section %q has no purpose", s.Name)
		}
	}
}
