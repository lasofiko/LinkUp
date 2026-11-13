package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/usecase"
)

func TestPairStoreLoadSave(t *testing.T) {
	store := NewPairStore()

	pair, err := interaction.NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}

	state, found, err := store.Load(pair)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if found || state != (interaction.State{}) {
		t.Fatalf("empty store: state=%#v found=%v", state, found)
	}

	if err := store.Save(interaction.State{}); err == nil {
		t.Fatal("Save without a pair: want error")
	}

	want := interaction.State{Pair: pair, FirstLikesSecond: true}
	if err := store.Save(want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, found, err := store.Load(pair)
	if err != nil || !found || got != want {
		t.Fatalf("Load after Save: %#v, %v, %v", got, found, err)
	}

	got.FirstLikesSecond = false
	again, _, err := store.Load(pair)
	if err != nil || !again.FirstLikesSecond {
		t.Fatalf("mutating a loaded state leaked into the store: %#v, %v", again, err)
	}
}

func TestPairStoreConcurrentOppositeAndRepeatedLikes(t *testing.T) {
	store := NewPairStore()
	pair, err := interaction.NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}

	const goroutines = 16
	const repeats = 100

	var wg sync.WaitGroup
	var mu sync.Mutex
	changedTotal := 0

	for g := 0; g < goroutines; g++ {
		from, to := 1, 2
		if g%2 == 1 {
			from, to = 2, 1
		}

		wg.Add(1)
		go func(from, to int) {
			defer wg.Done()
			for i := 0; i < repeats; i++ {
				_, err := store.Update(context.Background(), pair, func(state interaction.State) (interaction.State, bool, error) {
					next, changed, err := interaction.ApplyLike(state, from, to)
					if err != nil {
						return interaction.State{}, false, err
					}
					if changed {
						mu.Lock()
						changedTotal++
						mu.Unlock()
					}
					return next, changed, nil
				})
				if err != nil {
					t.Errorf("Update(%d -> %d): %v", from, to, err)
					return
				}
			}
		}(from, to)
	}
	wg.Wait()

	if changedTotal != 2 {
		t.Fatalf("likes lost or duplicated: %d state changes, want exactly 2 (one per direction)", changedTotal)
	}

	state, found, err := store.Load(pair)
	if err != nil || !found {
		t.Fatalf("final Load: %#v, %v, %v", state, found, err)
	}
	if !state.HasLike(1, 2) || !state.HasLike(2, 1) || !state.IsMutual() {
		t.Fatalf("final state lost a reaction or mutuality: %#v", state)
	}
}

func TestReactServiceConcurrentLikesThroughPairStore(t *testing.T) {
	users := NewUserRepository()
	for _, user := range []domain.User{{ID: 1, Name: "Соня"}, {ID: 2, Name: "Арина"}} {
		if err := users.Create(user); err != nil {
			t.Fatal(err)
		}
	}
	profiles := NewProfiles([]domain.Profile{
		{UserID: 1, City: "Moscow", Available: true, Interests: []domain.Interest{"Go"}},
		{UserID: 2, City: "Moscow", Available: true, Interests: []domain.Interest{"Go"}},
	})
	store := NewPairStore()
	react := usecase.NewReactService(users, profiles, store)

	const goroutines = 16
	const repeats = 50

	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		from, to := 1, 2
		if g%2 == 1 {
			from, to = 2, 1
		}

		wg.Add(1)
		go func(from, to int) {
			defer wg.Done()
			for i := 0; i < repeats; i++ {
				state, err := react.React(context.Background(), from, to)
				if err != nil {
					t.Errorf("React(%d -> %d): %v", from, to, err)
					return
				}
				if !state.HasLike(from, to) {
					t.Errorf("React(%d -> %d): own like missing from %#v", from, to, state)
					return
				}
			}
		}(from, to)
	}
	wg.Wait()

	pair, err := interaction.NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	state, found, err := store.Load(pair)
	if err != nil || !found {
		t.Fatalf("final Load: %#v, %v, %v", state, found, err)
	}
	if !state.IsMutual() {
		t.Fatalf("expected mutual pair after concurrent reactions, got %#v", state)
	}
}
