package fetch

import (
	"fmt"
	"strconv"
	"strings"
)

var jnil Nil

type J interface {
	// Q parses JQ-like patterns and returns according to the path value.
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

// F represents a JSON number
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

// S can't be a root value.
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

type nilStruct struct{}

// Nil represents any not found value.  The pointer's value is always nil.
// It exists to prevent nil pointer dereference when retrieving Raw value.
// Nil can't be a root value.
type Nil = *nilStruct

func (n Nil) Q(string) J {
	return n
}

func (n Nil) String() string {
	return "nil"
}

func (n Nil) Raw() any {
	return nil
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
		return err.Error()
	}
	return r
}
