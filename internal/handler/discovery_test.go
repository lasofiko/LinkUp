package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/handler"
)

type DiscoveryFixture struct {
	profiles []domain.Profile
	err      error
}

func (m *DiscoveryFixture) Discover(ctx context.Context, userID int) ([]domain.Profile, error) {
	return m.profiles, m.err
}

func TestDiscoveryHandler(t *testing.T) {
	mockProfiles := []domain.Profile{
		{UserID: 2, City: "msk", Interests: []domain.Interest{"chess", "music"}},
		{UserID: 3, City: "spb", Interests: []domain.Interest{"c++", "Politics"}},
		{UserID: 4, City: "msk", Interests: []domain.Interest{"Chess", "docker"}},
	}

	tests := []struct {
		name           string
		userID         int
		query          string
		expectedStatus int
		checkResponse  func(t *testing.T, body string)
	}{
		{
			name:           "successful discovery w/o filters",
			userID:         1,
			query:          "?limit=2&offset=0",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp handler.DiscoveryResponse
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to parse a response: %v", err)
				}
				if len(resp.Items) != 2 {
					t.Errorf("expected 2 items, got %d", len(resp.Items))
				}
				if resp.Limit != 2 || resp.Offset != 0 {
					t.Errorf("expected limit=2, offset=0, got limit=%d, offset=%d", resp.Limit, resp.Offset)
				}
			},
		},
		{
			name:           "successful discovery with interest filter",
			userID:         1,
			query:          "?interest=Chess",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				var resp handler.DiscoveryResponse
				if err := json.Unmarshal([]byte(body), &resp); err != nil {
					t.Fatalf("failed to parse a response: %v", err)
				}

				if len(resp.Items) != 2 {
					t.Errorf("expected 2 items matching 'Chess', got %d", len(resp.Items))
				}
				if resp.Items[0].UserID != 2 || resp.Items[1].UserID != 4 {
					t.Errorf("unexpected matched user IDs: %v", resp.Items)
				}
			},
		},
		{
			name:           "empty items result as empty slice not null",
			userID:         1,
			query:          "?interest=NonExistent",
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body string) {
				expected := `{"items":[],"limit":20,"offset":0}`
				if body != expected+"\n" && body != expected {
					t.Errorf("expected json %s, got %s", expected, body)
				}
			},
		},
		{
			name:           "error: unknown parameter",
			userID:         1,
			query:          "?unknown=1",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error: repeated parameter",
			userID:         1,
			query:          "?limit=10&limit=20",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error: invalid limit out of range",
			userID:         1,
			query:          "?limit=100",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "error: missing user in context",
			userID:         0,
			query:          "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disc := &DiscoveryFixture{profiles: mockProfiles, err: nil}
			handlerFunc := handler.DiscoveryHandler(disc)

			req := httptest.NewRequest(http.MethodGet, "/v1/me/discovery"+tt.query, nil)

			if tt.userID > 0 {
				ctx := context.WithValue(req.Context(), handler.CtxKeyUserID, tt.userID)
				req = req.WithContext(ctx)
			}

			rec := httptest.NewRecorder()
			handlerFunc.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d (body: %s)", tt.expectedStatus, rec.Code, rec.Body.String())
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, rec.Body.String())
			}
		})
	}
}
