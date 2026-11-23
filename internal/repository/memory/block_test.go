package memory

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lasofiko/LinkUp/internal/interaction"
)

func TestBlockUpdateErrorDoesNotCommit(t *testing.T) {
	store := NewPairStore()
	pair, _ := interaction.NewPair(1, 2)
	mutual := interaction.State{Pair: pair, FirstLikesSecond: true, SecondLikesFirst: true}
	if err := store.Save(mutual); err != nil {
		t.Fatal(err)
	}
	failure := errors.New("failure before commit")
	_, err := store.Update(context.Background(), pair, func(state interaction.State) (interaction.State, bool, error) {
		state.Blocked = true
		return state, true, failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("got %v", err)
	}
	got, _, err := store.Load(pair)
	if err != nil || got != mutual || !got.IsMutual() {
		t.Fatalf("partial change: %+v, %v", got, err)
	}
}

func TestBlockWaitCancellationReleasesPair(t *testing.T) {
	store := NewPairStore()
	pair, _ := interaction.NewPair(1, 2)
	entered := make(chan struct{})
	release := make(chan struct{})
	finished := make(chan error, 1)
	go func() {
		_, err := store.Update(context.Background(), pair, func(state interaction.State) (interaction.State, bool, error) {
			close(entered)
			<-release
			return state, false, nil
		})
		finished <- err
	}()
	<-entered
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := store.Block(ctx, pair); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("waiting: %v", err)
	}
	close(release)
	if err := <-finished; err != nil {
		t.Fatal(err)
	}
	if _, err := store.Block(context.Background(), pair); err != nil {
		t.Fatalf("lock not released: %v", err)
	}
}
