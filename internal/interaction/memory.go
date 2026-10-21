package interaction

import "sync"

type Memory struct {
	mu     sync.RWMutex
	states map[Pair]State
}

var (
	_ Reader = (*Memory)(nil)
	_ Store  = (*Memory)(nil)
)

func NewMemory() *Memory {
	return &Memory{states: make(map[Pair]State)}
}

func (m *Memory) Load(pair Pair) (State, bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	state, found := m.states[pair]
	return state, found, nil
}

func (m *Memory) Save(state State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.states[state.Pair] = state
	return nil
}

type FailingSave struct {
	*Memory
	Err error
}

var _ Store = FailingSave{}

func (s FailingSave) Save(State) error {
	return s.Err
}
