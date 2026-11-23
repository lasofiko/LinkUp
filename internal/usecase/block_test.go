package usecase

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
	"github.com/lasofiko/LinkUp/internal/interaction"
	"github.com/lasofiko/LinkUp/internal/repository/memory"
)

func blockFixture(t *testing.T) (*ReactService, *BlockService, *memory.PairStore) {
	t.Helper()
	users := memory.NewUserRepository()
	for _, id := range []int{1, 2} {
		if err := users.Create(domain.User{ID: id}); err != nil {
			t.Fatal(err)
		}
	}
	profiles := memory.NewProfiles([]domain.Profile{{UserID: 1, Available: true}, {UserID: 2, Available: true}})
	store := memory.NewPairStore()
	return NewReactService(users, profiles, store), NewBlockService(users, store), store
}

func TestBlockWithoutAndWithMatch(t *testing.T) {
	for _, withMatch := range []bool{false, true} {
		t.Run(map[bool]string{false: "without match", true: "with match"}[withMatch], func(t *testing.T) {
			react, blocks, store := blockFixture(t)
			ctx := context.Background()
			if withMatch {
				for _, ids := range [][2]int{{1, 2}, {2, 1}} {
					if _, err := react.React(ctx, ids[0], ids[1]); err != nil {
						t.Fatal(err)
					}
				}
			}
			state, err := blocks.Block(ctx, 1, 2)
			if err != nil || !state.Blocked || state.IsMutual() {
				t.Fatalf("block: %+v, %v", state, err)
			}
			pair, _ := interaction.NewPair(1, 2)
			stored, found, err := store.Load(pair)
			if err != nil || !found || stored != state {
				t.Fatalf("stored: %+v, %v, %v", stored, found, err)
			}
			if withMatch && (!state.HasLike(1, 2) || !state.HasLike(2, 1)) {
				t.Fatal("block erased existing reactions")
			}
			again, err := blocks.Block(ctx, 2, 1)
			if err != nil || again != state {
				t.Fatalf("repeat: %+v, %v", again, err)
			}
			for _, ids := range [][2]int{{1, 2}, {2, 1}} {
				if _, err := react.React(ctx, ids[0], ids[1]); !errors.Is(err, ErrPairBlocked) {
					t.Fatalf("like after block: %v", err)
				}
			}
		})
	}
}

func TestBlockInvalidInput(t *testing.T) {
	_, blocks, _ := blockFixture(t)
	for _, tt := range []struct {
		actor, target int
		want          error
	}{
		{0, 2, ErrInvalidBlockID}, {1, 0, ErrInvalidBlockID},
		{1, 1, interaction.ErrSameUser}, {1, 3, memory.ErrUserNotFound},
		{3, 2, memory.ErrUserNotFound},
	} {
		if _, err := blocks.Block(context.Background(), tt.actor, tt.target); !errors.Is(err, tt.want) {
			t.Fatalf("Block(%d,%d): %v", tt.actor, tt.target, err)
		}
	}
}

func TestBlockedPairIsHiddenBothWays(t *testing.T) {
	_, blocks, store := blockFixture(t)
	if _, err := blocks.Block(context.Background(), 1, 2); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int{1, 2} {
		other := 3 - id
		got, err := FilterPairCandidates(id, []domain.Profile{{UserID: other, Available: true}}, store)
		if err != nil || len(got) != 0 || got == nil {
			t.Fatalf("discovery for %d: %#v, %v", id, got, err)
		}
	}
}

func TestLikeAgainstBlock(t *testing.T) {
	react, blocks, store := blockFixture(t)
	start := make(chan struct{})
	var wg sync.WaitGroup
	var likeErr, blockErr error
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, likeErr = react.React(context.Background(), 2, 1)
	}()
	go func() {
		defer wg.Done()
		<-start
		_, blockErr = blocks.Block(context.Background(), 1, 2)
	}()
	close(start)
	wg.Wait()
	if blockErr != nil || (likeErr != nil && !errors.Is(likeErr, ErrPairBlocked)) {
		t.Fatalf("like=%v block=%v", likeErr, blockErr)
	}
	pair, _ := interaction.NewPair(1, 2)
	state, _, err := store.Load(pair)
	if err != nil || !state.Blocked || state.IsMutual() {
		t.Fatalf("final state: %+v, %v", state, err)
	}
	if _, err := react.React(context.Background(), 1, 2); !errors.Is(err, ErrPairBlocked) {
		t.Fatalf("new like after block: %v", err)
	}
}
