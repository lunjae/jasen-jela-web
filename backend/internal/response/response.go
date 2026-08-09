package response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

type ErrorBody struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}
type Envelope struct {
	Data  any        `json:"data"`
	Error *ErrorBody `json:"error"`
}
type APIError struct {
	Status        int
	Code, Message string
	Fields        map[string]string
}

func (e *APIError) Error() string { return e.Message }
func Write(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: data})
}
func WriteError(w http.ResponseWriter, err error) {
	var ae *APIError
	if !errors.As(err, &ae) {
		slog.Error("request failed", "error", err)
		ae = &APIError{Status: 500, Code: "internal_error", Message: "Došlo je do greške. Pokušajte ponovo."}
	}
	WriteJSONError(w, ae)
}
func WriteJSONError(w http.ResponseWriter, e *APIError) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(e.Status)
	_ = json.NewEncoder(w).Encode(Envelope{Data: nil, Error: &ErrorBody{Code: e.Code, Message: e.Message, Fields: e.Fields}})
}
func BadRequest(fields map[string]string) *APIError {
	return &APIError{Status: 400, Code: "validation_error", Message: "Uneti podaci nisu ispravni.", Fields: fields}
}
func NotFound(msg string) *APIError { return &APIError{Status: 404, Code: "not_found", Message: msg} }
