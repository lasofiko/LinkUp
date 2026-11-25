package interaction

import "testing"

func TestApplyLikeKeepsOnePairAndTwoDirections(t *testing.T) {
	first, changed, err := ApplyLike(State{}, 2, 1)
	if err != nil || !changed {
		t.Fatalf("first like: changed=%v, err=%v", changed, err)
	}
	if first.Pair != (Pair{FirstUserID: 1, SecondUserID: 2}) || !first.SecondLikesFirst {
		t.Fatalf("unexpected first state: %#v", first)
	}

	second, changed, err := ApplyLike(first, 1, 2)
	if err != nil || !changed || !second.IsMutual() {
		t.Fatalf("reply like: state=%#v, changed=%v, err=%v", second, changed, err)
	}

	repeated, changed, err := ApplyLike(second, 1, 2)
	if err != nil || changed || repeated != second {
		t.Fatalf("repeated like: state=%#v, changed=%v, err=%v", repeated, changed, err)
	}
}

func TestNewPairCanonicalKey(t *testing.T) {
	t.Parallel()

	forward, err := NewPair(1, 2)
	if err != nil {
		t.Fatal(err)
	}
	reverse, err := NewPair(2, 1)
	if err != nil {
		t.Fatal(err)
	}
	if forward != reverse || forward != (Pair{FirstUserID: 1, SecondUserID: 2}) {
		t.Fatalf("expected one key for (1,2) and (2,1), got %#v and %#v", forward, reverse)
	}

	if _, err := NewPair(4, 4); err != ErrSameUser {
		t.Fatalf("same user: %v", err)
	}
}

func TestStateKeepsDirectedLikes(t *testing.T) {
	t.Parallel()

	state := State{
		Pair:             Pair{FirstUserID: 1, SecondUserID: 2},
		FirstLikesSecond: true,
	}
	if !state.HasLike(1, 2) || state.HasLike(2, 1) || state.IsMutual() {
		t.Fatalf("one-way like: %#v", state)
	}

	state.SecondLikesFirst = true
	if !state.HasLike(1, 2) || !state.HasLike(2, 1) || !state.IsMutual() {
		t.Fatalf("both directions: %#v", state)
	}
	if state.HasLike(1, 3) || state.HasLike(3, 2) {
		t.Fatal("like must stay on its pair")
	}
}
