package memory

import "slices"

type DiscoveryRepository struct {
	likes   map[int][]int
	matches map[int][]int
}

func NewDiscoveryRepository() *DiscoveryRepository {
	return &DiscoveryRepository{
		likes:   make(map[int][]int),
		matches: make(map[int][]int),
	}
}

func (r *DiscoveryRepository) AddLike(fromID, toID int) {
	if !slices.Contains(r.likes[fromID], toID) {
		r.likes[fromID] = append(r.likes[fromID], toID)
	}
}

func (r *DiscoveryRepository) AddMatch(firstID, secondID int) {
	if !slices.Contains(r.matches[firstID], secondID) {
		r.matches[firstID] = append(r.matches[firstID], secondID)
	}

	if !slices.Contains(r.matches[secondID], firstID) {
		r.matches[secondID] = append(r.matches[secondID], firstID)
	}
}

func (r *DiscoveryRepository) GetLikedUserIDs(
	userID int,
) ([]int, error) {
	return slices.Clone(r.likes[userID]), nil
}

func (r *DiscoveryRepository) GetMatchedUserIDs(
	userID int,
) ([]int, error) {
	return slices.Clone(r.matches[userID]), nil
}
