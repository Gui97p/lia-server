package agent

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Gui97p/lia-server/internal/session"
)

const fromStepRefKey = "$fromStep"

type matchEntry struct {
	Field string `json:"field"`
	Value string `json:"value"`
}

func resolveParams(params map[string]any, stepResults map[string]session.ToolResult) (map[string]any, error) {
	resolved, err := resolveValue(params, stepResults)
	if err != nil {
		return nil, err
	}
	out, ok := resolved.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("resolved params is not an object")
	}
	return out, nil
}

func resolveValue(node any, stepResults map[string]session.ToolResult) (any, error) {
	switch v := node.(type) {
	case map[string]any:
		if stepID, ok := v[fromStepRefKey].(string); ok {
			return resolveFromStepRef(stepID, v["match"], stepResults)
		}
		out := make(map[string]any, len(v))
		for key, value := range v {
			resolved, err := resolveValue(value, stepResults)
			if err != nil {
				return nil, err
			}
			out[key] = resolved
		}
		return out, nil
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			resolved, err := resolveValue(item, stepResults)
			if err != nil {
				return nil, err
			}
			out[i] = resolved
		}
		return out, nil
	default:
		return v, nil
	}
}

func resolveFromStepRef(stepID string, rawMatch any, stepResults map[string]session.ToolResult) (any, error) {
	result, ok := stepResults[stepID]
	if !ok {
		return nil, fmt.Errorf("$fromStep references step %q, which has no recorded result", stepID)
	}

	var data any
	if err := json.Unmarshal(result.Result, &data); err != nil {
		return nil, fmt.Errorf("$fromStep: result of step %q is not valid JSON: %w", stepID, err)
	}

	matches, err := parseMatchEntries(rawMatch)
	if err != nil {
		return nil, err
	}

	list, isList := data.([]any)
	if !isList {
		return data, nil
	}

	var found []any
	for _, item := range list {
		if matchesAll(item, matches) {
			found = append(found, item)
		}
	}

	if len(found) != 1 {
		return nil, fmt.Errorf("$fromStep: match against step %q result found %d items, expected exactly 1", stepID, len(found))
	}

	return found[0], nil
}

func parseMatchEntries(rawMatch any) ([]matchEntry, error) {
	items, ok := rawMatch.([]any)
	if !ok {
		return nil, nil
	}

	entries := make([]matchEntry, 0, len(items))
	for _, item := range items {
		m, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("$fromStep: invalid match entry %v", item)
		}
		field, _ := m["field"].(string)
		value, _ := m["value"].(string)
		entries = append(entries, matchEntry{Field: field, Value: value})
	}
	return entries, nil
}

func matchesAll(item any, entries []matchEntry) bool {
	obj, ok := item.(map[string]any)
	if !ok {
		return false
	}
	for _, entry := range entries {
		if !strings.EqualFold(fmt.Sprintf("%v", obj[entry.Field]), entry.Value) {
			return false
		}
	}
	return true
}
