package thvclient

import (
	"slices"
	"testing"
)

const (
	toolA  = "tool_a"
	toolB  = "tool_b"
	myTool = "my_tool"
)

func TestParseToolsJSON_Valid(t *testing.T) {
	t.Parallel()
	output := `{"tools":[{"name":"tool_b","description":"B"},{"name":"tool_a","description":"A"}]}`
	tools, err := ParseToolsJSON(output)
	if err != nil {
		t.Fatalf("ParseToolsJSON failed: %v", err)
	}
	expected := []string{toolA, toolB}
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
	if len(tools) != 1 || tools[0] != myTool {
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
	expected := []string{toolA, toolB}
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

func TestParseToolDefinitions_Valid(t *testing.T) {
	t.Parallel()
	output := `{"tools":[
		{"name":"tool_b","description":"Does B","inputSchema":{"type":"object","properties":{"x":{"type":"string"}},"required":["x"]}},
		{"name":"tool_a","description":"Does A","inputSchema":{"type":"object","properties":{}},"annotations":{"readOnlyHint":true}}
	]}`
	defs, err := ParseToolDefinitions(output)
	if err != nil {
		t.Fatalf("ParseToolDefinitions failed: %v", err)
	}
	if len(defs) != 2 {
		t.Fatalf("expected 2 tool definitions, got %d", len(defs))
	}
	// Verify sorted by name
	if defs[0].Name != toolA || defs[1].Name != toolB {
		t.Errorf("expected sorted [tool_a, tool_b], got [%s, %s]", defs[0].Name, defs[1].Name)
	}
	// Verify descriptions
	if defs[0].Description != "Does A" {
		t.Errorf("expected description 'Does A', got %q", defs[0].Description)
	}
	if defs[1].Description != "Does B" {
		t.Errorf("expected description 'Does B', got %q", defs[1].Description)
	}
	// Verify inputSchema
	if defs[1].InputSchema.Type != "object" {
		t.Errorf("expected inputSchema type 'object', got %q", defs[1].InputSchema.Type)
	}
	if len(defs[1].InputSchema.Required) != 1 || defs[1].InputSchema.Required[0] != "x" {
		t.Errorf("expected required [x], got %v", defs[1].InputSchema.Required)
	}
	// Verify annotations
	if defs[0].Annotations.ReadOnlyHint == nil || !*defs[0].Annotations.ReadOnlyHint {
		t.Error("expected tool_a annotations.readOnlyHint to be true")
	}
}

func TestParseToolDefinitions_Sorted(t *testing.T) {
	t.Parallel()
	output := `{"tools":[
		{"name":"zeta","description":"Z","inputSchema":{"type":"object","properties":{}}},
		{"name":"alpha","description":"A","inputSchema":{"type":"object","properties":{}}},
		{"name":"mid","description":"M","inputSchema":{"type":"object","properties":{}}}
	]}`
	defs, err := ParseToolDefinitions(output)
	if err != nil {
		t.Fatalf("ParseToolDefinitions failed: %v", err)
	}
	names := make([]string, len(defs))
	for i, d := range defs {
		names[i] = d.Name
	}
	expected := []string{"alpha", "mid", "zeta"}
	if !slices.Equal(names, expected) {
		t.Errorf("expected %v, got %v", expected, names)
	}
}

func TestParseToolDefinitions_FallsBackToNil(t *testing.T) {
	t.Parallel()
	output := `TOOLS:
NAME         DESCRIPTION
my_tool      Does things`
	defs, err := ParseToolDefinitions(output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if defs != nil {
		t.Errorf("expected nil for text-only output, got %v", defs)
	}
}

func TestParseToolDefinitions_EmptyTools(t *testing.T) {
	t.Parallel()
	output := `{"tools":[]}`
	defs, err := ParseToolDefinitions(output)
	if err != nil {
		t.Fatalf("ParseToolDefinitions failed: %v", err)
	}
	if len(defs) != 0 {
		t.Errorf("expected empty slice, got %v", defs)
	}
}

func TestParseToolDefinitions_WithWarnings(t *testing.T) {
	t.Parallel()
	output := `WARNING: something happened
{"tools":[{"name":"my_tool","description":"test tool","inputSchema":{"type":"object","properties":{"a":{"type":"number"}}}}]}`
	defs, err := ParseToolDefinitions(output)
	if err != nil {
		t.Fatalf("ParseToolDefinitions failed: %v", err)
	}
	if len(defs) != 1 {
		t.Fatalf("expected 1 tool definition, got %d", len(defs))
	}
	if defs[0].Name != myTool {
		t.Errorf("expected 'my_tool', got %q", defs[0].Name)
	}
	if defs[0].Description != "test tool" {
		t.Errorf("expected 'test tool', got %q", defs[0].Description)
	}
}
