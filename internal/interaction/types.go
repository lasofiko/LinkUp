package interaction

import "errors"

var (
	ErrSameUser  = errors.New("interaction requires two different users")
	ErrStatePair = errors.New("state belongs to another pair")
)

// Pair is a canonical, unordered identity for one interaction between two users.
// FirstUserID is always smaller than SecondUserID.
type Pair struct {
	FirstUserID  int
	SecondUserID int
}

func NewPair(userID, otherUserID int) (Pair, error) {
	if userID == otherUserID {
		return Pair{}, ErrSameUser
	}
	if userID > otherUserID {
		userID, otherUserID = otherUserID, userID
	}
	return Pair{FirstUserID: userID, SecondUserID: otherUserID}, nil
}

// State stores two directed likes under one canonical pair key.
// A zero State means that there has been no interaction yet.
type State struct {
	Pair             Pair
	FirstLikesSecond bool
	SecondLikesFirst bool
}

func (s State) HasLike(fromUserID, toUserID int) bool {
	if fromUserID == s.Pair.FirstUserID && toUserID == s.Pair.SecondUserID {
		return s.FirstLikesSecond
	}
	if fromUserID == s.Pair.SecondUserID && toUserID == s.Pair.FirstUserID {
		return s.SecondLikesFirst
	}
	return false
}

func (s State) IsMutual() bool {
	return s.FirstLikesSecond && s.SecondLikesFirst
}

// Reader is the only dependency needed by candidate filtering (contract B).
type Reader interface {
	Load(pair Pair) (State, bool, error)
}

// Store is the dependency used by the reaction command (contract C).
// Save receives a fully formed state and must not partially mutate storage on error.
type Store interface {
	Reader
	Save(state State) error
}
