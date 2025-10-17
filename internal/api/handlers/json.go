package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var skipNormalizationFor = map[string]struct{}{
	"labels":         {},
	"annotations":    {},
	"versionAliases": {},
}

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
		camelCase := toCamelCase(k)
		if _, ok := skipNormalizationFor[camelCase]; ok {
			dst[camelCase] = v
			continue
		}

		// Decide whether we need to run recursively for other objects or arrays of
		// objects
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

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	if len(parts) < 2 {
		return s
	}

	var camelCase string
	for _, p := range parts[1:] {
		if len(p) == 0 {
			continue
		}

		// The first segment written should stay naturally cased
		if camelCase == "" {
			camelCase += p
		} else {
			camelCase += strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return camelCase
}
