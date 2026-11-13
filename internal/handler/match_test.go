package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/profileclient"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

type httpReactFixture struct {
	states     map[interaction.Pair]interaction.State
	saves      int
	profiles   map[int]domain.Profile
	loadErr    error
	profileErr error
}

func newHTTPReactFixture() *httpReactFixture {
	return &httpReactFixture{
		states: make(map[interaction.Pair]interaction.State),
		profiles: map[int]domain.Profile{
			1: {UserID: 1, Available: true},
			2: {UserID: 2, Available: true},
		},
	}
}

func (*httpReactFixture) GetByID(id int) (domain.User, error) {
	if id != 1 && id != 2 {
		return domain.User{}, fmt.Errorf("user %d: %w", id, domain.ErrUserNotFound)
	}
	return domain.User{ID: id}, nil
}

func (f *httpReactFixture) GetProfile(_ context.Context, id int) (domain.Profile, error) {
	if f.profileErr != nil {
		return domain.Profile{}, fmt.Errorf("profile %d: %w", id, f.profileErr)
	}
	p, ok := f.profiles[id]
	if !ok {
		return domain.Profile{}, fmt.Errorf("profile %d: %w", id, domain.ErrProfileNotFound)
	}
	return p, nil
}

func (f *httpReactFixture) Load(pair interaction.Pair) (interaction.State, bool, error) {
	if f.loadErr != nil {
		return interaction.State{}, false, fmt.Errorf("load: %w", f.loadErr)
	}
	state, ok := f.states[pair]
	return state, ok, nil
}

func (f *httpReactFixture) Save(state interaction.State) error {
	f.saves++
	f.states[state.Pair] = state
	return nil
}

func testPairMux(f *httpReactFixture) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("PUT /v1/me/likes/{targetID}", LikeHandler(usecase.NewReactService(f, f, f)))
	mux.Handle("GET /v1/me/pairs/{targetID}", PairHandler(f))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The runtime middleware owns header validation. Simulate only its context handoff.
		switch r.Header.Get("X-Demo-User") {
		case "1":
			r = r.WithContext(WithCurrentUserID(r.Context(), 1))
		case "2":
			r = r.WithContext(WithCurrentUserID(r.Context(), 2))
		}
		mux.ServeHTTP(w, r)
	})
}

func pairRequest(t *testing.T, h http.Handler, method, path, userID, body string) (int, PairResponse, ErrorResponse) {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if userID != "" {
		req.Header.Set("X-Demo-User", userID)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if got := w.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type=%q", got)
	}
	var pair PairResponse
	var apiErr ErrorResponse
	if w.Code == http.StatusOK {
		if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil {
			t.Fatal(err)
		}
	} else if err := json.Unmarshal(w.Body.Bytes(), &apiErr); err != nil {
		t.Fatal(err)
	}
	return w.Code, pair, apiErr
}

func TestPairHTTPFlow(t *testing.T) {
	f := newHTTPReactFixture()
	h := testPairMux(f)
	check := func(method, path, user string, want PairResponse, saves int) {
		t.Helper()
		status, got, _ := pairRequest(t, h, method, path, user, "")
		if status != http.StatusOK || got != want || f.saves != saves {
			t.Fatalf("%s %s user=%s: status=%d pair=%+v saves=%d; want %+v, %d", method, path, user, status, got, f.saves, want, saves)
		}
	}
	forward := PairResponse{TargetID: 2, Liked: true}
	reverse := PairResponse{TargetID: 1, LikedBy: true}
	check("GET", "/v1/me/pairs/2", "1", PairResponse{TargetID: 2}, 0)
	check("PUT", "/v1/me/likes/2", "1", forward, 1)
	check("GET", "/v1/me/pairs/2", "1", forward, 1)
	check("GET", "/v1/me/pairs/1", "2", reverse, 1)
	check("PUT", "/v1/me/likes/2", "1", forward, 1)
	check("PUT", "/v1/me/likes/1", "2", PairResponse{TargetID: 1, Liked: true, LikedBy: true, Mutual: true}, 2)
	check("GET", "/v1/me/pairs/2", "1", PairResponse{TargetID: 2, Liked: true, LikedBy: true, Mutual: true}, 2)
	check("PUT", "/v1/me/likes/1", "2", PairResponse{TargetID: 1, Liked: true, LikedBy: true, Mutual: true}, 2)
	if len(f.states) != 1 {
		t.Fatalf("saved %d pairs; want 1", len(f.states))
	}
}

func TestPairHTTPValidation(t *testing.T) {
	f := newHTTPReactFixture()
	h := testPairMux(f)
	for _, tt := range []struct {
		name, method, path, user, body string
		status                         int
		code                           string
	}{
		{"zero ID", "PUT", "/v1/me/likes/0", "1", "", 400, "invalid_request"},
		{"leading zero", "GET", "/v1/me/pairs/02", "1", "", 400, "invalid_request"},
		{"signed ID", "PUT", "/v1/me/likes/+2", "1", "", 400, "invalid_request"},
		{"overflow", "GET", "/v1/me/pairs/999999999999999999999999999", "1", "", 400, "invalid_request"},
		{"missing user", "PUT", "/v1/me/likes/2", "", "", 400, "invalid_request"},
		{"extra body", "PUT", "/v1/me/likes/2", "1", "{}", 400, "invalid_request"},
		{"extra query", "GET", "/v1/me/pairs/2?user_id=2", "1", "", 400, "invalid_request"},
		{"self like", "PUT", "/v1/me/likes/1", "1", "", 409, "invalid_transition"},
		{"self pair", "GET", "/v1/me/pairs/1", "1", "", 409, "invalid_transition"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			status, _, apiErr := pairRequest(t, h, tt.method, tt.path, tt.user, tt.body)
			if status != tt.status || apiErr.Code != tt.code || apiErr.Message == "" {
				t.Fatalf("status=%d error=%+v; want %d %s", status, apiErr, tt.status, tt.code)
			}
		})
	}
	if f.saves != 0 || len(f.states) != 0 {
		t.Fatalf("invalid requests changed store: saves=%d states=%v", f.saves, f.states)
	}
}

func TestPairHTTPWrappedErrors(t *testing.T) {
	for _, tt := range []struct {
		name   string
		cause  error
		status int
		code   string
	}{
		{"unavailable", usecase.ErrReactionUnavailable, 403, "reaction_unavailable"},
		{"missing profile", domain.ErrProfileNotFound, 404, "not_found"},
		{"dependency", profileclient.ErrDependency, 502, "dependency_failure"},
		{"dependency timeout", profileclient.ErrTimeout, 504, "dependency_timeout"},
		{"typed timeout", fmt.Errorf("%w: %w", profileclient.ErrDependency, &net.DNSError{IsTimeout: true}), 504, "dependency_timeout"},
		{"internal", errors.New("secret storage failure"), 500, "internal_error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newHTTPReactFixture()
			f.profileErr = fmt.Errorf("wrapped: %w", tt.cause)
			status, _, apiErr := pairRequest(t, testPairMux(f), "PUT", "/v1/me/likes/2", "1", "")
			if status != tt.status || apiErr.Code != tt.code || strings.Contains(apiErr.Message, "secret") || f.saves != 0 {
				t.Fatalf("status=%d error=%+v saves=%d", status, apiErr, f.saves)
			}
		})
	}
	f := newHTTPReactFixture()
	f.loadErr = fmt.Errorf("wrapped: %w", errors.New("secret store failure"))
	status, _, apiErr := pairRequest(t, testPairMux(f), "GET", "/v1/me/pairs/2", "1", "")
	if status != 500 || apiErr.Code != "internal_error" || strings.Contains(apiErr.Message, "secret") {
		t.Fatalf("status=%d error=%+v", status, apiErr)
	}
}
