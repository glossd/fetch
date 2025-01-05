package fetch

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

type mockWriter struct {
	status int
	header http.Header
	body   string
}

func newMockWriter() *mockWriter {
	return &mockWriter{
		status: 0,
		header: http.Header{},
		body:   "",
	}
}

func (mw *mockWriter) WriteHeader(status int) {
	mw.status = status
}

func (mw *mockWriter) Header() http.Header {
	return mw.header
}

func (mw *mockWriter) Write(b []byte) (int, error) {
	mw.body = string(b)
	return 0, nil
}

func TestRespond_String(t *testing.T) {
	mw := newMockWriter()
	err := respond(mw, "hello")
	assert(t, err, nil)
	assert(t, mw.status, 200)
	assert(t, mw.Header().Get("Content-Type"), "text/plain")
	assert(t, strings.TrimSpace(string(mw.body)), "hello")
}

func TestRespond_Struct(t *testing.T) {
	mw := newMockWriter()
	type TestStruct struct {
		Id string
	}
	err := respond(mw, &TestStruct{Id: "my-id"})
	assert(t, err, nil)
	assert(t, mw.status, 200)
	assert(t, mw.Header().Get("Content-Type"), "application/json")
	assert(t, strings.TrimSpace(mw.body), `{"id":"my-id"}`)
}

func TestRespond_InvalidJSON(t *testing.T) {
	mw := newMockWriter()
	type TestStruct struct {
		MyChan chan string
	}
	err := respond(mw, &TestStruct{})
	assertNotNil(t, err)
	assert(t, mw.status, 500)
	assert(t, mw.Header().Get("Content-Type"), "application/json")
	assert(t, strings.HasPrefix(mw.body, `{"error":`), true)
}

func TestRespond_InvalidStatus(t *testing.T) {
	mw := newMockWriter()
	err := respond(mw, "hello", RespondConfig{Status: 51})
	assertNotNil(t, err)
	assert(t, mw.status, 500)
	assert(t, mw.Header().Get("Content-Type"), "application/json")
	assert(t, strings.HasPrefix(mw.body, `{"error":`), true)
}

func TestSetRespondErrorFormat(t *testing.T) {
	defer func() {
		SetRespondErrorFormat(defaultRespondErrorFormat)
	}()

	SetRespondErrorFormat("%s")
	mw := newMockWriter()
	respond(mw, "hello", RespondConfig{Status: 51})
	assert(t, mw.header.Get("Content-Type"), "text/plain")
	assert(t, mw.body, "RespondConfig.Status is invalid")
}

func TestSetRespondErrorFormat_InvalidFormats(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		defer func() {
			SetRespondErrorFormat(defaultRespondErrorFormat)
			if r := recover(); r != nil {
				assert(t, fmt.Sprintf("%s", r), "RespondErrorFormat does not have '%s'")
			}
		}()
		SetRespondErrorFormat("")
	})
	t.Run(`double %s`, func(t *testing.T) {
		defer func() {
			SetRespondErrorFormat(defaultRespondErrorFormat)
			if r := recover(); r != nil {
				assert(t, fmt.Sprintf("%s", r), "RespondErrorFormat has more than one '%s'")
			}
		}()
		SetRespondErrorFormat(`{"%s":"%s"}`)
	})
}

func TestRespondResponseEmpty(t *testing.T) {
	mw := newMockWriter()
	err := respond(mw, Response[Empty]{Status: 204})
	assert(t, err, nil)
	if mw.status != 204 || len(mw.body) > 0 {
		t.Errorf("wrong writer: %+v", mw)
	}
}

func TestRespondResponse(t *testing.T) {
	type Pet struct {
		Name string
	}
	mw := newMockWriter()
	err := respond(mw, Response[Pet]{Status: 201, Body: Pet{Name: "Lola"}})
	assert(t, err, nil)
	if mw.status != 201 || mw.body != `{"name":"Lola"}` {
		t.Errorf("wrong writer: %+v, %s", mw, mw.body)
	}
}

func TestRespondError(t *testing.T) {
	mw := newMockWriter()
	err := respondError(mw, 400, fmt.Errorf("wrong"))
	assert(t, err, nil)
	assert(t, mw.status, 400)
	assert(t, mw.Header().Get("Content-Type"), "application/json")
	assert(t, mw.body, `{"error":"wrong"}`)
}

func assert[T comparable](t *testing.T, got, want T) {
	t.Helper()

	if got != want {
		t.Fatalf("got %v, wanted %v", got, want)
	}
}

func assertNotNil(t *testing.T, v any) {
	t.Helper()
	if v == nil {
		t.Fatalf("expected nil, but got %v", v)
	}
}
