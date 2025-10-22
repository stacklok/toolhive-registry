package registry

import (
	"encoding/json"
	"os"
	"regexp"
	"testing"

	"github.com/modelcontextprotocol/registry/pkg/model"
)

// TestSchemaVersionSync ensures that the schema reference in registry.schema.json
// matches the schema version from the Go package (model.CurrentSchemaVersion).
// This prevents schema drift when upgrading the registry package.
func TestSchemaVersionSync(t *testing.T) {
	// Read the schema file
	schemaPath := "../../schemas/registry.schema.json"
	schemaData, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	// Parse the schema JSON
	var schema map[string]interface{}
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		t.Fatalf("Failed to parse schema JSON: %v", err)
	}

	// Navigate to the $ref field
	servers, ok := schema["properties"].(map[string]interface{})["data"].(map[string]interface{})["properties"].(map[string]interface{})["servers"].(map[string]interface{})
	if !ok {
		t.Fatal("Failed to navigate to servers field in schema")
	}

	items, ok := servers["items"].(map[string]interface{})
	if !ok {
		t.Fatal("Failed to get items field from servers")
	}

	refURL, ok := items["$ref"].(string)
	if !ok {
		t.Fatal("Failed to get $ref URL from items")
	}

	// Extract the date from the URL
	// Expected format: https://static.modelcontextprotocol.io/schemas/2025-10-17/server.schema.json
	re := regexp.MustCompile(`/schemas/([0-9]{4}-[0-9]{2}-[0-9]{2})/`)
	matches := re.FindStringSubmatch(refURL)
	if len(matches) != 2 {
		t.Fatalf("Failed to extract date from schema URL: %s", refURL)
	}
	schemaDate := matches[1]

	// Compare with the Go package constant
	expectedDate := model.CurrentSchemaVersion
	if schemaDate != expectedDate {
		t.Errorf("Schema version mismatch!\n"+
			"  Schema file (%s): %s\n"+
			"  Go package (model.CurrentSchemaVersion): %s\n\n"+
			"To fix: Update schemas/registry.schema.json line 49 to use date %s:\n"+
			"  \"$ref\": \"https://static.modelcontextprotocol.io/schemas/%s/server.schema.json\"",
			schemaPath, schemaDate, expectedDate, expectedDate, expectedDate)
	}

	// Also check the _schema_version metadata for documentation
	schemaVersionMeta, ok := schema["_schema_version"].(map[string]interface{})
	if !ok {
		t.Log("Warning: _schema_version metadata not found (non-critical)")
		return
	}

	metaDate, ok := schemaVersionMeta["schema_date"].(string)
	if ok && metaDate != expectedDate {
		t.Errorf("Schema version metadata is out of sync!\n"+
			"  _schema_version.schema_date: %s\n"+
			"  Expected: %s\n\n"+
			"Update the _schema_version.schema_date field in schemas/registry.schema.json",
			metaDate, expectedDate)
	}
}
