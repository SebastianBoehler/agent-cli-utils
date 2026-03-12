package datax

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type pathToken struct {
	key     string
	index   int
	isIndex bool
}

func ParseStructured(input []byte) (any, string, error) {
	trimmed := strings.TrimSpace(string(input))
	if trimmed == "" {
		return nil, "", fmt.Errorf("input is empty")
	}

	var value any
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		if err := json.Unmarshal(input, &value); err != nil {
			return nil, "", fmt.Errorf("invalid JSON: %w", err)
		}

		return value, "json", nil
	}

	if err := yaml.Unmarshal(input, &value); err != nil {
		return nil, "", fmt.Errorf("invalid YAML: %w", err)
	}

	return normalize(value), "yaml", nil
}

func Query(value any, path string) (any, error) {
	tokens, err := tokenize(path)
	if err != nil {
		return nil, err
	}

	current := value
	for _, token := range tokens {
		if token.isIndex {
			items, ok := current.([]any)
			if !ok {
				return nil, fmt.Errorf("cannot index non-array value at [%d]", token.index)
			}

			if token.index < 0 || token.index >= len(items) {
				return nil, fmt.Errorf("array index out of range: %d", token.index)
			}

			current = items[token.index]
			continue
		}

		object, ok := current.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("cannot access key %q on non-object value", token.key)
		}

		next, exists := object[token.key]
		if !exists {
			return nil, fmt.Errorf("key not found: %s", token.key)
		}

		current = next
	}

	return current, nil
}

func Render(value any, format string) ([]byte, error) {
	switch format {
	case "json":
		payload, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return nil, err
		}

		return append(payload, '\n'), nil
	case "yaml":
		return yaml.Marshal(value)
	case "raw":
		switch typed := value.(type) {
		case nil:
			return []byte("null\n"), nil
		case string:
			return []byte(typed + "\n"), nil
		default:
			return []byte(fmt.Sprintf("%v\n", typed)), nil
		}
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}

func tokenize(path string) ([]pathToken, error) {
	if path == "" {
		return nil, nil
	}

	var tokens []pathToken
	i := 0
	if path[0] == '.' {
		i++
	}

	for i < len(path) {
		switch path[i] {
		case '.':
			i++
		case '[':
			end := strings.IndexByte(path[i:], ']')
			if end == -1 {
				return nil, fmt.Errorf("unterminated index in path %q", path)
			}

			rawIndex := path[i+1 : i+end]
			index, err := strconv.Atoi(rawIndex)
			if err != nil {
				return nil, fmt.Errorf("invalid array index %q", rawIndex)
			}

			tokens = append(tokens, pathToken{index: index, isIndex: true})
			i += end + 1
		default:
			start := i
			for i < len(path) && path[i] != '.' && path[i] != '[' {
				i++
			}

			key := path[start:i]
			if key == "" {
				return nil, fmt.Errorf("invalid path %q", path)
			}

			tokens = append(tokens, pathToken{key: key})
		}
	}

	return tokens, nil
}

func normalize(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[key] = normalize(item)
		}
		return out
	case map[any]any:
		out := make(map[string]any, len(typed))
		for key, item := range typed {
			out[fmt.Sprint(key)] = normalize(item)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, item := range typed {
			out[i] = normalize(item)
		}
		return out
	default:
		return typed
	}
}
