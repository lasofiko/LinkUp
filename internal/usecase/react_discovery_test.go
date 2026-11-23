package usecase

import (
	"context"
	"reflect"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/matching"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
)

func newReactDiscoveryFixture(t *testing.T) (*ReactService, *PairDiscoveryService, []domain.Profile) {
	t.Helper()
	users := memory.NewUserRepository()
	for _, user := range []domain.User{{ID: 1, Name: "Соня"}, {ID: 2, Name: "Арина"}} {
		if err := users.Create(user); err != nil {
			t.Fatal(err)
		}
	}
	profiles := []domain.Profile{
		{UserID: 1, City: "Moscow", Available: true, Interests: []domain.Interest{"Go", "Music"}},
		{UserID: 2, City: "Moscow", Available: true, Interests: []domain.Interest{"Go", "Books"}},
	}
	repo := memory.NewProfiles(profiles)
	store := pairStoreFixture{}
	return NewReactService(users, repo, store),
		NewPairDiscoveryService(matching.NewDiscoveryService(repo), store), profiles
}

func TestCandidatesThroughReact(t *testing.T) {
	for _, direction := range []struct {
		name             string
		sender, receiver int
	}{
		{"A likes B first", 1, 2},
		{"B likes A first", 2, 1},
	} {
		t.Run(direction.name, func(t *testing.T) {
			for _, scenario := range []struct {
				name     string
				commands [][2]int
				visible  [2]bool // Sender's and receiver's discovery results.
			}{
				{"before likes", nil, [2]bool{true, true}},
				{"after first like", [][2]int{{direction.sender, direction.receiver}}, [2]bool{false, true}},
				{"repeat first like", [][2]int{{direction.sender, direction.receiver}, {direction.sender, direction.receiver}}, [2]bool{false, true}},
				{"after mutual like", [][2]int{{direction.sender, direction.receiver}, {direction.receiver, direction.sender}}, [2]bool{false, false}},
				{"repeat both likes", [][2]int{{direction.sender, direction.receiver}, {direction.receiver, direction.sender}, {direction.sender, direction.receiver}, {direction.receiver, direction.sender}}, [2]bool{false, false}},
			} {
				t.Run(scenario.name, func(t *testing.T) {
					react, discovery, profiles := newReactDiscoveryFixture(t)
					ctx := context.Background()
					for _, command := range scenario.commands {
						if _, err := react.React(ctx, command[0], command[1]); err != nil {
							t.Fatalf("React(%d, %d): %v", command[0], command[1], err)
						}
					}
					for side, userID := range []int{direction.sender, direction.receiver} {
						want := []domain.Profile{}
						if scenario.visible[side] {
							want = append(want, profiles[2-userID])
						}
						// Repeated reads must remain stable, including non-nil empty results.
						for request := 0; request < 2; request++ {
							got, err := discovery.Discover(ctx, userID)
							if err != nil {
								t.Fatal(err)
							}
							if got == nil {
								t.Fatalf("user %d: expected non-nil candidates", userID)
							}
							if !reflect.DeepEqual(got, want) {
								t.Fatalf("user %d: got %#v, want %#v", userID, got, want)
							}
						}
					}
				})
			}
		})
	}
}

func TestCandidatesThroughReactResultIsolation(t *testing.T) {
	react, discovery, profiles := newReactDiscoveryFixture(t)
	ctx := context.Background()
	for _, userID := range []int{1, 2} {
		got, err := discovery.Discover(ctx, userID)
		if err != nil || len(got) != 1 {
			t.Fatalf("user %d: got %#v, %v", userID, got, err)
		}
		got[0].UserID = 99
		got[0].City = "Changed"
		got[0].Available = false
		got[0].Interests[0] = "Changed"
		got[0].Interests = append(got[0].Interests, "Extra")
		again, err := discovery.Discover(ctx, userID)
		want := []domain.Profile{profiles[2-userID]}
		if err != nil || !reflect.DeepEqual(again, want) {
			t.Fatalf("user %d: mutation leaked into discovery: %#v, %v", userID, again, err)
		}
	}
	// Mutating returned profiles must not make the real candidate unavailable to React.
	if _, err := react.React(ctx, 1, 2); err != nil {
		t.Fatalf("reaction after mutating discovery results: %v", err)
	}
	got, err := discovery.Discover(ctx, 2)
	if err != nil || !reflect.DeepEqual(got, []domain.Profile{profiles[0]}) {
		t.Fatalf("receiver's discovery after reaction: %#v, %v", got, err)
	}
}
