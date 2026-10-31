package handler

import (
	"encoding/json"
	"net/http"
)

func LikeHandler(_ Reactions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeNotImplemented(w)
	})
}

func PairHandler(_ Pairs) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeNotImplemented(w)
	})
}

func writeNotImplemented(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Code: "not_implemented", Message: "sprint 3 handler is not implemented"})
}
