package handler

import "net/http"

func DiscoveryHandler(_ Discovery) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeNotImplemented(w)
	})
}
