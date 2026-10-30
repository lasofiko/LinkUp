package handler

import "net/http"

// TODO(Ваня): accept Discovery, validate query, filter before pagination,
// map profiles to DiscoveryResponse, and cover the API contract with tests.
func DiscoveryHandler(_ Discovery) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeNotImplemented(w)
	})
}
