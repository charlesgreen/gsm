package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// decodeJSON into dst from the reader, normalizing snake_case keys to camelCase
//
// This exists because the GSM SDK in Go uses snake_case JSON struct tags, while
// their public documentation shows camelCase. We need to be compliant with their
// advertised schema, while also working for people using the official SDK in Go.
func decodeJSON[T any](r io.Reader, dst *T) error {
	var src map[string]any
	if err := json.NewDecoder(r).Decode(&src); err != nil {
		return err
	}

	normalized := normalizeKeys(src)
	b, err := json.Marshal(normalized)
	if err != nil {
		return fmt.Errorf("encoding normalized json: %w", err)
	}
	return json.Unmarshal(b, dst)
}

func normalizeKeys(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		parts := strings.Split(k, "_")
		if len(parts) == 0 {
			continue
		}

		var camelCase string
		for i, p := range parts {
			if i == 0 {
				camelCase = p
				continue
			}
			camelCase += strings.ToUpper(p[:1]) + p[1:]
		}

		switch vv := v.(type) {
		case map[string]any:
			v = normalizeKeys(vv)
		case []any:
			for i, inner := range vv {
				if m, ok := inner.(map[string]any); ok {
					vv[i] = normalizeKeys(m)
				}
			}
			v = vv
		}
		dst[camelCase] = v
	}
	return dst
}
