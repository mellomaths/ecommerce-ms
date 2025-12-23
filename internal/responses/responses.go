package responses

import (
	"encoding/json"
	"net/http"
)

func NewJsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func NewJsonErrorResponse(w http.ResponseWriter, status int, errCode string, errMsg string) {
	type ErrorResponse struct {
		ErrorCode    string `json:"error_code"`
		ErrorMessage string `json:"error_message"`
	}
	NewJsonResponse(w, status, ErrorResponse{
		ErrorCode:    errCode,
		ErrorMessage: errMsg,
	})
}
