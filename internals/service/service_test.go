package service

import "testing"

func TestDefinition_HasComponent(t *testing.T) {
	d := Definition{
		Components: []string{"a", "b", "c"},
	}

	if !d.HasComponent("a") {
		t.Error("Component not found")
	}
	if !d.HasComponent("b") {
		t.Error("Component not found")
	}
	if !d.HasComponent("c") {
		t.Error("Component not found")
	}
	if d.HasComponent("d") {
		t.Error("Component found")
	}
}
