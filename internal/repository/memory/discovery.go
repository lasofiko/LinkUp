package memory

import (
	"context"
	"slices"
)

func (l *Likes) GetLikedUserIDs(
	ctx context.Context,
	userID int,
) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	result := make([]int, 0)

	for pair, exists := range l.likes {
		if exists && pair[0] == userID {
			result = append(result, pair[1])
		}
	}

	slices.Sort(result)

	return result, nil
}

func (m *Matches) GetMatchedUserIDs(
	ctx context.Context,
	userID int,
) ([]int, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]int, 0)

	for match, exists := range m.matches {
		if !exists {
			continue
		}

		switch {
		case match.User1ID == userID:
			result = append(result, match.User2ID)
		case match.User2ID == userID:
			result = append(result, match.User1ID)
		}
	}

	slices.Sort(result)

	return result, nil
}
