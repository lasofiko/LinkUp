package interaction

import (
	"errors"
	"testing"
)

func TestApplyLikeFirstLike(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name           string
		from, to       int
		wantFirstLike  bool
		wantSecondLike bool
		wantFromTo     bool
		wantToFrom     bool
	}{
		{
			name:           "smaller likes larger",
			from:           1,
			to:             2,
			wantFirstLike:  true,
			wantSecondLike: false,
			wantFromTo:     true,
			wantToFrom:     false,
		},
		{
			name:           "larger likes smaller",
			from:           2,
			to:             1,
			wantFirstLike:  false,
			wantSecondLike: true,
			wantFromTo:     true,
			wantToFrom:     false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, changed, err := ApplyLike(State{}, tt.from, tt.to)
			if err != nil || !changed {
				t.Fatalf("first like: changed=%v, err=%v", changed, err)
			}
			if got.Pair != (Pair{FirstUserID: 1, SecondUserID: 2}) {
				t.Fatalf("canonical pair: %#v", got.Pair)
			}
			if got.FirstLikesSecond != tt.wantFirstLike || got.SecondLikesFirst != tt.wantSecondLike {
				t.Fatalf("directions: %#v", got)
			}
			if got.HasLike(tt.from, tt.to) != tt.wantFromTo || got.HasLike(tt.to, tt.from) != tt.wantToFrom {
				t.Fatalf("HasLike lost direction: %#v", got)
			}
			if got.IsMutual() {
				t.Fatal("first like must not be mutual")
			}
		})
	}
}

func TestApplyLikeReplyLike(t *testing.T) {
	t.Parallel()

	first, changed, err := ApplyLike(State{}, 1, 2)
	if err != nil || !changed || first.IsMutual() {
		t.Fatalf("first like: %#v, changed=%v, err=%v", first, changed, err)
	}

	reply, changed, err := ApplyLike(first, 2, 1)
	if err != nil || !changed {
		t.Fatalf("reply like: %#v, changed=%v, err=%v", reply, changed, err)
	}
	if reply.Pair != first.Pair {
		t.Fatalf("reply must reuse the same pair key: %#v", reply.Pair)
	}
	if !reply.FirstLikesSecond || !reply.SecondLikesFirst {
		t.Fatalf("reply must keep both directed likes: %#v", reply)
	}
	if !reply.HasLike(1, 2) || !reply.HasLike(2, 1) {
		t.Fatalf("directions after reply: %#v", reply)
	}
}

func TestApplyLikeMutuality(t *testing.T) {
	t.Parallel()

	oneWay, _, err := ApplyLike(State{}, 2, 1)
	if err != nil || oneWay.IsMutual() {
		t.Fatalf("one-way: %#v, err=%v", oneWay, err)
	}

	mutual, changed, err := ApplyLike(oneWay, 1, 2)
	if err != nil || !changed || !mutual.IsMutual() {
		t.Fatalf("mutual: %#v, changed=%v, err=%v", mutual, changed, err)
	}
	if mutual.Pair != (Pair{FirstUserID: 1, SecondUserID: 2}) {
		t.Fatalf("mutual pair key: %#v", mutual.Pair)
	}
}

func TestApplyLikeRepeat(t *testing.T) {
	t.Parallel()

	first, _, err := ApplyLike(State{}, 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	repeated, changed, err := ApplyLike(first, 1, 2)
	if err != nil || changed || repeated != first {
		t.Fatalf("repeat first like: %#v, changed=%v, err=%v", repeated, changed, err)
	}

	mutual, _, err := ApplyLike(first, 2, 1)
	if err != nil {
		t.Fatal(err)
	}

	for _, command := range [][2]int{{1, 2}, {2, 1}} {
		got, changed, err := ApplyLike(mutual, command[0], command[1])
		if err != nil || changed || got != mutual || !got.IsMutual() {
			t.Fatalf("repeat %v: %#v, changed=%v, err=%v", command, got, changed, err)
		}
	}
}

func TestApplyLikeForbiddenTransition(t *testing.T) {
	t.Parallel()

	state, _, err := ApplyLike(State{}, 1, 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("same user", func(t *testing.T) {
		t.Parallel()
		got, changed, err := ApplyLike(state, 1, 1)
		if !errors.Is(err, ErrSameUser) || changed || got != (State{}) {
			t.Fatalf("got %#v, changed=%v, err=%v", got, changed, err)
		}
	})

	t.Run("other pair", func(t *testing.T) {
		t.Parallel()
		got, changed, err := ApplyLike(state, 1, 3)
		if !errors.Is(err, ErrStatePair) || changed || got != (State{}) {
			t.Fatalf("got %#v, changed=%v, err=%v", got, changed, err)
		}
	})
}

func TestApplyLikeDoesNotModifyInput(t *testing.T) {
	t.Parallel()

	input := State{
		Pair:             Pair{FirstUserID: 1, SecondUserID: 2},
		FirstLikesSecond: true,
	}
	before := input

	next, changed, err := ApplyLike(input, 2, 1)
	if err != nil || !changed || !next.IsMutual() {
		t.Fatalf("next=%#v, changed=%v, err=%v", next, changed, err)
	}
	if input != before {
		t.Fatalf("input mutated: %#v", input)
	}
	if !before.FirstLikesSecond || before.SecondLikesFirst || before.IsMutual() {
		t.Fatalf("original snapshot changed: %#v", before)
	}

	failed, changed, err := ApplyLike(input, 3, 4)
	if !errors.Is(err, ErrStatePair) || changed || failed != (State{}) {
		t.Fatalf("forbidden: %#v, changed=%v, err=%v", failed, changed, err)
	}
	if input != before {
		t.Fatalf("failed transition mutated input: %#v", input)
	}
}
