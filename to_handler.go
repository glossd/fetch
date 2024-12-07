package fetch

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type handleTag = string

const (
	pathvalTag handleTag = "pathval"
	headerTag  handleTag = "header"
)

var defaultHandlerConfig = HandlerConfig{
	ErrorHook: func(err error) {
		fmt.Printf("fetch.Handle failed to respond: %s\n", err)
	},
	Middleware: func(w http.ResponseWriter, r *http.Request) bool {
		return false
	},
}

// SetDefaultHandlerConfig sets HandlerConfig globally to be applied for every ToHandlerFunc.
func SetDefaultHandlerConfig(hc HandlerConfig) {
	if hc.ErrorHook == nil {
		hc.ErrorHook = defaultHandlerConfig.ErrorHook
	}
	if hc.Middleware == nil {
		hc.Middleware = defaultHandlerConfig.Middleware
	}
	defaultHandlerConfig = hc
}

type HandlerConfig struct {
	// ErrorHook is called if an error happens while sending an HTTP response
	ErrorHook func(err error)
	// Middleware is applied before ToHandlerFunc processes the request.
	// Return true to end the request processing.
	Middleware func(w http.ResponseWriter, r *http.Request) bool
}

func (cfg HandlerConfig) respondError(w http.ResponseWriter, err error) {
	cfg.ErrorHook(err)
	err = RespondError(w, 400, err)
	if err != nil {
		cfg.ErrorHook(err)
	}
}

// ApplyFunc represents a simple function to be converted to http.Handler with
// In type as a request body and Out type as a response body.
type ApplyFunc[In any, Out any] func(in In) (Out, error)

/*
ToHandlerFunc converts ApplyFunc into http.HandlerFunc,
which can be used later in http.ServeMux#HandleFunc.
It unmarshals the HTTP request body into the ApplyFunc argument and
then marshals the returned value into the HTTP response body.
To access HTTP request attributes, wrap your input in fetch.Request.
*/
func ToHandlerFunc[In any, Out any](apply ApplyFunc[In, Out]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := defaultHandlerConfig
		if cfg.Middleware(w, r) {
			return
		}
		var in In
		if isRequestWrapper(in) {
			typeOf := reflect.TypeOf(in)
			resType, ok := typeOf.FieldByName("Body")
			if !ok {
				panic("field Body is not found in Request")
			}
			resInstance := reflect.New(resType.Type).Interface()
			if !isEmptyType(resInstance) {
				reqBody, err := io.ReadAll(r.Body)
				if err != nil {
					cfg.respondError(w, err)
					return
				}
				err = parseBodyInto(reqBody, resInstance)
				if err != nil {
					cfg.respondError(w, fmt.Errorf("failed to parse request body: %s", err))
					return
				}
			}
			valueOf := reflect.Indirect(reflect.ValueOf(&in))
			valueOf.FieldByName("Context").Set(reflect.ValueOf(r.Context()))
			valueOf.FieldByName("PathValues").Set(reflect.ValueOf(extractPathValues(r)))
			valueOf.FieldByName("Headers").Set(reflect.ValueOf(uniqueHeaders(r.Header)))
			valueOf.FieldByName("Body").Set(reflect.ValueOf(resInstance).Elem())
		} else if !isEmptyType(in) {
			reqBody, err := io.ReadAll(r.Body)
			if err != nil {
				cfg.respondError(w, err)
				return
			}
			err = parseBodyInto(reqBody, &in)
			if err != nil {
				cfg.respondError(w, fmt.Errorf("failed to parse request body: %s", err))
				return
			}
		}

		out, err := apply(in)
		if err != nil {
			err = RespondError(w, 500, err)
			if err != nil {
				cfg.ErrorHook(err)
			}
			return
		}
		err = Respond(w, out)
		if err != nil {
			cfg.ErrorHook(err)
		}
	}
}

func extractPathValues(r *http.Request) map[string]string {
	parts := strings.Split(r.Pattern, "/")
	result := make(map[string]string)
	for _, part := range parts {
		if len(part) > 2 && strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			wildcard := part[1 : len(part)-1]
			v := r.PathValue(wildcard)
			if v != "" {
				result[wildcard] = v
			}
		}
	}
	return result
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
