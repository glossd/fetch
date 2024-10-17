package fetch

import (
	"github.com/glossd/fetch/internal/json"
)

// todo
// Instead of panicking skips any invalid types or fields.
//func Stringify(v any) string {
//	return ""
//}

func StringifySafe(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	if s, ok := v.([]byte); ok {
		return string(s), nil
	}
	return Marshal(v)
}

// Marshal calls the patched json.Marshal function.
// The only difference between them is it lowercases
// the first letter of the public struct fields.
func Marshal(v any) (string, error) {
	rs, err := json.Marshal(v)
	return string(rs), err
}
