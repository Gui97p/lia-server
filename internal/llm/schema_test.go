package llm_test

import (
	"testing"

	"github.com/Gui97p/lia-server/internal/llm"
)

func TestBuildPlanSchema_NormalizesSloppyCapability(t *testing.T) {
	sloppy := llm.ToolDefinition{
		Name:        "openApp",
		Description: "Abre um aplicativo",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"app": map[string]any{"type": "string"},
				"os":  map[string]any{"type": "string"},
			},
			"required": []string{"app"},
		},
	}

	schema := llm.BuildPlanSchema([]llm.ToolDefinition{sloppy})

	branches := schema["properties"].(map[string]any)["steps"].(map[string]any)["items"].(map[string]any)["oneOf"].([]any)
	if len(branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(branches))
	}
	branch := branches[0].(map[string]any)
	params := branch["properties"].(map[string]any)["params"].(map[string]any)

	if params["additionalProperties"] != false {
		t.Fatalf("expected additionalProperties: false, got %v", params["additionalProperties"])
	}

	required, ok := params["required"].([]string)
	if !ok || len(required) != 2 {
		t.Fatalf("expected required to list both properties, got %v", params["required"])
	}
	requiredSet := map[string]bool{}
	for _, r := range required {
		requiredSet[r] = true
	}
	if !requiredSet["app"] || !requiredSet["os"] {
		t.Fatalf("expected both app and os in required, got %v", required)
	}

	props := params["properties"].(map[string]any)

	appType := props["app"].(map[string]any)["type"]
	if appType != "string" {
		t.Fatalf("expected app's type to stay a plain string, got %v", appType)
	}

	osType, ok := props["os"].(map[string]any)["type"].([]string)
	if !ok || len(osType) != 2 || osType[0] != "string" || osType[1] != "null" {
		t.Fatalf("expected os's type to become [\"string\",\"null\"], got %v", props["os"].(map[string]any)["type"])
	}
}
