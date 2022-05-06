package httputils

import (
	"encoding/json"
	"net/http"
)

type JSONNode map[string]interface{}

// HTTPRespondJSON will send JSON data to the client.
func HTTPRespondJSON(w http.ResponseWriter, code int, data JSONNode) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(data)
}

// HTTPRespondSuccess will send success JSON data to the client.
func HTTPRespondSuccess(w http.ResponseWriter, version string, code int, data JSONNode) {
	d := JSONNode{
		"apiVersion": version,
		"data":       data,
	}

	HTTPRespondJSON(w, code, d)
}

// HTTPRespondSuccess will send fail JSON message to the client.
func HTTPRespondFailed(w http.ResponseWriter, version string, code int, errMsg string, err interface{}) {
	d := JSONNode{
		"apiVersion": version,
		"error": JSONNode{
			"code":    code,
			"message": errMsg,
			"errors":  err,
		},
	}

	HTTPRespondJSON(w, code, d)
}
