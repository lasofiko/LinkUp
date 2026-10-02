package domain

type Match struct {
	User1ID int
	User2ID int
}

func NewMatch(a, b int) (Match, error) {
	if a == b {
		return Match{}, ErrSelfMatch
	}

	if a > b {
		a, b = b, a
	}

	return Match{
		User1ID: a,
		User2ID: b,
	}, nil
}
