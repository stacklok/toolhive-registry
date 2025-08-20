package toolhive

import (
	"bufio"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/stacklok/toolhive/pkg/logger"
)

// Tool represents an MCP tool
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema,omitempty"`
	Annotations map[string]interface{} `json:"annotations,omitempty"`
}

// MCPListOutput represents the JSON output from thv mcp list
type MCPListOutput struct {
	Tools []Tool `json:"tools"`
}

// ParseToolsJSON parses JSON output from thv mcp list tools --format json
func ParseToolsJSON(output string) ([]string, error) {
	// Find the JSON part (skip any warning messages before the JSON)
	jsonStart := strings.Index(output, "{")
	if jsonStart == -1 {
		// No JSON found, try text parsing as fallback
		return ParseToolsText(output)
	}
	jsonOutput := output[jsonStart:]
	
	var result MCPListOutput
	if err := json.Unmarshal([]byte(jsonOutput), &result); err != nil {
		logger.Debugf("Failed to parse JSON output: %v", err)
		// Fallback to text parsing
		return ParseToolsText(output)
	}
	
	var tools []string
	for _, tool := range result.Tools {
		tools = append(tools, tool.Name)
	}
	
	// Sort tools alphabetically
	sort.Strings(tools)
	
	return tools, nil
}

// ParseToolsText parses text output from thv mcp list (fallback parser)
func ParseToolsText(output string) ([]string, error) {
	var tools []string
	foundToolsSection := false
	foundHeader := false

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		
		// Look for TOOLS: section
		if strings.HasPrefix(line, "TOOLS:") {
			foundToolsSection = true
			continue
		}
		
		// Skip the NAME/DESCRIPTION header
		if foundToolsSection && strings.HasPrefix(line, "NAME") {
			foundHeader = true
			continue
		}
		
		// Extract tool names (first column)
		if foundToolsSection && foundHeader && len(line) > 0 {
			// Split by whitespace and get the first field
			fields := strings.Fields(line)
			if len(fields) > 0 {
				tools = append(tools, fields[0])
			}
		}
	}

	if !foundToolsSection {
		return nil, fmt.Errorf("no TOOLS section found in output")
	}

	// Sort tools alphabetically
	sort.Strings(tools)
	
	return tools, nil
}