package memory

import (
	"errors"

	"github.com/lasofiko/LinkUp/internal/interaction"
)

var ErrPairStoreNotImplemented = errors.New("memory pair store is not implemented")

// PairStore is the single shared store for ReactService, PairDiscoveryService
// and pair reads. Do not use the legacy Likes and Matches as a second truth.
// TODO(Арина): implement concurrency-safe Load and Save, and arrange an
// atomic read/transition/write path for concurrent React calls.
type PairStore struct{}

func NewPairStore() *PairStore { return &PairStore{} }

func (*PairStore) Load(interaction.Pair) (interaction.State, bool, error) {
	return interaction.State{}, false, ErrPairStoreNotImplemented
}

func (*PairStore) Save(interaction.State) error { return ErrPairStoreNotImplemented }

var _ interaction.Store = (*PairStore)(nil)
