package fetch

import (
	"os"
	"slices"
	"testing"
)

func TestMain(m *testing.M) {
	mock = true
	code := m.Run()
	os.Exit(code)
}

func TestDoString(t *testing.T) {
	res, err := Do[string]("my.ip")
	if err != nil {
		t.Fatal(err)
	}
	if res != "8.8.8.8" {
		t.Errorf("wrong ip")
	}
}

func TestDoBytes(t *testing.T) {
	res, err := Do[[]byte]("array.int")
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(res, []byte(`[1, 2, 3]`)) {
		t.Errorf("the bytes mistmatch, expected raw bytes from the body")
	}
}

func TestDoArray(t *testing.T) {
	res, err := Do[[]int]("array.int")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res, []int{1, 2, 3}) {
		t.Errorf("wrong array of integers")
	}
}

func TestDoAny(t *testing.T) {
	res, err := Do[any]("key.value")
	if err != nil {
		t.Fatal(err)
	}
	m, ok := res.(M)
	if !ok {
		t.Fatalf("not a map, got=%T", res)
	}
	if m["key"] != "value" {
		t.Errorf("map wasn't parsed")
	}

	res2, err := Do[any]("array.int")
	if err != nil {
		t.Fatal(err)
	}
	a, ok := res2.(A)
	if !ok {
		t.Fatalf("not an array, got=%T", res)
	}
	if a[0] != 1.0 {
		t.Errorf("array wasn't parsed")
	}
}

func TestDo_ResponseT(t *testing.T) {
	type TestStruct struct {
		Key string
	}
	res, err := Do[Response[TestStruct]]("key.value")
	if err != nil {
		t.Error(err)
	}
	if res.Status != 200 {
		t.Errorf("wrong status")
	}

	if res.Headers["Content-type"] != "application/json" {
		t.Errorf("wrong headers")
	}

	if res.Body.Key != "value" {
		t.Errorf("wrong body")
	}

	res2, err := Do[Response[string]]("my.ip")
	if err != nil {
		t.Fatal(err)
	}
	if res2.Body != "8.8.8.8" {
		t.Errorf("response string mismatch, got=%s", res2.Body)
	}
}

func TestDo_ResponseEmpty(t *testing.T) {
	res, err := Do[Response[Empty]]("key.value")
	if err != nil {
		t.Error(err)
	}
	if res.Status != 200 {
		t.Errorf("response status isn't 200")
	}
	if res.Headers["Content-type"] != "application/json" {
		t.Errorf("wrong headers")
	}

	_, err = Do[Response[Empty]]("400.error")
	if err == nil || err.(*Error).Body != "Bad Request" {
		t.Errorf("Even with ResponseEmpty error should read the body")
	}

	resEm, err := Do[ResponseEmpty]("key.value")
	if err != nil {
		t.Error(err)
	}
	if resEm.Status != 200 {
		t.Errorf("response status isn't 200")
	}
	if resEm.Headers["Content-type"] != "application/json" {
		t.Errorf("wrong headers")
	}
}

func TestDo_Error(t *testing.T) {
	_, err := Do[string]("400.error")
	if err == nil {
		t.Fatal(err)
	}

	castErr := err.(*Error)
	if castErr.Status != 400 {
		t.Errorf("expected status 400")
	}
	if castErr.Headers["Content-type"] != "text/plain" {
		t.Errorf("expected headers")
	}
	if castErr.Body != "Bad Request" {
		t.Errorf("expected body")
	}
}

func TestPostString(t *testing.T) {
	j, err := Post[M]("echo.me", `{"hello":"whosthere"}`)
	if err != nil {
		t.Fatal(err)
	}
	if j["hello"] != "whosthere" {
		t.Errorf("wrong post response")
	}
}

func TestPostBytes(t *testing.T) {
	j, err := Post[M]("echo.me", []byte(`{"hello":"whosthere"}`))
	if err != nil {
		t.Fatal(err)
	}
	if j["hello"] != "whosthere" {
		t.Errorf("wrong post response")
	}
}

func TestIssue20(t *testing.T) {
	var err error
	_, err = Get[J]("key.value")
	if err != nil {
		t.Errorf("err should be nil")
	}
}
