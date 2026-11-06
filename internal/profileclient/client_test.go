package profileclient_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/profileclient"
	"github.com/lasofiko/LinkUp/internal/profilemock"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func newTestClient(
	t *testing.T,
	baseURL string,
	httpClient *http.Client,
) *profileclient.Client {
	t.Helper()

	client, err := profileclient.NewClient(
		baseURL,
		httpClient,
		64<<10,
	)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

func TestGetProfileSuccess(t *testing.T) {
	server := httptest.NewServer(profilemock.NewHandler())
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client())

	profile, err := client.GetProfile(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}

	if profile.UserID != 2 ||
		profile.City != "Moscow" ||
		!profile.Available ||
		len(profile.Interests) != 1 ||
		profile.Interests[0] != "Go" {
		t.Fatalf("unexpected profile: %+v", profile)
	}

	unavailable, err := client.GetProfile(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}

	if unavailable.Available {
		t.Fatal("profile 3 should be unavailable")
	}
}

func TestGetProfileInvalidID(t *testing.T) {
	server := httptest.NewServer(profilemock.NewHandler())
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client())

	_, err := client.GetProfile(context.Background(), 0)

	if !errors.Is(err, profileclient.ErrInvalidUserID) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetProfileInvalidResponses(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantError error
	}{
		{
			name:      "missing",
			status:    http.StatusNotFound,
			body:      `{"error":"not found"}`,
			wantError: domain.ErrProfileNotFound,
		},
		{
			name:      "server error",
			status:    http.StatusInternalServerError,
			body:      `{"error":"internal"}`,
			wantError: profileclient.ErrDependency,
		},
		{
			name:      "invalid JSON",
			status:    http.StatusOK,
			body:      `{invalid}`,
			wantError: profileclient.ErrDependency,
		},
		{
			name:      "wrong user ID",
			status:    http.StatusOK,
			body:      `{"user_id":3,"city":"Moscow","interests":["Go"],"available":true}`,
			wantError: profileclient.ErrDependency,
		},
		{
			name:      "missing required field",
			status:    http.StatusOK,
			body:      `{"user_id":2,"city":"Moscow","interests":["Go"]}`,
			wantError: profileclient.ErrDependency,
		},
		{
			name:      "unknown field",
			status:    http.StatusOK,
			body:      `{"user_id":2,"city":"Moscow","interests":["Go"],"available":true,"extra":1}`,
			wantError: profileclient.ErrDependency,
		},
		{
			name:      "extra JSON value",
			status:    http.StatusOK,
			body:      `{"user_id":2,"city":"Moscow","interests":["Go"],"available":true} {}`,
			wantError: profileclient.ErrDependency,
		},
		{
			name:      "null interests",
			status:    http.StatusOK,
			body:      `{"user_id":2,"city":"Moscow","interests":null,"available":true}`,
			wantError: profileclient.ErrDependency,
		},
		{
			name:      "empty city",
			status:    http.StatusOK,
			body:      `{"user_id":2,"city":" ","interests":["Go"],"available":true}`,
			wantError: profileclient.ErrDependency,
		},
		{
			name:      "empty interest",
			status:    http.StatusOK,
			body:      `{"user_id":2,"city":"Moscow","interests":[" "],"available":true}`,
			wantError: profileclient.ErrDependency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.status)
					_, _ = io.WriteString(w, tt.body)
				},
			))
			defer server.Close()

			client := newTestClient(t, server.URL, server.Client())

			_, err := client.GetProfile(context.Background(), 2)

			if !errors.Is(err, tt.wantError) {
				t.Fatalf(
					"error = %v, want %v",
					err,
					tt.wantError,
				)
			}
		})
	}
}

func TestGetProfileTooLarge(t *testing.T) {
	body := `{"user_id":2,"city":"` +
		strings.Repeat("M", 70<<10) +
		`","interests":["Go"],"available":true}`

	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, body)
		},
	))
	defer server.Close()

	client := newTestClient(t, server.URL, server.Client())

	_, err := client.GetProfile(context.Background(), 2)

	if !errors.Is(err, profileclient.ErrDependency) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetProfileNetworkError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {},
	))

	address := server.URL
	server.Close()

	client := newTestClient(t, address, &http.Client{
		Timeout: time.Second,
	})

	_, err := client.GetProfile(context.Background(), 2)

	if !errors.Is(err, profileclient.ErrDependency) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetProfileCancellation(t *testing.T) {
	tests := []struct {
		name         string
		manualCancel bool
		wantError    error
	}{
		{
			name:         "explicit cancellation",
			manualCancel: true,
			wantError:    context.Canceled,
		},
		{
			name:         "client timeout",
			manualCancel: false,
			wantError:    profileclient.ErrTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			started := make(chan struct{})
			canceled := make(chan struct{})

			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					close(started)

					<-r.Context().Done()

					close(canceled)
				},
			))
			defer server.Close()

			timeout := 3 * time.Second
			if !tt.manualCancel {
				timeout = 100 * time.Millisecond
			}

			client := newTestClient(t, server.URL, &http.Client{
				Timeout: timeout,
			})

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			result := make(chan error, 1)

			go func() {
				_, err := client.GetProfile(ctx, 2)
				result <- err
			}()

			select {
			case <-started:
			case <-time.After(3 * time.Second):
				t.Fatal("server did not receive request")
			}

			if tt.manualCancel {
				cancel()
			}

			select {
			case err := <-result:
				if !errors.Is(err, tt.wantError) {
					t.Fatalf(
						"error = %v, want %v",
						err,
						tt.wantError,
					)
				}

			case <-time.After(3 * time.Second):
				t.Fatal("client did not stop")
			}

			select {
			case <-canceled:
			case <-time.After(3 * time.Second):
				t.Fatal("server did not observe cancellation")
			}
		})
	}
}

type testUsers struct{}

func (testUsers) GetByID(id int) (domain.User, error) {
	if id != 1 && id != 2 {
		return domain.User{}, domain.ErrUserNotFound
	}

	return domain.User{ID: id}, nil
}

type testPairStore struct {
	state interaction.State
	found bool
	saves int
}

func (s *testPairStore) Load(
	pair interaction.Pair,
) (interaction.State, bool, error) {
	if !s.found || s.state.Pair != pair {
		return interaction.State{}, false, nil
	}

	return s.state, true, nil
}

func (s *testPairStore) Save(state interaction.State) error {
	s.state = state
	s.found = true
	s.saves++

	return nil
}

func setMockMode(
	t *testing.T,
	server *httptest.Server,
	mode string,
) {
	t.Helper()

	req, err := http.NewRequest(
		http.MethodPut,
		server.URL+"/__admin/mode/"+mode,
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"switch mode %s: status = %d",
			mode,
			resp.StatusCode,
		)
	}
}

func TestReactServiceDependencyFailures(t *testing.T) {
	server := httptest.NewServer(profilemock.NewHandler())
	defer server.Close()

	httpClient := &http.Client{
		Timeout: 100 * time.Millisecond,
	}

	client := newTestClient(t, server.URL, httpClient)
	store := &testPairStore{}

	reactions := usecase.NewReactService(
		testUsers{},
		client,
		store,
	)

	tests := []struct {
		mode      string
		wantError error
	}{
		{"missing", domain.ErrProfileNotFound},
		{"error", profileclient.ErrDependency},
		{"slow", profileclient.ErrTimeout},
		{"large", profileclient.ErrDependency},
	}

	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			setMockMode(t, server, tt.mode)

			_, err := reactions.React(
				context.Background(),
				1,
				2,
			)

			if !errors.Is(err, tt.wantError) {
				t.Fatalf(
					"error = %v, want %v",
					err,
					tt.wantError,
				)
			}

			if store.saves != 0 || store.found {
				t.Fatal("failed request changed pair state")
			}
		})
	}

	// Восстановление зависимости.
	setMockMode(t, server, "ok")

	state, err := reactions.React(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	if !state.HasLike(1, 2) || store.saves != 1 {
		t.Fatal("successful like was not saved")
	}

	// Повторный лайк не должен создавать новую запись.
	_, err = reactions.React(context.Background(), 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	if store.saves != 1 {
		t.Fatal("repeated like created an extra save")
	}
}
