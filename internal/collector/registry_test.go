package collector

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()
	if r == nil {
		t.Error("Expected non-nil registry")
	}

	types := r.List()
	if len(types) != 0 {
		t.Errorf("Expected empty registry, got %d collectors", len(types))
	}
}

func TestRegistry_Has(t *testing.T) {
	r := NewRegistry()
	
	if r.Has("mikrotik") {
		t.Error("Expected registry to not have mikrotik collector")
	}
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	types := r.List()
	
	if types == nil {
		t.Error("Expected non-nil list")
	}
	if len(types) != 0 {
		t.Errorf("Expected 0 types, got %d", len(types))
	}
}
