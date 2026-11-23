package handler

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type middlewareEcho struct {
	userID    int
	requestID string
	called    bool
}

func newMiddlewareChain(logOutput *bytes.Buffer, echo *middlewareEcho) http.Handler {
	logger := slog.New(slog.NewJSONHandler(logOutput, nil))

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		echo.called = true
		echo.userID, _ = r.Context().Value(currentUserIDKey{}).(int)
		echo.requestID = RequestIDFromContext(r.Context())
		writeJSON(w, http.StatusOK, map[string]int{"user_id": echo.userID})
	})

	return RequestIDMiddleware(
		LoggingMiddleware(logger)(
			TimeoutMiddleware(time.Second)(
				IdentityMiddleware(inner),
			),
		),
	)
}

func TestIdentityMiddlewareAcceptsValidHeader(t *testing.T) {
	var logOutput bytes.Buffer
	echo := &middlewareEcho{}
	chain := newMiddlewareChain(&logOutput, echo)

	req := httptest.NewRequest(http.MethodGet, "/v1/me/pairs/2", nil)
	req.Header.Set(DemoUserHeader, "7")
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	if !echo.called || echo.userID != 7 {
		t.Fatalf("handler saw called=%v userID=%d", echo.called, echo.userID)
	}

	requestID := rec.Header().Get(RequestIDHeader)
	if requestID == "" || requestID != echo.requestID {
		t.Fatalf("X-Request-ID=%q, handler context ID=%q", requestID, echo.requestID)
	}
	if len(requestID) != 32 {
		t.Fatalf("X-Request-ID=%q: want 32 hex chars", requestID)
	}

	var entry map[string]any
	if err := json.Unmarshal(logOutput.Bytes(), &entry); err != nil {
		t.Fatalf("log is not structured JSON: %v\n%s", err, logOutput.String())
	}
	for _, key := range []string{"request_id", "method", "path", "status", "duration"} {
		if _, ok := entry[key]; !ok {
			t.Errorf("log misses %q: %s", key, logOutput.String())
		}
	}
	if entry["request_id"] != requestID ||
		entry["method"] != http.MethodGet ||
		entry["path"] != "/v1/me/pairs/2" ||
		entry["status"] != float64(http.StatusOK) {
		t.Errorf("unexpected log fields: %s", logOutput.String())
	}
}

func TestIdentityMiddlewareRejectsInvalidHeaders(t *testing.T) {
	for _, tt := range []struct {
		name   string
		values []string
	}{
		{"missing", nil},
		{"zero", []string{"0"}},
		{"leading zero", []string{"01"}},
		{"negative", []string{"-1"}},
		{"positive sign", []string{"+1"}},
		{"float", []string{"1.5"}},
		{"letters", []string{"abc"}},
		{"inner space", []string{"1 2"}},
		{"padding space", []string{" 1"}},
		{"overflow", []string{"999999999999999999999999999"}},
		{"repeated", []string{"1", "2"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var logOutput bytes.Buffer
			echo := &middlewareEcho{}
			chain := newMiddlewareChain(&logOutput, echo)

			req := httptest.NewRequest(http.MethodPut, "/v1/me/likes/2", nil)
			for _, value := range tt.values {
				req.Header.Add(DemoUserHeader, value)
			}
			rec := httptest.NewRecorder()
			chain.ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status=%d, want 400", rec.Code)
			}
			var apiErr ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &apiErr); err != nil {
				t.Fatalf("error body: %v", err)
			}
			if apiErr.Code != "invalid_request" || apiErr.Message == "" {
				t.Fatalf("error=%+v", apiErr)
			}
			if echo.called {
				t.Fatal("rejected request reached the business handler")
			}

			if rec.Header().Get(RequestIDHeader) == "" {
				t.Fatal("X-Request-ID is missing on the error response")
			}
			if !strings.Contains(logOutput.String(), `"status":400`) {
				t.Fatalf("log misses the 400 status: %s", logOutput.String())
			}
		})
	}
}

func TestTimeoutMiddlewareBoundsContext(t *testing.T) {
	deadlineSet := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, deadlineSet = r.Context().Deadline()
		w.WriteHeader(http.StatusOK)
	})

	chain := TimeoutMiddleware(800 * time.Millisecond)(inner)

	req := httptest.NewRequest(http.MethodGet, "/v1/me/discovery", nil)
	rec := httptest.NewRecorder()
	chain.ServeHTTP(rec, req)

	if !deadlineSet {
		t.Fatal("handler context has no deadline")
	}
}
