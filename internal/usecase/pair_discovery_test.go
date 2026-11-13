package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/matching"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
)

type pairStoreFixture map[interaction.Pair]interaction.State

var _ interaction.Store = pairStoreFixture{}

func (s pairStoreFixture) Load(pair interaction.Pair) (interaction.State, bool, error) {
	state, found := s[pair]
	return state, found, nil
}

func (s pairStoreFixture) Save(state interaction.State) error {
	s[state.Pair] = state
	return nil
}

func TestPairDiscoveryReadsSavedChanges(t *testing.T) {
	profiles := memory.NewProfiles([]domain.Profile{
		{UserID: 1, City: "Moscow", Available: true, Interests: []domain.Interest{"Go"}},
		{UserID: 2, City: "Moscow", Available: true, Interests: []domain.Interest{"Go"}},
		{UserID: 3, City: "Moscow", Available: false, Interests: []domain.Interest{"Go"}},
		{UserID: 4, City: "Other", Available: true, Interests: []domain.Interest{"Go"}},
	})
	store := pairStoreFixture{}
	service := NewPairDiscoveryService(matching.NewDiscoveryService(profiles), store)
	check := func(userID, want int) {
		t.Helper()
		got, err := service.Discover(context.Background(), userID)
		if err != nil || got == nil || len(got) != want {
			t.Fatalf("user %d: got %#v, %v; want %d candidates", userID, got, err, want)
		}
		if want == 1 && got[0].UserID != 3-userID {
			t.Fatalf("unexpected candidate: %#v", got)
		}
	}
	check(1, 1)
	check(2, 1)
	pair, err := interaction.NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, command := range [][2]int{{1, 2}, {2, 1}, {1, 2}, {2, 1}} {
		state, _, err := store.Load(pair)
		if err != nil {
			t.Fatal(err)
		}
		next, changed, err := interaction.ApplyLike(state, command[0], command[1])
		if err != nil {
			t.Fatal(err)
		}
		if changed {
			if err := store.Save(next); err != nil {
				t.Fatal(err)
			}
		}
		check(1, 0)
		if next.IsMutual() {
			check(2, 0)
		} else {
			check(2, 1)
		}
	}
}

type failingPairCandidates struct{ err error }

func (s failingPairCandidates) Recommend(context.Context, int) ([]domain.Profile, error) {
	return nil, s.err
}

func TestPairDiscoverySearchError(t *testing.T) {
	root := errors.New("search failed")
	service := NewPairDiscoveryService(failingPairCandidates{root}, nil)
	got, err := service.Discover(context.Background(), 1)
	if got != nil || !errors.Is(err, root) {
		t.Fatalf("got %#v, %v; want original search error", got, err)
	}
}

type pairReaderStub struct {
	state interaction.State
	found bool
	err   error
}

func (r *pairReaderStub) Load(pair interaction.Pair) (interaction.State, bool, error) {
	return r.state, r.found, r.err
}

func TestPairCandidatesStates(t *testing.T) {
	pair, err := interaction.NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name                      string
		found, outgoing, incoming bool
		wantA, wantB              int
	}{
		{"before likes", false, false, false, 1, 1},
		{"one like", true, true, false, 0, 1},
		{"match", true, true, true, 0, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			reader := &pairReaderStub{state: interaction.State{Pair: pair, FirstLikesSecond: tt.outgoing, SecondLikesFirst: tt.incoming}, found: tt.found}
			profiles := []domain.Profile{{UserID: 1, Available: true, Interests: []domain.Interest{"Go"}}, {UserID: 2, Available: true, Interests: []domain.Interest{"Go"}}, {UserID: 3}}
			for _, side := range []struct{ id, want int }{{1, tt.wantA}, {2, tt.wantB}} {
				got, err := FilterPairCandidates(side.id, profiles, reader)
				if err != nil {
					t.Fatal(err)
				}
				if got == nil || len(got) != side.want {
					t.Fatalf("user %d: got %#v, want length %d", side.id, got, side.want)
				}
				if len(got) > 0 {
					got[0].Interests[0] = "Changed"
					again, err := FilterPairCandidates(side.id, profiles, reader)
					if err != nil {
						t.Fatal(err)
					}
					if again[0].Interests[0] != "Go" {
						t.Fatal("result changed source")
					}
				}
			}
		})
	}
}

func TestPairCandidatesReadErrors(t *testing.T) {
	root := errors.New("read failed")
	for _, reader := range []*pairReaderStub{
		{err: root},
		{found: true, state: interaction.State{Pair: interaction.Pair{FirstUserID: 1, SecondUserID: 3}}},
	} {
		got, err := FilterPairCandidates(1, []domain.Profile{{UserID: 2, Available: true}}, reader)
		if got != nil || err == nil {
			t.Fatalf("expected error and nil result, got %#v, %v", got, err)
		}
		if reader.err != nil && !errors.Is(err, root) {
			t.Fatal("lost source error")
		}
		if reader.err == nil && !errors.Is(err, interaction.ErrStatePair) {
			t.Fatal("lost pair error")
		}
	}
}

func TestPairCandidatesRepeatedTransitions(t *testing.T) {
	reader := &pairReaderStub{}
	profiles := []domain.Profile{{UserID: 1, Available: true}, {UserID: 2, Available: true}}
	for _, command := range [][2]int{{1, 2}, {2, 1}, {1, 2}, {2, 1}} {
		next, changed, err := interaction.ApplyLike(reader.state, command[0], command[1])
		if err != nil {
			t.Fatal(err)
		}
		if reader.state.IsMutual() && changed {
			t.Fatal("repeated like changed mutual state")
		}
		reader.state, reader.found = next, true
		for _, id := range []int{1, 2} {
			got, err := FilterPairCandidates(id, profiles, reader)
			if err != nil {
				t.Fatal(err)
			}
			want := 0
			if id == 2 && !next.IsMutual() {
				want = 1
			}
			if got == nil || len(got) != want {
				t.Fatalf("user %d: got %#v", id, got)
			}
		}
	}
}

type isolatedPairReader struct{ state interaction.State }

func (r isolatedPairReader) Load(pair interaction.Pair) (interaction.State, bool, error) {
	if pair == r.state.Pair {
		return r.state, true, nil
	}
	return interaction.State{}, false, nil
}

func TestPairCandidatesUnrelatedPairs(t *testing.T) {
	state, _, err := interaction.ApplyLike(interaction.State{}, 1, 2)
	if err != nil {
		t.Fatal(err)
	}
	got, err := FilterPairCandidates(3, []domain.Profile{{UserID: 2, Available: true}, {UserID: 1, Available: true}}, isolatedPairReader{state: state})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].UserID != 2 || got[1].UserID != 1 {
		t.Fatalf("unrelated pair affected order or candidates: %#v", got)
	}
}
