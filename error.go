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

func (e *JQError) Q(pattern string) J {
	if e == nil {
		return S("")
	}
	return S(e.Error())
}

func (e *JQError) String() string {
	if e == nil {
		return ""
	}
	return e.Error()
}

func (e *JQError) Raw() any {
	return e
}

func jqerr(format string, a ...any) *JQError {
	return &JQError{s: fmt.Sprintf(format, a...)}
}

func IsJQError(v any) bool {
	return reflect.TypeOf(v) == reflect.TypeFor[*JQError]()
}
