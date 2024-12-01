package fetch

import (
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
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

// ApplyFunc represents a simple function to be converted to http.Handler with
// In type as a request body and Out type as a response body.
type ApplyFunc[In any, Out any] func(in In) (Out, error)

/*
ToHandlerFunc converts ApplyFunc into http.HandlerFunc,
which can be used later in http.ServeMux#HandleFunc.
It unmarshals the HTTP request body into the ApplyFunc argument and
then marshals the returned value into the HTTP response body.
To insert PathValue into a field of the unmarshaled entity, specify `pathval` tag
to match the pattern's wildcard:

	type Pet struct {
		Id int `pathval:"id"`
	}

`header` tag can be used to insert HTTP headers into struct field.

	type Pet struct {
		Content string `header:"Content-Type"`
	}
*/
func ToHandlerFunc[In any, Out any](apply ApplyFunc[In, Out]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := defaultHandlerConfig
		if cfg.Middleware(w, r) {
			return
		}
		var in In
		if reflect.TypeFor[In]() != reflect.TypeOf(Empty{}) {
			reqBody, err := io.ReadAll(r.Body)
			if err != nil {
				cfg.ErrorHook(err)
				err = RespondError(w, 400, err)
				if err != nil {
					cfg.ErrorHook(err)
				}
				return
			}
			if len(reqBody) > 0 || shouldValidateInput(in) {
				in, err = Unmarshal[In](string(reqBody))
				if err != nil {
					cfg.ErrorHook(fmt.Errorf("failed to unmarshal request body: %s", err))
					err = RespondError(w, 400, err)
					if err != nil {
						cfg.ErrorHook(err)
					}
					return
				}
			}
		}
		in = enrichEntity(in, r)
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

// Input entity just might have a field with pathval tag
// and nothing else, we don't need to unmarshal it.
// In case it has some untagged fields, then it must be validated.
func shouldValidateInput(v any) bool {
	if t, ok := isStructType(v); ok {
		for i := 0; i < t.NumField(); i++ {
			tag := t.Field(i).Tag
			if tag.Get(headerTag) == "" && tag.Get(pathvalTag) == "" {
				return true
			}
		}
		return false
	} else {
		return false
	}
}

func isStructType(v any) (reflect.Type, bool) {
	typeOf := reflect.TypeOf(v)
	switch typeOf.Kind() {
	case reflect.Pointer:
		t := reflect.ValueOf(v).Elem().Type()
		return t, t.Kind() == reflect.Struct
	case reflect.Struct:
		return typeOf, true
	default:
		return typeOf, false
	}
}

func enrichEntity[T any](entity T, r *http.Request) T {
	typeOf, ok := isStructType(entity)
	if !ok {
		return entity
	}
	var elem reflect.Value
	if reflect.TypeOf(entity).Kind() == reflect.Pointer {
		elem = reflect.ValueOf(entity).Elem()
	} else { // struct
		elem = reflect.ValueOf(&entity).Elem()
	}
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if header := field.Tag.Get(headerTag); header != "" {
			elem.Field(i).SetString(r.Header.Get(header))
		}
		if pathval := field.Tag.Get(pathvalTag); pathval != "" {
			pathvar := r.PathValue(pathval)
			if pathvar != "" {
				switch field.Type.Kind() {
				case reflect.String:
					elem.Field(i).SetString(pathvar)
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					valInt64, err := strconv.ParseInt(pathvar, 10, 64)
					if err != nil {
						continue
					}
					elem.Field(i).SetInt(valInt64)
				}
			}
		}
	}
	return entity
}
