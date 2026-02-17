package thvclient

import (
	"slices"
	"testing"
)

func TestParseToolsJSON_Valid(t *testing.T) {
	t.Parallel()
	output := `{"tools":[{"name":"tool_b","description":"B"},{"name":"tool_a","description":"A"}]}`
	tools, err := ParseToolsJSON(output)
	if err != nil {
		t.Fatalf("ParseToolsJSON failed: %v", err)
	}
	expected := []string{"tool_a", "tool_b"}
	if !slices.Equal(tools, expected) {
		t.Errorf("expected %v, got %v", expected, tools)
	}
}

func TestParseToolsJSON_WithWarnings(t *testing.T) {
	t.Parallel()
	output := `WARNING: something happened
{"tools":[{"name":"my_tool","description":"test"}]}`
	tools, err := ParseToolsJSON(output)
	if err != nil {
		t.Fatalf("ParseToolsJSON failed: %v", err)
	}
	if len(tools) != 1 || tools[0] != "my_tool" {
		t.Errorf("expected [my_tool], got %v", tools)
	}
}

func TestParseToolsJSON_Empty(t *testing.T) {
	t.Parallel()
	output := `{"tools":[]}`
	tools, err := ParseToolsJSON(output)
	if err != nil {
		t.Fatalf("ParseToolsJSON failed: %v", err)
	}
	if len(tools) != 0 {
		t.Errorf("expected empty tools, got %v", tools)
	}
}

func TestParseToolsJSON_FallsBackToText(t *testing.T) {
	t.Parallel()
	output := `TOOLS:
NAME         DESCRIPTION
my_tool      Does things
other_tool   Does other things`
	tools, err := ParseToolsJSON(output)
	if err != nil {
		t.Fatalf("ParseToolsJSON fallback failed: %v", err)
	}
	expected := []string{"my_tool", "other_tool"}
	if !slices.Equal(tools, expected) {
		t.Errorf("expected %v, got %v", expected, tools)
	}
}

func TestParseToolsText_Valid(t *testing.T) {
	t.Parallel()
	output := `TOOLS:
NAME         DESCRIPTION
tool_b       Does B
tool_a       Does A`
	tools, err := ParseToolsText(output)
	if err != nil {
		t.Fatalf("ParseToolsText failed: %v", err)
	}
	expected := []string{"tool_a", "tool_b"}
	if !slices.Equal(tools, expected) {
		t.Errorf("expected %v, got %v", expected, tools)
	}
}

func TestParseToolsText_NoToolsSection(t *testing.T) {
	t.Parallel()
	output := `Some random output without tools`
	_, err := ParseToolsText(output)
	if err == nil {
		t.Fatal("expected error for missing TOOLS section")
	}
}
