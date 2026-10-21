package matching

import (
	"slices"
	"sort"
	"strings"

	"github.com/lasofiko/LinkUp/internal/domain"
)

func MatchCandidates(target domain.Profile, candidates []domain.Profile) []domain.Profile { //matches profiles from the same city (w/o considering upper/lower case), at least one matching interest. output is stable sorted by the amount of matched interests (non-ai comment btw)
	result := []domain.Profile{}

	targetCity := normalize(target.City)
	targetInterests := interestSet(target.Interests)
	if targetCity == "" || len(targetInterests) == 0 {
		return result
	}

	type scored struct {
		profile domain.Profile
		score   int
	}
	var found []scored

	for _, c := range candidates {
		if c.UserID == target.UserID {
			continue
		}
		if !c.Available {
			continue
		}
		if normalize(c.City) != targetCity {
			continue
		}
		score := countShared(targetInterests, c.Interests)
		if score == 0 {
			continue
		}
		found = append(found, scored{profile: c, score: score})
	}

	sort.SliceStable(found, func(i, j int) bool {
		return found[i].score > found[j].score
	})

	for _, f := range found {
		p := f.profile
		p.Interests = slices.Clone(f.profile.Interests)
		result = append(result, p)
	}
	return result
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func interestSet(list []domain.Interest) map[string]struct{} {
	set := make(map[string]struct{}, len(list))
	for _, s := range list {
		if n := normalize(string(s)); n != "" {
			set[n] = struct{}{}
		}
	}
	return set
}

func countShared(target map[string]struct{}, other []domain.Interest) int {
	count := 0
	for interest := range interestSet(other) {
		if _, ok := target[interest]; ok {
			count++
		}
	}
	return count
}
