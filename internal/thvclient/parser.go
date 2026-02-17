package thvclient

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
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
