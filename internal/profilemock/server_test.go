package profilemock

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProfileModes(t *testing.T) {
	handler := NewHandler()

	tests := []struct {
		mode       string
		wantStatus int
	}{
		{"ok", http.StatusOK},
		{"missing", http.StatusNotFound},
		{"error", http.StatusInternalServerError},
		{"large", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPut,
				"/__admin/mode/"+tt.mode,
				nil,
			)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != http.StatusOK {
				t.Fatalf("switch mode: status = %d", recorder.Code)
			}

			req = httptest.NewRequest(
				http.MethodGet,
				"/v1/profiles/2",
				nil,
			)
			recorder = httptest.NewRecorder()
			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf(
					"status = %d, want %d",
					recorder.Code,
					tt.wantStatus,
				)
			}

			if tt.mode == "large" && recorder.Body.Len() <= 64<<10 {
				t.Fatal("large response does not exceed limit")
			}

			if tt.mode == "ok" {
				var profile profileResponse

				if err := json.Unmarshal(
					recorder.Body.Bytes(),
					&profile,
				); err != nil {
					t.Fatal(err)
				}

				if profile.UserID != 2 ||
					!profile.Available ||
					profile.City != "Moscow" {
					t.Fatalf("unexpected profile: %+v", profile)
				}
			}
		})
	}
}

func TestProfileValidation(t *testing.T) {
	handler := NewHandler()

	tests := []struct {
		path       string
		wantStatus int
	}{
		{"/v1/profiles/0", http.StatusBadRequest},
		{"/v1/profiles/02", http.StatusBadRequest},
		{"/v1/profiles/abc", http.StatusBadRequest},
		{"/v1/profiles/999", http.StatusNotFound},
		{"/v1/profiles/3", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			if recorder.Code != tt.wantStatus {
				t.Fatalf(
					"status = %d, want %d",
					recorder.Code,
					tt.wantStatus,
				)
			}

			if tt.path == "/v1/profiles/3" {
				var profile profileResponse

				if err := json.Unmarshal(
					recorder.Body.Bytes(),
					&profile,
				); err != nil {
					t.Fatal(err)
				}

				if profile.Available {
					t.Fatal("profile 3 must be unavailable")
				}
			}
		})
	}
}

func TestInvalidMode(t *testing.T) {
	handler := NewHandler()

	req := httptest.NewRequest(
		http.MethodPut,
		"/__admin/mode/unknown",
		nil,
	)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", recorder.Code)
	}

	// Некорректный режим не должен изменить исходный ok.
	req = httptest.NewRequest(
		http.MethodGet,
		"/v1/profiles/2",
		nil,
	)
	recorder = httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("mode changed unexpectedly: %d", recorder.Code)
	}
}

func TestSlowModeCancellation(t *testing.T) {
	handler := NewHandler()

	modeReq := httptest.NewRequest(
		http.MethodPut,
		"/__admin/mode/slow",
		nil,
	)
	handler.ServeHTTP(httptest.NewRecorder(), modeReq)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req := httptest.NewRequest(
		http.MethodGet,
		"/v1/profiles/2",
		nil,
	).WithContext(ctx)

	started := make(chan struct{})
	done := make(chan struct{})

	wrapped := http.HandlerFunc(func(
		w http.ResponseWriter,
		r *http.Request,
	) {
		close(started)
		handler.ServeHTTP(w, r)
	})

	recorder := httptest.NewRecorder()

	go func() {
		defer close(done)
		wrapped.ServeHTTP(recorder, req)
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("handler did not start")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("slow handler did not stop after cancellation")
	}

	if recorder.Body.Len() != 0 {
		t.Fatal("canceled handler wrote a response")
	}
}
