package matching

import "testing"

func TestDetermineMatch_NoLikes(t *testing.T) {
	if got := DetermineMatch(false, false, false); got != NoMatch {
		t.Fatalf("want NoMatch, got %v", got)
	}
}

func TestDetermineMatch_LikeOnlyOneWay(t *testing.T) {
	cases := []struct {
		name           string
		userLikedOther bool
		otherLikedUser bool
	}{
		{"A liked B only", true, false},
		{"B liked A only", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := DetermineMatch(tc.userLikedOther, tc.otherLikedUser, false); got != NoMatch {
				t.Fatalf("want NoMatch, got %v", got)
			}
		})
	}
}

func TestDetermineMatch_MutualLikes(t *testing.T) {
	if got := DetermineMatch(true, true, false); got != MatchCreated {
		t.Fatalf("want MatchCreated, got %v", got)
	}
}

func TestDetermineMatch_ExistingMatch(t *testing.T) {
	if got := DetermineMatch(true, true, true); got != ExistingMatch {
		t.Fatalf("want ExistingMatch, got %v", got)
	}
}

func TestDetermineMatch_ExistingMatchWithoutLikes(t *testing.T) {
	if got := DetermineMatch(false, false, true); got != ExistingMatch {
		t.Fatalf("want ExistingMatch, got %v", got)
	}
}

func TestDetermineMatch_DoesNotModifyInput(t *testing.T) {
	a, b, c := true, false, true
	_ = DetermineMatch(a, b, c)
	if a != true || b != false || c != true {
		t.Fatalf("input must not be modified")
	}
}