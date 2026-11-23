package memory

import (
	"context"
	"errors"
	"sync"

	"github.com/lasofiko/LinkUp/internal/interaction"
)

var errStateWithoutPair = errors.New("state has no pair")

// PairStore is the single shared store for ReactService, PairDiscoveryService
// and pair reads. Do not use the legacy Likes and Matches as a second truth.
type PairStore struct {
	mu     sync.RWMutex
	states map[interaction.Pair]interaction.State
	locks  map[interaction.Pair]chan struct{}
}

func NewPairStore() *PairStore {
	return &PairStore{
		states: make(map[interaction.Pair]interaction.State),
		locks:  make(map[interaction.Pair]chan struct{}),
	}
}

func (s *PairStore) Load(pair interaction.Pair) (interaction.State, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	state, found := s.states[pair]
	return state, found, nil
}

func (s *PairStore) Save(state interaction.State) error {
	if state.Pair == (interaction.Pair{}) {
		return errStateWithoutPair
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.states[state.Pair] = state
	return nil
}

func (s *PairStore) Update(
	ctx context.Context,
	pair interaction.Pair,
	apply func(interaction.State) (interaction.State, bool, error),
) (interaction.State, error) {
	lock := s.lockFor(pair)
	select {
	case <-ctx.Done():
		return interaction.State{}, ctx.Err()
	case <-lock:
	}
	defer func() { lock <- struct{}{} }()

	if err := ctx.Err(); err != nil {
		return interaction.State{}, err
	}

	s.mu.RLock()
	state, found := s.states[pair]
	s.mu.RUnlock()
	if !found {
		state = interaction.State{}
	}

	next, changed, err := apply(state)
	if err != nil {
		return interaction.State{}, err
	}

	if changed {
		if err := ctx.Err(); err != nil {
			return interaction.State{}, err
		}

		s.mu.Lock()
		s.states[pair] = next
		s.mu.Unlock()
	}

	return next, nil
}

func (s *PairStore) lockFor(pair interaction.Pair) chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock, ok := s.locks[pair]
	if !ok {
		lock = make(chan struct{}, 1)
		lock <- struct{}{}
		s.locks[pair] = lock
	}
	return lock
}

// Block uses the same pair lock as ReactService.Update. Repeating it is a no-op.
func (s *PairStore) Block(ctx context.Context, pair interaction.Pair) (interaction.State, error) {
	return s.Update(ctx, pair, func(state interaction.State) (interaction.State, bool, error) {
		if state.Blocked {
			return state, false, nil
		}
		state.Pair = pair
		state.Blocked = true
		return state, true, nil
	})
}

var _ interaction.Store = (*PairStore)(nil)
