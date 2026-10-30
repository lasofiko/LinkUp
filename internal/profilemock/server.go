package profilemock

import "net/http"

// TODO(Оля): implement GET /v1/profiles/{userID} with ok, missing, error,
// slow and large modes. The body and status must follow the API contract.
func NewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "profile API mock is not implemented", http.StatusNotImplemented)
	})
}
