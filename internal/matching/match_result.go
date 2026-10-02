package matching

import "fmt"

type MatchResult int

const (
	NoMatch MatchResult = iota
	MatchCreated
	ExistingMatch
)

func (r MatchResult) String() string {
	switch r {
	case NoMatch:
		return "NoMatch"
	case MatchCreated:
		return "MatchCreated"
	case ExistingMatch:
		return "ExistingMatch"
	default:
		return fmt.Sprintf("MatchResult(%d)", int(r))
	}
}

func DetermineMatch(userLikedOther, otherLikedUser, matchExists bool) MatchResult {
	if matchExists {
		return ExistingMatch
	}
	if userLikedOther && otherLikedUser {
		return MatchCreated
	}
	return NoMatch
}