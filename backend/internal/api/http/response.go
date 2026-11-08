package httpapi

import (
	"encoding/json"
	"io"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

func writeNoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func requireMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method == method {
		return true
	}

	writeMethodNotAllowed(w, method)
	return false
}

func writeMethodNotAllowed(w http.ResponseWriter, allow string) {
	w.Header().Set("Allow", allow)
	writeError(w, newAPIError(http.StatusMethodNotAllowed, codeMethodNotAllowed, "method not allowed", "", nil))
}

func decodeJSON(r *http.Request, dst any) *apiError {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return newAPIError(http.StatusBadRequest, codeInvalidRequest, "invalid JSON request body", "", err)
	}

	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return newAPIError(http.StatusBadRequest, codeInvalidRequest, "request body must contain a single JSON document", "", nil)
	}

	return nil
}
