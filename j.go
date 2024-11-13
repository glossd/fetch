package fetch

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// J represents arbitrary JSON.
// Depending on the JSON data type the queried `fetch.J` could be one of these types
//
// | Type      | Go definition   | JSON data type                      |
// |-----------|-----------------|-------------------------------------|
// | fetch.M   | map[string]any  | object                              |
// | fetch.A   | []any           | array                               |
// | fetch.F   | float64         | number                              |
// | fetch.S   | string          | string                              |
// | fetch.B   | bool            | boolean                             |
// | fetch.Nil | (nil) *struct{} | null, undefined, anything not found |
type J interface {
	// Q parses jq-like patterns and returns according to the path value.
	// E.g.
	//{
	//  "name": "Jason",
	//  "category": {
	//    "name":"dogs"
	//  }
	//  "tags": [{"name":"briard"}]
	//}
	//
	// Whole json:  fmt.Println(j) or fmt.Println(j.Q("."))
	// Retrieve name: j.Q(".name")
	// Retrieve category name: j.Q(".category.name")
	// Retrieve first tag's name: j.Q(".tags[0].name")
	// If the value wasn't found, instead of nil value it will return Nil.
	// if the pattern syntax is invalid, it returns JQError.
	Q(pattern string) J

	// String returns JSON formatted string.
	String() string

	// Raw converts the value to its definition and returns it.
	// Type-Definitions:
	// M -> map[string]any
	// A -> []any
	// F -> float64
	// S -> string
	// B -> bool
	// Nil -> nil
	Raw() any

	// AsObject is a convenient type assertion if the underlying value holds a map[string]any.
	AsObject() (map[string]any, bool)
	// AsArray is a convenient type assertion if the underlying value holds a slice of type []any.
	AsArray() ([]any, bool)
	// AsNumber is a convenient type assertion if the underlying value holds a float64.
	AsNumber() (float64, bool)
	// AsString is a convenient type assertion if the underlying value holds a string.
	AsString() (string, bool)
	// AsBoolean is a convenient type assertion if the underlying value holds a bool.
	AsBoolean() (bool, bool)
	// IsNil check if the underlying value is fetch.Nil
	IsNil() bool
}

// M represents a JSON object.
type M map[string]any

func (m M) Q(pattern string) J {
	if strings.HasPrefix(pattern, ".") {
		pattern = pattern[1:]
	}
	if pattern == "" {
		return m
	}
	i, sep := nextSep(pattern)
	if i < 0 {
		return convert(m[pattern])
	}

	key, remaining := pattern[:i], pattern[i:]

	v, ok := m[key]
	if !ok {
		return jnil
	}

	return parseValue(v, remaining, sep)
}

func (m M) String() string {
	return marshalJ(m)
}

func (m M) Raw() any {
	return map[string]any(m)
}

func (m M) AsObject() (map[string]any, bool) { return m, true }
func (m M) AsArray() ([]any, bool)           { return nil, false }
func (m M) AsNumber() (float64, bool)        { return 0, false }
func (m M) AsString() (string, bool)         { return "", false }
func (m M) AsBoolean() (bool, bool)          { return false, false }
func (m M) IsNil() bool                      { return false }

// A represents a JSON array
type A []any

func (a A) Q(pattern string) J {
	if strings.HasPrefix(pattern, ".") {
		pattern = pattern[1:]
	}
	if pattern == "" {
		return a
	}

	if pattern[0] != '[' {
		// expected array index, got object key
		return jnil
	}
	closeBracket := strings.Index(pattern, "]")
	if closeBracket == -1 {
		return jqerr("expected ] for array index")
	}
	indexStr := pattern[1:closeBracket]
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return jqerr("expected a number for array index, got: '%s'", indexStr)
	}
	if index < 0 || index >= len(a) {
		// index out of range
		return jnil
	}

	v := a[index]
	remaining := pattern[closeBracket+1:]
	if remaining == "" {
		return convert(v)
	}
	i, sep := nextSep(remaining)
	if i != 0 {
		return jqerr("expected . or [, got: '%s'", beforeSep(remaining))
	}

	return parseValue(v, remaining, sep)
}

func (a A) String() string {
	return marshalJ(a)
}

func (a A) Raw() any {
	return []any(a)
}

func (a A) AsObject() (map[string]any, bool) { return nil, false }
func (a A) AsArray() ([]any, bool)           { return a, true }
func (a A) AsNumber() (float64, bool)        { return 0, false }
func (a A) AsString() (string, bool)         { return "", false }
func (a A) AsBoolean() (bool, bool)          { return false, false }
func (a A) IsNil() bool                      { return false }

func parseValue(v any, remaining string, sep string) J {
	if j, ok := v.(J); ok {
		return j.Q(remaining)
	}

	if sep == "." {
		return convert(v).Q(remaining)
	}
	if sep == "[" {
		arr, ok := v.([]any)
		if !ok {
			// expected an array
			return jnil
		}
		return A(arr).Q(remaining)
	}
	panic("glossd/fetch panic, please report to github: array only expected . or [ ")
}

// F represents JSON number.
type F float64

func (f F) Q(pattern string) J {
	if pattern == "" {
		return f
	}
	if pattern == "." {
		return f
	}
	return jnil
}

func (f F) String() string {
	return strconv.FormatFloat(float64(f), 'f', -1, 64)
}

func (f F) Raw() any {
	return float64(f)
}

func (f F) AsObject() (map[string]any, bool) { return nil, false }
func (f F) AsArray() ([]any, bool)           { return nil, false }
func (f F) AsNumber() (float64, bool)        { return float64(f), true }
func (f F) AsString() (string, bool)         { return "", false }
func (f F) AsBoolean() (bool, bool)          { return false, false }
func (f F) IsNil() bool                      { return false }

// S represents JSON string.
type S string

func (s S) Q(pattern string) J {
	if pattern == "" {
		return s
	}
	if pattern == "." {
		return s
	}
	return jnil
}

func (s S) String() string {
	return string(s)
}

func (s S) Raw() any {
	return string(s)
}

func (s S) AsObject() (map[string]any, bool) { return nil, false }
func (s S) AsArray() ([]any, bool)           { return nil, false }
func (s S) AsNumber() (float64, bool)        { return 0, false }
func (s S) AsString() (string, bool)         { return string(s), true }
func (s S) AsBoolean() (bool, bool)          { return false, false }
func (s S) IsNil() bool                      { return false }

// B represents a JSON boolean
type B bool

func (b B) Q(pattern string) J {
	if pattern == "" {
		return b
	}
	if pattern == "." {
		return b
	}
	return jnil
}

func (b B) String() string {
	return strconv.FormatBool(bool(b))
}

func (b B) Raw() any {
	return bool(b)
}

func (b B) AsObject() (map[string]any, bool) { return nil, false }
func (b B) AsArray() ([]any, bool)           { return nil, false }
func (b B) AsNumber() (float64, bool)        { return 0, false }
func (b B) AsString() (string, bool)         { return "", false }
func (b B) AsBoolean() (bool, bool)          { return bool(b), true }
func (b B) IsNil() bool                      { return false }

type nilStruct struct{}

// Nil represents any not found or null values. The pointer's value is always nil.
// However, when returned from any method, it doesn't equal nil, because
// a Go interface is not nil when it has a type.
// It exists to prevent nil pointer dereference when retrieving Raw value.
// It can be the root in J tree, because null alone is a valid JSON.
type Nil = *nilStruct

// the single instance of Nil.
var jnil Nil

func (n Nil) Q(string) J {
	return n
}

func (n Nil) String() string {
	return "nil"
}

func (n Nil) Raw() any {
	return nil
}

func (n Nil) AsObject() (map[string]any, bool) { return nil, false }
func (n Nil) AsArray() ([]any, bool)           { return nil, false }
func (n Nil) AsNumber() (float64, bool)        { return 0, false }
func (n Nil) AsString() (string, bool)         { return "", false }
func (n Nil) AsBoolean() (bool, bool)          { return false, false }
func (n Nil) IsNil() bool                      { return true }

func isJNil(v any) bool {
	return v == nil || reflect.TypeOf(v) == typeFor[Nil]()
}

func nextSep(pattern string) (int, string) {
	dot := strings.Index(pattern, ".")
	bracket := strings.Index(pattern, "[")
	if dot == -1 && bracket == -1 {
		return -1, ""
	}
	if dot == -1 {
		return bracket, "["
	}
	if bracket == -1 {
		return dot, "."
	}
	if dot < bracket {
		return dot, "."
	} else {
		return bracket, "["
	}
}

func beforeSep(pattern string) string {
	i, _ := nextSep(pattern)
	if i == -1 {
		return pattern
	} else {
		return pattern[:i]
	}
}

func convert(v any) J {
	if v == nil {
		return jnil
	}
	switch t := v.(type) {
	case bool:
		return B(t)
	case float64:
		return F(t)
	case string:
		return S(t)
	case map[string]any:
		return M(t)
	case []any:
		return A(t)
	default:
		panic(fmt.Sprintf("unexpected fetch.J type: %T", v))
	}
}

func marshalJ(v any) string {
	r, err := Marshal(v)
	if err != nil {
		// shouldn't happen, A and M are marshalable.
		return err.Error()
	}
	return r
}
