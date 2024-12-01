package fetch

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

var httpClient = &http.Client{}
var baseURL = ""

type Config struct {
	// Defaults to context.Background()
	Ctx context.Context
	// Defaults to GET
	Method  string
	Body    string
	Headers map[string]string
}

func Get[T any](url string, config ...Config) (T, error) {
	if len(config) == 0 {
		config = []Config{{}}
	}
	config[0].Method = http.MethodGet
	return Request[T](url, config...)
}

// GetJ is a wrapper for Get[fetch.J]
func GetJ(url string, config ...Config) (J, error) {
	return Get[J](url, config...)
}

func Post[T any](url string, body any, config ...Config) (T, error) {
	return requestWithBody[T](url, http.MethodPost, body, config...)
}

func Put[T any](url string, body any, config ...Config) (T, error) {
	return requestWithBody[T](url, http.MethodPut, body, config...)
}

func Patch[T any](url string, body any, config ...Config) (T, error) {
	return requestWithBody[T](url, http.MethodPatch, body, config...)
}

func requestWithBody[T any](url string, method string, body any, config ...Config) (T, error) {
	if len(config) == 0 {
		config = []Config{{}}
	}
	config[0].Method = method
	b, err := bodyToString(body)
	if err != nil {
		var t T
		return t, nonHttpErr("invalid body: ", err)
	}
	config[0].Body = b
	return Request[T](url, config...)
}

func bodyToString(v any) (string, error) {
	if s, ok := v.(string); ok {
		return s, nil
	}
	if s, ok := v.([]byte); ok {
		return string(s), nil
	}
	return Marshal(v)
}

func Delete[T any](url string, config ...Config) (T, error) {
	if len(config) == 0 {
		config = []Config{{}}
	}
	config[0].Method = http.MethodDelete
	return Request[T](url, config...)
}

func Head[T any](url string, config ...Config) (T, error) {
	if len(config) == 0 {
		config = []Config{{}}
	}
	config[0].Method = http.MethodHead
	return Request[T](url, config...)
}

func Options[T any](url string, config ...Config) (T, error) {
	if len(config) == 0 {
		config = []Config{{}}
	}
	config[0].Method = http.MethodOptions
	return Request[T](url, config...)
}

func Request[T any](url string, config ...Config) (T, error) {
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}

	if cfg.Ctx == nil {
		cfg.Ctx = context.Background()
	}
	if cfg.Method == "" {
		cfg.Method = "GET"
	}
	if cfg.Headers == nil || !hasContentType(cfg) {
		if cfg.Headers == nil {
			cfg.Headers = make(map[string]string, 1)
		}
		cfg.Headers["Content-type"] = "application/json"
	}

	fullURL := baseURL + url
	if hasProtocol(url) {
		fullURL = url
	}
	if !hasProtocol(fullURL) {
		if strings.HasPrefix(fullURL, "localhost") {
			fullURL = "http://" + fullURL
		} else {
			fullURL = "https://" + fullURL
		}
	}

	req, err := http.NewRequest(cfg.Method, fullURL, bytes.NewBuffer([]byte(cfg.Body)))
	if err != nil {
		var t T
		return t, nonHttpErr("invalid request: ", err)
	}

	req = req.WithContext(cfg.Ctx)

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}

	var res *http.Response
	if mock {
		res = mockDNS(url, req).response()
	} else {
		res, err = httpClient.Do(req)
		if err != nil {
			var t T
			return t, nonHttpErr("failed request: ", err)
		}
	}

	defer func() {
		if res != nil && res.Body != nil {
			// the body needs to be closed even it wasn't read.
			err := res.Body.Close()
			if err != nil {
				fmt.Println(fmt.Sprintf("resource leak: fetch %s failed to close the response body: %s", req.URL.String(), err.Error()))
			}
		}
	}()

	var t T
	typeOf := reflect.TypeOf(t)

	if typeOf != nil && typeOf == reflect.TypeFor[Empty]() && firstDigit(res.StatusCode) == 2 {
		return t, nil
	}
	if typeOf != nil && typeOf == reflect.TypeFor[ResponseEmpty]() && firstDigit(res.StatusCode) == 2 {
		re := any(&t).(*ResponseEmpty)
		re.Status = res.StatusCode
		re.Headers = uniqueHeaders(res.Header)
		re.DuplicateHeaders = res.Header
		return t, nil
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return t, httpErr("read http body: ", err, res, nil)
	}

	if firstDigit(res.StatusCode) != 2 {
		return t, httpErr(fmt.Sprintf("http status=%d, body=", res.StatusCode), errors.New(string(body)), res, body)
	}

	if typeOf != nil && typeOf.PkgPath() == "github.com/glossd/fetch" && strings.HasPrefix(typeOf.Name(), "Response[") {
		resType, ok := typeOf.FieldByName("Body")
		if !ok {
			panic("field Body is not found in Response")
		}

		resInstance := reflect.New(resType.Type).Interface()
		err = parseBodyInto(body, resInstance)
		if err != nil {
			var t T
			return t, httpErr("parsing JSON error: ", err, res, body)
		}

		valueOf := reflect.Indirect(reflect.ValueOf(&t))
		valueOf.FieldByName("Status").SetInt(int64(res.StatusCode))
		valueOf.FieldByName("DuplicateHeaders").Set(reflect.ValueOf(res.Header))
		valueOf.FieldByName("Headers").Set(reflect.ValueOf(uniqueHeaders(res.Header)))
		valueOf.FieldByName("BodyBytes").SetBytes(body)
		valueOf.FieldByName("Body").Set(reflect.ValueOf(resInstance).Elem())

		return t, nil
	}
	err = parseBodyInto(body, &t)
	if err != nil {
		var t T
		return t, httpErr("parsing JSON error: ", err, res, body)
	}
	return t, nil
}

func hasProtocol(url string) bool {
	return strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
}

func parseBodyInto(body []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return &json.InvalidUnmarshalError{Type: reflect.TypeOf(v)}
	}
	rve := rv.Elem()
	if rve.Kind() == reflect.String {
		rve.SetString(string(body))
		return nil
	}
	if rve.Kind() == reflect.Slice && rve.Type().Elem().Kind() == reflect.Uint8 {
		rve.SetBytes(body)
		return nil
	}
	err := UnmarshalInto(string(body), v)
	if err != nil {
		return err
	}
	return nil
}

func SetHttpClient(c *http.Client) {
	if c == nil {
		return
	}
	httpClient = c
}

func SetBaseURL(b string) {
	baseURL = b
}

func firstDigit(n int) int {
	var i int
	for i = n; i >= 10; i = i / 10 {
	}
	return i
}

func hasContentType(c Config) bool {
	for k := range c.Headers {
		if strings.ToLower(k) == "content-type" {
			return true
		}
	}
	return false
}
