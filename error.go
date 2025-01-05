package fetch

import (
	"net/http"
)

type Error struct {
	inner   error
	Msg     string
	Status  int
	Headers map[string]string
	Body    string
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

func httpErr(prefix string, err error, r *http.Response, body []byte) *Error {
	if r == nil {
		return nonHttpErr(prefix, err)
	}
	return &Error{
		inner:   err,
		Msg:     prefix + err.Error(),
		Status:  r.StatusCode,
		Headers: mapFlatten(r.Header),
		Body:    string(body),
	}
}
