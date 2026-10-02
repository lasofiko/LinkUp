package memory

import (
	"context"
	"sync"
)

type Likes struct {
	mu    sync.RWMutex
	likes map[[2]int]bool
}

func NewLikes() *Likes {
	return &Likes{likes: make(map[[2]int]bool)}
}

func (l *Likes) ExistsLike(_ context.Context, from, to int) (bool, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.likes[[2]int{from, to}], nil
}

func (l *Likes) Add(from, to int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.likes[[2]int{from, to}] = true
}