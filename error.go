package fetch

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
)

type Error struct {
	inner   error
	Msg     string
	Status  int
	Headers map[string]string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return e.Msg
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.inner
}

func nonHttpErr(prefix string, err error) *Error {
	return &Error{inner: err, Msg: prefix + err.Error()}
}

func httpErr(prefix string, err error, r *http.Response) *Error {
	if r == nil {
		return nonHttpErr(prefix, err)
	}
	return &Error{inner: err, Msg: prefix + err.Error(), Status: r.StatusCode, Headers: uniqueHeaders(r.Header)}
}

// JQError is returned from J.Q on invalid syntax.
// It seems to be better to return this than panic.
type JQError struct {
	s string
}

func (e *JQError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("JQError: %s", e.s)
}

func (e *JQError) Unwrap() error {
	if e == nil {
		return nil
	}
	return errors.New(e.s)
}

func (e *JQError) Q(pattern string) J { return e }

func (e *JQError) String() string {
	if e == nil {
		return ""
	}
	return e.Error()
}

func (e *JQError) Raw() any                         { return e }
func (e *JQError) AsObject() (map[string]any, bool) { return nil, false }
func (e *JQError) AsArray() ([]any, bool)           { return nil, false }
func (e *JQError) AsNumber() (float64, bool)        { return 0, false }
func (e *JQError) AsString() (string, bool)         { return "", false }
func (e *JQError) AsBoolean() (bool, bool)          { return false, false }
func (e *JQError) IsNil() bool                      { return false }

func jqerr(format string, a ...any) *JQError {
	return &JQError{s: fmt.Sprintf(format, a...)}
}

func IsJQError(v any) bool {
	return reflect.TypeOf(v) == typeFor[*JQError]()
}

// reflect.TypeFor was introduced in go1.22
func typeFor[T any]() reflect.Type {
	var v T
	if t := reflect.TypeOf(v); t != nil {
		return t
	}
	return reflect.TypeOf((*T)(nil)).Elem()
}
