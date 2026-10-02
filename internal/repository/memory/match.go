package memory

import (
    "context"
    "sync"

    "github.com/lasofiko/LinkUp/internal/domain"
)

type Matches struct {
    mu      sync.RWMutex
    matches map[domain.Match]bool
}

func NewMatches() *Matches {
    return &Matches{matches: make(map[domain.Match]bool)}
}

func (m *Matches) ExistsMatch(_ context.Context, a, b int) (bool, error) {
    match, err := domain.NewMatch(a, b)
    if err != nil {
        return false, err
    }
    m.mu.RLock()
    defer m.mu.RUnlock()
    return m.matches[match], nil
}

func (m *Matches) Save(_ context.Context, a, b int) error {
    match, err := domain.NewMatch(a, b)
    if err != nil {
        return err
    }
    m.mu.Lock()
    defer m.mu.Unlock()
    m.matches[match] = true
    return nil
}

func (m *Matches) Count() int {
    m.mu.RLock()
    defer m.mu.RUnlock()
    return len(m.matches)
}