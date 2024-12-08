package fetch

import (
	"context"
	"reflect"
	"strings"
)

/*
Response is a wrapper type for (generic) ReturnType to be used in
the HTTP methods. It allows you to access HTTP attributes
of the HTTP response and unmarshal the HTTP body.
e.g.

	type User struct {
		FirstName string
	}
	res, err := Get[Response[User]]("/users/1")
	if err != nil {panic(err)}
	if res.Status != 202 {
	   panic("unexpected status")
	}
	// Body is User type
	fmt.Println(res.Body.FirstName)
*/
type Response[T any] struct {
	Status  int
	Headers map[string]string
	Body    T
}

func uniqueHeaders(headers map[string][]string) map[string]string {
	h := make(map[string]string, len(headers))
	for key, val := range headers {
		if len(val) > 0 {
			// it takes the last element intentionally.
			h[key] = val[len(val)-1]
		}
	}
	return h
}

/*
Request can be used in ApplyFunc as a wrapper
for the input entity to access http attributes.
e.g.

	type Pet struct {
		Name string
	}
	http.HandleFunc("POST /pets/{id}", fetch.ToHandlerFunc(func(in fetch.Request[Pet]) (fetch.Empty, error) {
		in.Context()
		return fetch.Empty{}, nil
	}))
*/
type Request[T any] struct {
	Context context.Context
	// Only available in go1.23 and above.
	// PathValue was introduced in go1.22 but
	// there was no reliable way to extract them.
	// go1.23 introduced http.Request.Pattern allowing to list the wildcards.
	PathValues map[string]string
	Headers    map[string]string
	Body       T
}

// Empty represents an empty response or request body, skipping JSON handling.
// Can be used with the wrappers Response and Request or to fit the signature of ApplyFunc.
type Empty struct{}

func isResponseWrapper(v any) bool {
	if v == nil {
		return false
	}
	typeOf := reflect.TypeOf(v)
	return typeOf.PkgPath() == "github.com/glossd/fetch" && strings.HasPrefix(typeOf.Name(), "Response[")
}
func isResponseWithEmpty(v any) bool {
	return reflect.TypeOf(v) == reflectTypeFor[Response[Empty]]()
}

func isRequestWrapper(v any) bool {
	typeOf := reflect.TypeOf(v)
	return typeOf != nil && typeOf.PkgPath() == "github.com/glossd/fetch" && strings.HasPrefix(typeOf.Name(), "Request[")
}

func isEmptyType(v any) bool {
	st, ok := isStructType(v)
	if !ok {
		return false
	}
	return st == reflect.TypeOf(Empty{})
}

func isStructType(v any) (reflect.Type, bool) {
	typeOf := reflect.TypeOf(v)
	if v == nil {
		return typeOf, false
	}
	switch typeOf.Kind() {
	case reflect.Pointer:
		valueOf := reflect.ValueOf(v)
		if valueOf.IsNil() {
			return typeOf, false
		}
		t := reflect.ValueOf(v).Elem().Type()
		return t, t.Kind() == reflect.Struct
	case reflect.Struct:
		return typeOf, true
	default:
		return typeOf, false
	}
}

// reflect.TypeFor was introduced in go1.22
func reflectTypeFor[T any]() reflect.Type {
	var v T
	if t := reflect.TypeOf(v); t != nil {
		return t
	}
	return reflect.TypeOf((*T)(nil)).Elem()
}
