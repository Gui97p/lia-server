package llm

import "sort"

func BuildPlanSchema(tools []ToolDefinition) map[string]any {
	branches := make([]any, 0, len(tools))
	for _, t := range tools {
		branches = append(branches, map[string]any{
			"type": "object",
			"properties": map[string]any{
				"id": map[string]any{"type": "string"},
				"dependsOn": map[string]any{
					"type":  "array",
					"items": map[string]any{"type": "string"},
				},
				"capability": map[string]any{"const": t.Name},
				"params":     normalizeStrict(t.Parameters),
			},
			"required":             []string{"id", "dependsOn", "capability", "params"},
			"additionalProperties": false,
		})
	}

	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"steps": map[string]any{
				"type":  "array",
				"items": map[string]any{"oneOf": branches},
			},
		},
		"required":             []string{"steps"},
		"additionalProperties": false,
	}
}

func fromStepRefSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"$fromStep": map[string]any{
				"type":        "string",
				"description": "ID (Step.ID) do step anterior cujo resultado deve ser usado aqui.",
			},
			"match": map[string]any{
				"type":        "array",
				"description": "Filtros de igualdade exata a aplicar sobre o resultado do step referenciado, se ele for uma lista. Deve resultar em exatamente 1 item; deixe vazio se o step referenciado já retorna um único resultado.",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"field": map[string]any{"type": "string"},
						"value": map[string]any{"type": "string"},
					},
					"required":             []string{"field", "value"},
					"additionalProperties": false,
				},
			},
		},
		"required":             []string{"$fromStep", "match"},
		"additionalProperties": false,
	}
}

func isFromStepRefMarker(v any) bool {
	b, ok := v.(bool)
	return ok && b
}

func normalizeStrict(node any) any {
	switch v := node.(type) {
	case map[string]any:
		if isFromStepRefMarker(v["fromStepRef"]) {
			withoutMarker := make(map[string]any, len(v))
			for key, value := range v {
				if key == "fromStepRef" {
					continue
				}
				withoutMarker[key] = value
			}
			return map[string]any{
				"anyOf": []any{normalizeStrict(withoutMarker), fromStepRefSchema()},
			}
		}

		out := make(map[string]any, len(v))
		for key, value := range v {
			out[key] = normalizeStrict(value)
		}

		if out["type"] == "object" {
			if props, ok := out["properties"].(map[string]any); ok {
				originalRequired := stringSet(out["required"])
				required := make([]string, 0, len(props))
				for name, propSchema := range props {
					required = append(required, name)
					if !originalRequired[name] {
						props[name] = makeNullable(propSchema)
					}
				}
				sort.Strings(required)
				out["required"] = required
				out["additionalProperties"] = false
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = normalizeStrict(item)
		}
		return out
	default:
		return v
	}
}

func stringSet(value any) map[string]bool {
	set := map[string]bool{}
	switch v := value.(type) {
	case []string:
		for _, s := range v {
			set[s] = true
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				set[s] = true
			}
		}
	}
	return set
}

func makeNullable(schema any) any {
	m, ok := schema.(map[string]any)
	if !ok {
		return schema
	}

	if anyOf, ok := m["anyOf"].([]any); ok {
		return map[string]any{"anyOf": append(anyOf, map[string]any{"type": "null"})}
	}

	var types []string
	switch t := m["type"].(type) {
	case string:
		types = []string{t}
	case []string:
		types = t
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok {
				types = append(types, s)
			}
		}
	}

	for _, t := range types {
		if t == "null" {
			return m
		}
	}
	m["type"] = append(types, "null")
	return m
}

func ToGeminiSchema(node any) any {
	switch v := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for key, value := range v {
			switch key {
			case "additionalProperties":
				continue
			case "const":
				out["enum"] = []any{value}
			case "oneOf":
				out["anyOf"] = ToGeminiSchema(value)
			case "type":
				if nonNull, hasNull, ok := splitNullableType(value); ok {
					out["type"] = nonNull
					if hasNull {
						out["nullable"] = true
					}
					continue
				}
				out["type"] = value
			default:
				out[key] = ToGeminiSchema(value)
			}
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = ToGeminiSchema(item)
		}
		return out
	default:
		return v
	}
}

func splitNullableType(value any) (nonNull string, hasNull bool, ok bool) {
	var types []string
	switch t := value.(type) {
	case []string:
		types = t
	case []any:
		for _, item := range t {
			if s, ok := item.(string); ok {
				types = append(types, s)
			}
		}
	default:
		return "", false, false
	}

	for _, t := range types {
		if t == "null" {
			hasNull = true
			continue
		}
		nonNull = t
	}
	return nonNull, hasNull, true
}
