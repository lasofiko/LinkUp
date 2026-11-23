package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func TestBlockHTTP(t *testing.T) {
	f := newHTTPReactFixture()
	blocks := usecase.NewBlockService(f, &httpBlockStore{f: f})
	mux := http.NewServeMux()
	mux.Handle("PUT /v1/me/blocks/{targetID}", BlockHandler(blocks))
	mux.Handle("PUT /v1/me/likes/{targetID}", LikeHandler(usecase.NewReactService(f, f, f)))
	mux.Handle("GET /v1/me/pairs/{targetID}", PairHandler(f))
	h := IdentityMiddleware(mux)
	for _, tt := range []struct {
		path, header, body string
		status             int
	}{
		{"/v1/me/blocks/2", "1", "", 200},
		{"/v1/me/blocks/2", "1", "", 200},
		{"/v1/me/blocks/0", "1", "", 400},
		{"/v1/me/blocks/2", "", "", 400},
		{"/v1/me/blocks/1", "1", "", 409},
		{"/v1/me/blocks/2", "1", "{}", 400},
		{"/v1/me/blocks/3", "1", "", 404},
		{"/v1/me/blocks/2", "3", "", 404},
	} {
		r := httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(tt.body))
		if tt.header != "" {
			r.Header.Set(DemoUserHeader, tt.header)
		}
		w := httptest.NewRecorder()
		h.ServeHTTP(w, r)
		if w.Code != tt.status {
			t.Fatalf("%s: got %d, want %d: %s", tt.path, w.Code, tt.status, w.Body.String())
		}
		if w.Code == 200 {
			var got BlockResponse
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got != (BlockResponse{TargetID: 2, Blocked: true}) {
				t.Fatalf("response: %+v, %v", got, err)
			}
		}
	}
	if f.saves != 1 {
		t.Fatalf("repeated block saved %d times", f.saves)
	}
	like := httptest.NewRequest(http.MethodPut, "/v1/me/likes/2", nil)
	like.Header.Set(DemoUserHeader, "1")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, like)
	if w.Code != http.StatusForbidden || f.saves != 1 {
		t.Fatalf("like after block: status=%d saves=%d", w.Code, f.saves)
	}
	get := httptest.NewRequest(http.MethodGet, "/v1/me/pairs/2", nil)
	get.Header.Set(DemoUserHeader, "1")
	w = httptest.NewRecorder()
	h.ServeHTTP(w, get)
	var pair PairResponse
	if err := json.Unmarshal(w.Body.Bytes(), &pair); err != nil || w.Code != http.StatusOK || pair.Mutual {
		t.Fatalf("pair after block: status=%d pair=%+v err=%v", w.Code, pair, err)
	}
}

// This test adapter applies the same state transition as the shared memory
// store, while retaining the existing HTTP fixture's save counter.
type httpBlockStore struct{ f *httpReactFixture }

func (s *httpBlockStore) Block(_ context.Context, pair interaction.Pair) (interaction.State, error) {
	state, ok := s.f.states[pair]
	if !ok {
		state.Pair = pair
	}
	if !state.Blocked {
		state.Blocked = true
		if err := s.f.Save(state); err != nil {
			return interaction.State{}, err
		}
	}
	return state, nil
}
