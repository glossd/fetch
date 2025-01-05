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

//// StringifySafe tries to fix possible errors during marshalling and then calls Marshal.
//func StringifySafe(v any) (string, error) {
//	// skip unsupported types e.g. channel fields
//	// I can't rely on Go's encoding/json to escape unsupported fields.
//	return Marshal(v)
//}

/*
Marshal calls the patched json.Marshal function.
There are only two patches
 1. It lowercases the first letter of the public struct fields.
 2. It omits empty fields by default.

They are only applied if the `json` tag is not specified.
*/
func Marshal(v any) (string, error) {
	rs, err := json.Marshal(v)
	return string(rs), err
}
