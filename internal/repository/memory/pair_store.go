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
	locks  map[interaction.Pair]*sync.Mutex
}

func NewPairStore() *PairStore {
	return &PairStore{
		states: make(map[interaction.Pair]interaction.State),
		locks:  make(map[interaction.Pair]*sync.Mutex),
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
	lock.Lock()
	defer lock.Unlock()

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

func (s *PairStore) lockFor(pair interaction.Pair) *sync.Mutex {
	s.mu.Lock()
	defer s.mu.Unlock()

	lock, ok := s.locks[pair]
	if !ok {
		lock = &sync.Mutex{}
		s.locks[pair] = lock
	}
	return lock
}

var _ interaction.Store = (*PairStore)(nil)
