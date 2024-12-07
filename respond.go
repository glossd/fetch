package fetch

import (
	"fmt"
	"net/http"
	"strings"
)

const defaultRespondErrorFormat = `{"error":"%s"}`

var respondErrorFormat = defaultRespondErrorFormat
var isRespondErrorFormatJSON = true

// SetRespondErrorFormat is a global setter, configuring how Respond sends the errors.
// format argument must contain only one %s verb which would be the error message.
// E.g.
// fetch.SetDefaultErrorRespondFormat(`{"msg":"%s"}`)
// fetch.SetDefaultErrorRespondFormat("%s") - just plain error text
// fetch.SetDefaultErrorRespondFormat(`{"error":{"message":"%s"}}`)
func SetRespondErrorFormat(format string) {
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

type RespondConfig struct {
	// HTTP response status. Defaults to 200.
	Status int
	// Additional HTTP response headers.
	Headers map[string]string
	// HTTP response status in case of an error. Defaults to 500.
	ErrorStatus int
}

// Deprecated, rely on ToHandlerFunc.
// Respond tries to marshal the body and send HTTP response.
// The HTTP response is sent even if an error occurs.
// It should be used for the standard HTTP handlers.
func Respond(w http.ResponseWriter, body any, config ...RespondConfig) error {
	var cfg RespondConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.Status == 0 {
		cfg.Status = 200
	}
	if cfg.ErrorStatus == 0 {
		cfg.ErrorStatus = 500
	}
	// todo handle ResponseEmpty, Response
	var err error
	if !isValidHTTPStatus(cfg.Status) {
		err := fmt.Errorf("RespondConfig.Status is invalid")
		_ = respond(w, 500, fmt.Sprintf(respondErrorFormat, err), isRespondErrorFormatJSON, cfg)
		return err
	}
	if !isValidHTTPStatus(cfg.ErrorStatus) {
		err := fmt.Errorf("RespondConfig.ErrorStatus is invalid")
		_ = respond(w, 500, fmt.Sprintf(respondErrorFormat, err), isRespondErrorFormatJSON, cfg)
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
		case Empty, *Empty:
			bodyStr = ""
		default:
			isString = false
		}
	}
	if !isString {
		bodyStr, err = Marshal(body)
		if err != nil {
			_ = respond(w, cfg.ErrorStatus, fmt.Sprintf(respondErrorFormat, err), isRespondErrorFormatJSON, cfg)
			return fmt.Errorf("failed to marshal response body: %s", err)
		}
	}

	return respond(w, cfg.Status, bodyStr, !isString, cfg)
}

func respond(w http.ResponseWriter, status int, bodyStr string, isJSON bool, cfg RespondConfig) error {
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

// Deprecated, rely on ToHandlerFunc.
// RespondError sends HTTP response in the error format of Respond.
// It should be used when your handler experiences an error
// before marshalling and responding with fetch.Respond.
func RespondError(w http.ResponseWriter, status int, errToRespond error, config ...RespondConfig) error {
	var cfg RespondConfig
	if len(config) > 0 {
		cfg = config[0]
	}
	if !isValidHTTPStatus(status) {
		rerr := RespondError(w, 500, errToRespond, config...)
		if rerr != nil {
			return rerr
		}
		return fmt.Errorf("RespondError, status is invalid")
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
