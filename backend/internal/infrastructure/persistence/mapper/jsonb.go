package mapper

import "encoding/json"

// StringSliceToJSON converts a string slice to a JSON string for JSONB columns.
func StringSliceToJSON(s []string) string {
	if s == nil {
		return "[]"
	}
	b, err := json.Marshal(s)
	if err != nil {
		return "[]"
	}
	return string(b)
}

// JSONToStringSlice converts a JSON string from a JSONB column to a string slice.
func JSONToStringSlice(s string) []string {
	if s == "" {
		return []string{}
	}
	var result []string
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return []string{}
	}
	return result
}
