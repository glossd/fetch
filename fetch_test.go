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

func TestRequestString(t *testing.T) {
	res, err := Request[string]("my.ip")
	if err != nil {
		t.Fatal(err)
	}
	if res != "8.8.8.8" {
		t.Errorf("wrong ip")
	}
}

func TestRequestBytes(t *testing.T) {
	res, err := Request[[]byte]("array.int")
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(res, []byte(`[1, 2, 3]`)) {
		t.Errorf("the bytes mistmatch, expected raw bytes from the body")
	}
}

func TestRequestArray(t *testing.T) {
	res, err := Request[[]int]("array.int")
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res, []int{1, 2, 3}) {
		t.Errorf("wrong array of integers")
	}
}

func TestRequestAny(t *testing.T) {
	res, err := Request[any]("key.value")
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

	res2, err := Request[any]("array.int")
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

func TestRequest_ResponseT(t *testing.T) {
	type TestStruct struct {
		Key string
	}
	res, err := Request[Response[TestStruct]]("key.value")
	if err != nil {
		t.Error(err)
	}
	if res.Status != 200 {
		t.Errorf("wrong status")
	}

	if res.Headers()["Content-type"] != "application/json" {
		t.Errorf("wrong headers")
	}

	if res.Body.Key != "value" {
		t.Errorf("wrong body")
	}

	res2, err := Request[Response[string]]("my.ip")
	if err != nil {
		t.Fatal(err)
	}
	if res2.Body != "8.8.8.8" {
		t.Errorf("response string mismatch, got=%s", res2.Body)
	}
}

func TestRequest_ResponseEmpty(t *testing.T) {
	res, err := Request[ResponseEmpty]("key.value")
	if err != nil {
		t.Error(err)
	}
	if res.Status != 200 {
		t.Errorf("response status isn't 200")
	}
	if res.Headers()["Content-type"] != "application/json" {
		t.Errorf("wrong headers")
	}

	_, err = Request[ResponseEmpty]("400.error")
	if err == nil || err.Body != "Bad Request" {
		t.Errorf("Even with ResponseEmpty error should read the body")
	}
}

func TestRequest_Error(t *testing.T) {
	_, err := Request[string]("400.error")
	if err == nil {
		t.Fatal(err)
	}

	if err.Status != 400 {
		t.Errorf("expected status 400")
	}
	if err.Headers["Content-type"] != "text/plain" {
		t.Errorf("expected headers")
	}
	if err.Body != "Bad Request" {
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
