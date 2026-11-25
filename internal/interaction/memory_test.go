package interaction

import (
	"errors"
	"testing"
)

func TestMemoryAbsence(t *testing.T) {
	t.Parallel()

	store := NewMemory()
	pair, err := NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}

	got, found, err := store.Load(pair)
	if err != nil || found || got != (State{}) {
		t.Fatalf("absent pair: %#v, found=%v, err=%v", got, found, err)
	}
}

func TestMemorySaveThenLoad(t *testing.T) {
	t.Parallel()

	store := NewMemory()
	pair, err := NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	want := State{Pair: pair, FirstLikesSecond: true}

	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, found, err := store.Load(pair)
	if err != nil || !found || got != want {
		t.Fatalf("Save → Load: %#v, found=%v, err=%v", got, found, err)
	}
	if !got.HasLike(1, 2) || got.HasLike(2, 1) || got.IsMutual() {
		t.Fatalf("B read of one-way like: %#v", got)
	}
}

func TestMemoryReplaceState(t *testing.T) {
	t.Parallel()

	store := NewMemory()
	pair, err := NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	first := State{Pair: pair, FirstLikesSecond: true}
	replaced := State{Pair: pair, FirstLikesSecond: true, SecondLikesFirst: true}

	if err := store.Save(first); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(replaced); err != nil {
		t.Fatal(err)
	}
	got, found, err := store.Load(pair)
	if err != nil || !found || got != replaced {
		t.Fatalf("replaced state: %#v, found=%v, err=%v", got, found, err)
	}
	if !got.IsMutual() || !got.HasLike(1, 2) || !got.HasLike(2, 1) {
		t.Fatalf("B read of mutual like: %#v", got)
	}
}

func TestMemoryReverseParticipantOrder(t *testing.T) {
	t.Parallel()

	store := NewMemory()
	forward, err := NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	reverse, err := NewPair(2, 1)
	if err != nil {
		t.Fatal(err)
	}
	want := State{Pair: forward, SecondLikesFirst: true}

	if err := store.Save(want); err != nil {
		t.Fatal(err)
	}
	got, found, err := store.Load(reverse)
	if err != nil || !found || got != want || reverse != forward {
		t.Fatalf("reverse order: %#v, found=%v, err=%v, keys %#v %#v", got, found, err, forward, reverse)
	}
	if !got.HasLike(2, 1) || got.HasLike(1, 2) {
		t.Fatalf("direction after reverse key: %#v", got)
	}
}

func TestMemoryPairIndependence(t *testing.T) {
	t.Parallel()

	store := NewMemory()
	one, err := NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	other, err := NewPair(1, 3)
	if err != nil {
		t.Fatal(err)
	}
	first := State{Pair: one, FirstLikesSecond: true}
	second := State{Pair: other, SecondLikesFirst: true}

	if err := store.Save(first); err != nil {
		t.Fatal(err)
	}
	if err := store.Save(second); err != nil {
		t.Fatal(err)
	}

	gotOne, foundOne, err := store.Load(one)
	if err != nil || !foundOne || gotOne != first {
		t.Fatalf("pair 1-2: %#v, found=%v, err=%v", gotOne, foundOne, err)
	}
	gotOther, foundOther, err := store.Load(other)
	if err != nil || !foundOther || gotOther != second {
		t.Fatalf("pair 1-3: %#v, found=%v, err=%v", gotOther, foundOther, err)
	}
}

func TestMemoryDataIsolation(t *testing.T) {
	t.Parallel()

	store := NewMemory()
	pair, err := NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	input := State{Pair: pair, FirstLikesSecond: true}
	if err := store.Save(input); err != nil {
		t.Fatal(err)
	}
	input.SecondLikesFirst = true
	input.FirstLikesSecond = false

	got, found, err := store.Load(pair)
	if err != nil || !found || got.FirstLikesSecond != true || got.SecondLikesFirst || got.IsMutual() {
		t.Fatalf("saved input leaked: %#v, found=%v, err=%v", got, found, err)
	}

	got.SecondLikesFirst = true
	got.FirstLikesSecond = false
	again, found, err := store.Load(pair)
	if err != nil || !found || again.FirstLikesSecond != true || again.SecondLikesFirst || again.IsMutual() {
		t.Fatalf("loaded value leaked: %#v, found=%v, err=%v", again, found, err)
	}
}

func TestFailingSaveDoesNotChangeMemory(t *testing.T) {
	t.Parallel()

	memory := NewMemory()
	pair, err := NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	saved := State{Pair: pair, FirstLikesSecond: true}
	if err := memory.Save(saved); err != nil {
		t.Fatal(err)
	}

	sentinel := errors.New("save failed")
	wrapper := FailingSave{Memory: memory, Err: sentinel}
	if err := wrapper.Save(State{Pair: pair, FirstLikesSecond: true, SecondLikesFirst: true}); !errors.Is(err, sentinel) {
		t.Fatalf("Save error: %v", err)
	}

	got, found, err := wrapper.Load(pair)
	if err != nil || !found || got != saved {
		t.Fatalf("memory changed after failed Save: %#v, found=%v, err=%v", got, found, err)
	}
}
