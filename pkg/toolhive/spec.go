package toolhive

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// UpdateSpecTools updates the tools field in a spec file
func UpdateSpecTools(path string, tools []string) error {
	// Read the original file
	data, err := os.ReadFile(path) // #nosec G304 - path is controlled by application
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Parse with yaml.v3 to preserve structure
	var doc yaml.Node
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Update the tools field
	if err := updateToolsInNode(&doc, tools); err != nil {
		return fmt.Errorf("failed to update tools: %w", err)
	}

	// Marshal back preserving structure
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&doc); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	// Write back to file
	return os.WriteFile(path, buf.Bytes(), 0600)
}

// updateToolsInNode updates the tools field in the YAML node tree
func updateToolsInNode(node *yaml.Node, tools []string) error {
	// Navigate to the document content
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return updateToolsInNode(node.Content[0], tools)
	}

	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected mapping node, got %v", node.Kind)
	}

	// Find or create tools section
	toolsIndex := -1
	for i := 0; i < len(node.Content); i += 2 {
		if node.Content[i].Value == "tools" {
			toolsIndex = i
			break
		}
	}

	// Create new tools array node
	toolsNode := &yaml.Node{
		Kind:    yaml.SequenceNode,
		Content: make([]*yaml.Node, 0, len(tools)),
	}

	for _, tool := range tools {
		toolsNode.Content = append(toolsNode.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: tool,
		})
	}

	if toolsIndex >= 0 {
		// Replace existing tools
		node.Content[toolsIndex+1] = toolsNode
	} else {
		// Add new tools section
		node.Content = append(node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "tools"},
			toolsNode,
		)
	}

	return nil
}

// AddWarningComment adds a warning comment to a spec file
func AddWarningComment(path, warning, detail string) error {
	// Read the original file
	data, err := os.ReadFile(path) // #nosec G304 - path is controlled by application
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Check if warning already exists
	if bytes.Contains(data, []byte(warning)) {
		// Warning already exists
		return nil
	}

	// Add warning comment at the beginning (after any existing header comments)
	lines := bytes.Split(data, []byte("\n"))
	var output bytes.Buffer
	warningAdded := false

	for i, line := range lines {
		// Write existing line
		if i > 0 {
			output.WriteByte('\n')
		}
		output.Write(line)

		// Add warning after initial comments but before content
		if !warningAdded && !bytes.HasPrefix(bytes.TrimSpace(line), []byte("#")) && len(bytes.TrimSpace(line)) > 0 {
			// Insert warning before this line
			output.Reset()
			if i > 0 {
				// Write previous lines
				for j := 0; j < i; j++ {
					if j > 0 {
						output.WriteByte('\n')
					}
					output.Write(lines[j])
				}
				output.WriteByte('\n')
			}

			// Add warning
			output.WriteString(fmt.Sprintf("# WARNING: %s on %s\n", warning, time.Now().Format("2006-01-02")))
			output.WriteString(fmt.Sprintf("# %s\n", detail))

			// Write current line
			output.Write(line)
			warningAdded = true
		}
	}

	// Write back to file
	return os.WriteFile(path, output.Bytes(), 0600)
}
