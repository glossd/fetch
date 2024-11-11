package fetch

import (
	"github.com/glossd/fetch/internal/json"
)

// Stringify tries to marshal v.
// If an error happens, Stringify returns an empty string.
// An empty string is not a valid JSON, indicating Stringify failed.
//func Stringify(v any) string {
//	s, err := StringifySafe(v)
//	if err != nil {
//		return ""
//	}
//	return s
//}

// StringifySafe tries to fix possible errors during marshalling and then calls Marshal.
func StringifySafe(v any) (string, error) {
	//todo skip unsupported types e.g. channel fields
	// I can't rely on Go's encoding/json to escape unsupported fields.
	return Marshal(v)
}

// Marshal calls the patched json.Marshal function.
// The only difference between them is it lowercases
// the first letter of the public struct fields.
func Marshal(v any) (string, error) {
	rs, err := json.Marshal(v)
	return string(rs), err
}
