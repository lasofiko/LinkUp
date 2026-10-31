package profilemock

import "net/http"


func NewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "profile API mock is not implemented", http.StatusNotImplemented)
	})
}
