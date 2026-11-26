package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	StatusCode   int     `json:"statusCode"`
	IsSuccess    bool    `json:"isSuccess"`
	Result       any     `json:"result"`
	ErrorCode    *string `json:"errorCode"`
	ErrorMessage *string `json:"errorMessage"`
}

func JSON(w http.ResponseWriter, statusCode int, result any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(APIResponse{
		StatusCode:   statusCode,
		IsSuccess:    true,
		Result:       result,
		ErrorCode:    nil,
		ErrorMessage: nil,
	})
}

func Error(w http.ResponseWriter, statusCode int, errCode, errMsg string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(APIResponse{
		StatusCode:   statusCode,
		IsSuccess:    false,
		Result:       nil,
		ErrorCode:    &errCode,
		ErrorMessage: &errMsg,
	})
}
