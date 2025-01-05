package fetch

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"
)

const defaultRespondErrorFormat = `{"error":"%s"}`

var respondErrorFormat = defaultRespondErrorFormat
var isRespondErrorFormatJSON = true

// SetHandlerErrorFormat is a global setter configuring how ToHandlerFunc converts errors returned from ApplyFunc.
// format argument must contain only one %s verb which would be the error message.
// Defaults to {"error":"%s"}
// Examples:
// fetch.SetHandlerErrorFormat(`{"msg":"%s"}`)
// fetch.SetHandlerErrorFormat("%s") - just plain error text
// fetch.SetHandlerErrorFormat(`{"error":{"message":"%s"}}`)
func SetHandlerErrorFormat(format string) {
	spl := strings.Split(format, "%s")
	if len(spl) < 2 {
		panic("RespondErrorFormat does not have '%s'")
	}
	if len(spl) > 2 {
		panic("RespondErrorFormat has more than one '%s'")
	}
	_, err := Unmarshal[J](format)
	isRespondErrorFormatJSON = err == nil
	respondErrorFormat = format
}

type respondConfig struct {
	// HTTP response status. Defaults to 200.
	Status int
	// Additional HTTP response headers.
	Headers map[string]string
	// HTTP response status in case of an error. Defaults to 500.
	ErrorStatus int
}

// respond tries to marshal the body and send HTTP response.
// The HTTP response is sent even if an error occurs.
// It should be used for the standard HTTP handlers.
func respond(w http.ResponseWriter, body any, config ...respondConfig) error {
	var cfg respondConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Status == 0 {
		cfg.Status = 200
	}
	if cfg.ErrorStatus == 0 {
		cfg.ErrorStatus = 500
	}
	if isResponseWrapper(body) {
		wrapper := reflect.ValueOf(body)
		status := wrapper.FieldByName("Status").Int()
		cfg.Status = int(status)
		mapRange := wrapper.FieldByName("Headers").MapRange()
		headers := make(map[string]string)
		for mapRange.Next() {
			headers[mapRange.Key().String()] = mapRange.Value().String()
		}
		cfg.Headers = headers
	}
	var err error
	if !isValidHTTPStatus(cfg.Status) {
		err := fmt.Errorf("respondConfig.Status is invalid")
		_ = doRespond(w, 500, fmt.Sprintf(respondErrorFormat, err), isRespondErrorFormatJSON, cfg)
		return err
	}
	if !isValidHTTPStatus(cfg.ErrorStatus) {
		err := fmt.Errorf("respondConfig.ErrorStatus is invalid")
		_ = doRespond(w, 500, fmt.Sprintf(respondErrorFormat, err), isRespondErrorFormatJSON, cfg)
		return err
	}
	var bodyStr string
	var isString = true
	if body != nil {
		switch u := body.(type) {
		case string:
			bodyStr = u
		case []byte:
			bodyStr = string(u)
		case Empty, *Empty, Response[Empty]:
			bodyStr = ""
		default:
			isString = false
		}
	}
	if !isString {
		if isResponseWrapper(body) {
			bodyStr, err = Marshal(reflect.ValueOf(body).FieldByName("Body").Interface())
		} else {
			bodyStr, err = Marshal(body)
		}
		if err != nil {
			_ = doRespond(w, cfg.ErrorStatus, fmt.Sprintf(respondErrorFormat, err), isRespondErrorFormatJSON, cfg)
			return fmt.Errorf("failed to marshal response body: %s", err)
		}
	}

	return doRespond(w, cfg.Status, bodyStr, !isString, cfg)
}

func doRespond(w http.ResponseWriter, status int, bodyStr string, isJSON bool, cfg respondConfig) error {
	w.WriteHeader(status)
	for k, v := range cfg.Headers {
		w.Header().Set(k, v)
	}
	if isJSON {
		w.Header().Set("Content-Type", "application/json")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
	_, err := w.Write([]byte(bodyStr))
	return err
}

// respondError sends HTTP response in the error format of respond.
// It should be used when your handler experiences an error
// before marshalling and responding with fetch.respond.
func respondError(w http.ResponseWriter, status int, errToRespond error, config ...respondConfig) error {
	var cfg respondConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	if !isValidHTTPStatus(status) {
		rerr := respondError(w, 500, errToRespond, config...)
		if rerr != nil {
			return rerr
		}
		return fmt.Errorf("error status is invalid")
	}
	w.WriteHeader(status)
	for k, v := range cfg.Headers {
		w.Header().Set(k, v)
	}
	if isRespondErrorFormatJSON {
		w.Header().Set("Content-Type", "application/json")
	} else {
		w.Header().Set("Content-Type", "text/plain")
	}
	bodyStr := fmt.Sprintf(respondErrorFormat, errToRespond.Error())
	_, err := w.Write([]byte(bodyStr))
	return err
}

func isValidHTTPStatus(status int) bool {
	return status >= 100 && status <= 599
}
