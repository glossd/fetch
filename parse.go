package fetch

import (
	"fmt"
	"github.com/glossd/fetch/internal/json"
	"reflect"
)

// todo
//// Instead of panicking skips any invalid fields.
//func Parse[T any](j string) T {
//	var t T
//	return t
//}

// Unmarshal is a generic wrapper for UnmarshalInto
func Unmarshal[T any](j string) (T, error) {
	var t T
	err := UnmarshalInto(j, &t)
	return t, err
}

// UnmarshalInto calls the patched json.Unmarshal function.
// The only difference between them is it handles `fetch.J`
// and transforms `any` into fetch.J.
func UnmarshalInto(j string, v any) error {

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &json.InvalidUnmarshalError{reflect.TypeOf(v)}
	}

	rve := rv.Elem()
	var isAny = rve.Kind() == reflect.Interface && rve.NumMethod() == 0
	if isAny || rve.Type() == typeFor[J]() {
		var a any
		err := json.Unmarshal([]byte(j), &a)
		if err != nil {
			return err
		}
		switch u := a.(type) {
		case bool:
			rve.Set(reflect.ValueOf(B(u)))
		case float64:
			rve.Set(reflect.ValueOf(F(u)))
		case map[string]any:
			rve.Set(reflect.ValueOf(M(u)))
		case []any:
			rve.Set(reflect.ValueOf(A(u)))
		default:
			return fmt.Errorf("glossd/fetch: unmarshal unexpected type: %T", a)
		}
		return nil
	}
	return json.Unmarshal([]byte(j), v)
}
