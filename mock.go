package fetch

import (
	"bytes"
	"io"
	"net/http"
)

var mock bool

type mockResponse struct {
	Status  int
	Headers map[string][]string
	Body    string
}

type bufferClose struct {
	b *bytes.Buffer
}

func newBC(str string) *bufferClose {
	return &bufferClose{b: bytes.NewBuffer([]byte(str))}
}

func (bc *bufferClose) Read(p []byte) (n int, err error) {
	return bc.b.Read(p)
}

func (bc *bufferClose) Close() error {
	return nil
}

func mockDNS(url string, req *http.Request) *mockResponse {
	switch url {
	case "key.value":
		return &mockResponse{Status: 200, Headers: map[string][]string{"Content-type": {"application/json"}}, Body: `{"key":"value"}`}
	case "array.int":
		return &mockResponse{Status: 200, Body: `[1, 2, 3]`}
	case "my.ip":
		return &mockResponse{Status: 200, Headers: map[string][]string{"Content-type": {"text/plain"}}, Body: `8.8.8.8`}
	case "400.error":
		return &mockResponse{Status: 400, Headers: map[string][]string{"Content-type": {"text/plain"}}, Body: `Bad Request`}
	case "echo.me":
		body, err := io.ReadAll(req.Body)
		if err != nil {
			panic(err)
		}
		return &mockResponse{Status: 200, Body: string(body)}
	default:
		panic("mockDNS unknown url " + url)
	}
}

func (mr *mockResponse) response() *http.Response {
	return &http.Response{
		StatusCode:    mr.Status,
		Header:        mr.Headers,
		Body:          newBC(mr.Body),
		ContentLength: int64(len([]byte(mr.Body))),
	}
}
