package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

type profileHTTPStore struct {
	profile domain.Profile
	updates int
	err     error
}

func (s *profileHTTPStore) GetProfile(_ context.Context, userID int) (domain.Profile, error) {
	if s.err != nil {
		return domain.Profile{}, s.err
	}
	if userID != s.profile.UserID {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	return s.profile, nil
}

func (s *profileHTTPStore) UpdateProfile(_ context.Context, userID int, version int64, city string, interests []domain.Interest) (domain.Profile, error) {
	s.updates++
	if s.err != nil {
		return domain.Profile{}, s.err
	}
	if userID != s.profile.UserID {
		return domain.Profile{}, domain.ErrProfileNotFound
	}
	if version != s.profile.Version {
		return domain.Profile{}, domain.ErrVersionConflict
	}
	s.profile.City = city
	s.profile.Interests = append([]domain.Interest{}, interests...)
	s.profile.Version++
	return s.profile, nil
}

func newProfileHTTPFixture() (http.Handler, *profileHTTPStore) {
	store := &profileHTTPStore{profile: domain.Profile{
		UserID: 2, City: "Moscow", Interests: []domain.Interest{"Go"}, Available: true, Version: 1,
	}}
	service := usecase.NewProfileService(store)
	mux := http.NewServeMux()
	mux.Handle("GET /v1/profiles/{userID}", GetProfileHandler(service))
	mux.Handle("PUT /v1/profiles/{userID}", UpdateProfileHandler(service))
	return IdentityMiddleware(mux), store
}

func requestProfileHTTP(h http.Handler, method, path, actor, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if actor != "" {
		req.Header.Set(DemoUserHeader, actor)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	return w
}

func TestProfileHTTPGetAndUpdate(t *testing.T) {
	h, store := newProfileHTTPFixture()

	read := requestProfileHTTP(h, http.MethodGet, "/v1/profiles/2", "1", "")
	if read.Code != http.StatusOK {
		t.Fatalf("GET status=%d: %s", read.Code, read.Body.String())
	}
	var before ProfileResponse
	if err := json.Unmarshal(read.Body.Bytes(), &before); err != nil || before.UserID != 2 || before.Version != 1 || !before.Available {
		t.Fatalf("GET response=%+v err=%v", before, err)
	}

	updated := requestProfileHTTP(h, http.MethodPut, "/v1/profiles/2", "2", `{"expected_version":1,"city":" Paris ","interests":[" Go ","Music"]}`)
	if updated.Code != http.StatusOK {
		t.Fatalf("PUT status=%d: %s", updated.Code, updated.Body.String())
	}
	var after ProfileResponse
	if err := json.Unmarshal(updated.Body.Bytes(), &after); err != nil || after.Version != 2 || after.City != "Paris" || !after.Available || len(after.Interests) != 2 || after.Interests[0] != "Go" || after.Interests[1] != "Music" {
		t.Fatalf("PUT response=%+v err=%v", after, err)
	}
	if store.updates != 1 {
		t.Fatalf("updates=%d, want 1", store.updates)
	}

	stale := requestProfileHTTP(h, http.MethodPut, "/v1/profiles/2", "2", `{"expected_version":1,"city":"Berlin","interests":[]}`)
	if stale.Code != http.StatusConflict {
		t.Fatalf("stale PUT status=%d: %s", stale.Code, stale.Body.String())
	}
	var e ErrorResponse
	if err := json.Unmarshal(stale.Body.Bytes(), &e); err != nil || e.Code != "version_conflict" {
		t.Fatalf("stale error=%+v err=%v", e, err)
	}
	if store.profile.City != "Paris" || store.profile.Version != 2 {
		t.Fatalf("first edit overwritten: %+v", store.profile)
	}

	read = requestProfileHTTP(h, http.MethodGet, "/v1/profiles/2", "2", "")
	if read.Code != http.StatusOK {
		t.Fatalf("second GET status=%d: %s", read.Code, read.Body.String())
	}
	if err := json.Unmarshal(read.Body.Bytes(), &after); err != nil || after.Version != 2 || after.City != "Paris" {
		t.Fatalf("second GET response=%+v err=%v", after, err)
	}
}

func TestProfileHTTPValidation(t *testing.T) {
	const valid = `{"expected_version":1,"city":"Paris","interests":["Go"]}`
	const withAvailable = `{"expected_version":1,"city":"Paris","interests":["Go"],"available":false}`
	large := `{"expected_version":1,"city":"` + strings.Repeat("x", int(maxProfileEditBodyBytes)) + `","interests":[]}`

	tests := []struct {
		name, method, path, actor, body string
		status                          int
		code                            string
	}{
		{"missing identity", http.MethodPut, "/v1/profiles/2", "", valid, 400, "invalid_request"},
		{"invalid profile id", http.MethodPut, "/v1/profiles/02", "2", valid, 400, "invalid_request"},
		{"query rejected", http.MethodPut, "/v1/profiles/2?debug=1", "2", valid, 400, "invalid_request"},
		{"nonowner forbidden", http.MethodPut, "/v1/profiles/2", "1", valid, 403, "forbidden"},
		{"missing profile", http.MethodPut, "/v1/profiles/9", "9", valid, 404, "not_found"},
		{"missing version", http.MethodPut, "/v1/profiles/2", "2", `{"city":"Paris","interests":[]}`, 400, "invalid_request"},
		{"null interests", http.MethodPut, "/v1/profiles/2", "2", `{"expected_version":1,"city":"Paris","interests":null}`, 400, "invalid_request"},
		{"unknown field", http.MethodPut, "/v1/profiles/2", "2", withAvailable, 400, "invalid_request"},
		{"duplicate interest", http.MethodPut, "/v1/profiles/2", "2", `{"expected_version":1,"city":"Paris","interests":["Go"," go "]}`, 400, "invalid_request"},
		{"invalid JSON", http.MethodPut, "/v1/profiles/2", "2", `{`, 400, "invalid_request"},
		{"two JSON objects", http.MethodPut, "/v1/profiles/2", "2", valid + `{}`, 400, "invalid_request"},
		{"too large", http.MethodPut, "/v1/profiles/2", "2", large, 400, "invalid_request"},
		{"GET missing profile", http.MethodGet, "/v1/profiles/9", "2", "", 404, "not_found"},
		{"GET with body", http.MethodGet, "/v1/profiles/2", "2", "{}", 400, "invalid_request"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h, store := newProfileHTTPFixture()
			w := requestProfileHTTP(h, tt.method, tt.path, tt.actor, tt.body)
			if w.Code != tt.status {
				t.Fatalf("status=%d want=%d, body=%s", w.Code, tt.status, w.Body.String())
			}
			var e ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &e); err != nil || e.Code != tt.code {
				t.Fatalf("error=%+v, want code=%q, decode error=%v", e, tt.code, err)
			}
			if store.profile.Version != 1 || store.profile.City != "Moscow" {
				t.Fatalf("failed request modified profile: %+v", store.profile)
			}
		})
	}
}

func TestProfileHTTPInternalErrorIsNotLeaked(t *testing.T) {
	h, store := newProfileHTTPFixture()
	store.err = errors.New("private database credentials must not leak")
	w := requestProfileHTTP(h, http.MethodGet, "/v1/profiles/2", "2", "")
	if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "credentials") {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
}
