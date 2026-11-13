package matching_test

import (
	"testing"

	"github.com/lasofiko/LinkUp/internal/matching"
)

func TestDetermineMatch_NoLikes(t *testing.T) {
	got := matching.DetermineMatch(false, false, false)

	if got != matching.NoMatch {
		t.Fatalf("want NoMatch, got %v", got)
	}
}

func TestDetermineMatch_LikeOnlyOneWay(t *testing.T) {
	cases := []struct {
		name           string
		userLikedOther bool
		otherLikedUser bool
	}{
		{
			name:           "A liked B only",
			userLikedOther: true,
			otherLikedUser: false,
		},
		{
			name:           "B liked A only",
			userLikedOther: false,
			otherLikedUser: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := matching.DetermineMatch(
				tc.userLikedOther,
				tc.otherLikedUser,
				false,
			)

			if got != matching.NoMatch {
				t.Fatalf("want NoMatch, got %v", got)
			}
		})
	}
}

func TestDetermineMatch_MutualLikes(t *testing.T) {
	got := matching.DetermineMatch(true, true, false)

	if got != matching.MatchCreated {
		t.Fatalf("want MatchCreated, got %v", got)
	}
}

func TestDetermineMatch_ExistingMatch(t *testing.T) {
	got := matching.DetermineMatch(true, true, true)

	if got != matching.ExistingMatch {
		t.Fatalf("want ExistingMatch, got %v", got)
	}
}

func TestDetermineMatch_ExistingMatchWithoutLikes(t *testing.T) {
	got := matching.DetermineMatch(false, false, true)

	if got != matching.ExistingMatch {
		t.Fatalf("want ExistingMatch, got %v", got)
	}
}

func TestDetermineMatch_DoesNotModifyInput(t *testing.T) {
	a, b, c := true, false, true

	_ = matching.DetermineMatch(a, b, c)

	if !a || b || !c {
		t.Fatal("input must not be modified")
	}
}
