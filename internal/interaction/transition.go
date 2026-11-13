package interaction

// ApplyLike is pure: it never changes the input state and performs no I/O.
// Changed is false when the same directed like was already recorded.
func ApplyLike(state State, fromUserID, toUserID int) (next State, changed bool, err error) {
	pair, err := NewPair(fromUserID, toUserID)
	if err != nil {
		return State{}, false, err
	}

	if state.Pair != (Pair{}) && state.Pair != pair {
		return State{}, false, ErrStatePair
	}

	next = state
	next.Pair = pair
	if fromUserID == pair.FirstUserID {
		if next.FirstLikesSecond {
			return next, false, nil
		}
		next.FirstLikesSecond = true
		return next, true, nil
	}

	if next.SecondLikesFirst {
		return next, false, nil
	}
	next.SecondLikesFirst = true
	return next, true, nil
}
