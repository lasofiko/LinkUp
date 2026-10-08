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
