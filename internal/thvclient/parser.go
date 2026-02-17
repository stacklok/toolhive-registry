package thvclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
)

// tool represents an MCP tool from thv mcp list output.
type tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema,omitempty"`
	Annotations map[string]any `json:"annotations,omitempty"`
}

// mcpListOutput represents the JSON output from thv mcp list.
type mcpListOutput struct {
	Tools []tool `json:"tools"`
}

// ParseToolsJSON parses JSON output from thv mcp list tools --format json.
// Falls back to text parsing if JSON parsing fails.
func ParseToolsJSON(output string) ([]string, error) {
	jsonStart := strings.Index(output, "{")
	if jsonStart == -1 {
		return ParseToolsText(output)
	}
	jsonOutput := output[jsonStart:]

	var result mcpListOutput
	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		return ParseToolsText(output)
	}

	var tools []string
	for _, t := range result.Tools {
		tools = append(tools, t.Name)
	}

	sort.Strings(tools)
	return tools, nil
}

// ParseToolDefinitions parses JSON output from thv mcp list tools --format json
// into full mcp.Tool definitions. Returns nil (not an error) if the output is
// text-only, since text format contains only tool names.
func ParseToolDefinitions(output string) ([]mcp.Tool, error) {
	jsonStart := strings.Index(output, "{")
	if jsonStart == -1 {
		// Text format has no schema information — return nil.
		return nil, nil
	}
	jsonOutput := output[jsonStart:]

	var result mcpListOutput
	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		// Not valid JSON — treat as text-only.
		return nil, nil
	}

	defs := make([]mcp.Tool, 0, len(result.Tools))
	for _, t := range result.Tools {
		// JSON round-trip: internal tool struct → bytes → mcp.Tool.
		// This works because mcp.Tool uses default UnmarshalJSON and its
		// nested types (ToolInputSchema, ToolAnnotation) have custom
		// unmarshalers that handle map[string]any → typed struct conversion.
		b, err := json.Marshal(t)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal tool %s: %w", t.Name, err)
		}
		var mcpTool mcp.Tool
		if err := json.Unmarshal(b, &mcpTool); err != nil {
			return nil, fmt.Errorf("failed to unmarshal tool %s into mcp.Tool: %w", t.Name, err)
		}
		defs = append(defs, mcpTool)
	}

	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Name < defs[j].Name
	})

	return defs, nil
}

// ParseToolsText parses text output from thv mcp list (fallback parser).
func ParseToolsText(output string) ([]string, error) {
	var tools []string
	foundToolsSection := false
	foundHeader := false

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "TOOLS:") {
			foundToolsSection = true
			continue
		}

		if foundToolsSection && strings.HasPrefix(line, "NAME") {
			foundHeader = true
			continue
		}

		if foundToolsSection && foundHeader && len(line) > 0 {
			fields := strings.Fields(line)
			if len(fields) > 0 {
				tools = append(tools, fields[0])
			}
		}
	}

	if !foundToolsSection {
		return nil, fmt.Errorf("no TOOLS section found in output")
	}

	sort.Strings(tools)
	return tools, nil
}
