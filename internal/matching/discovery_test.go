package matching

import (
	"reflect"
	"slices"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
)

func profile(id int, city string, available bool, interests ...string) domain.Profile {
	list := make([]domain.Interest, 0, len(interests))
	for _, s := range interests {
		list = append(list, domain.Interest(s))
	}
	return domain.Profile{UserID: id, City: city, Interests: list, Available: available}
}

func ids(profiles []domain.Profile) []int {
	out := make([]int, 0, len(profiles))
	for _, p := range profiles {
		out = append(out, p.UserID)
	}
	return out
}

func cloneProfiles(in []domain.Profile) []domain.Profile {
	out := make([]domain.Profile, len(in))
	for i, p := range in {
		p.Interests = slices.Clone(p.Interests)
		out[i] = p
	}
	return out
}

func TestMatchCandidates(t *testing.T) {
	target := profile(1, "Москва", true, "go", "music")

	tests := []struct {
		name       string
		target     domain.Profile
		candidates []domain.Profile
		want       []int
	}{
		{
			name:       "full candidate match",
			target:     target,
			candidates: []domain.Profile{profile(2, "Москва", true, "go", "chess")},
			want:       []int{2},
		},
		{
			name:       "no matching interests",
			target:     target,
			candidates: []domain.Profile{profile(2, "Москва", true, "chess", "cooking")},
			want:       []int{},
		},
		{
			name:       "иmatching interests but different cities",
			target:     target,
			candidates: []domain.Profile{profile(2, "Санкт-Петербург", true, "go", "music")},
			want:       []int{},
		},
		{
			name:   "the city is not set",
			target: profile(1, "", true, "go"),
			candidates: []domain.Profile{
				profile(2, "Москва", true, "go"),
				profile(3, "", true, "go"),
			},
			want: []int{},
		},
		{
			name:       "the starting list is empty",
			target:     target,
			candidates: nil,
			want:       []int{},
		},
		{
			name:   "not a single profile matches",
			target: target,
			candidates: []domain.Profile{
				profile(1, "Москва", true, "go"),    // сам пользователь
				profile(2, "Казань", true, "go"),    // другой город
				profile(3, "Москва", false, "go"),   // недоступен
				profile(4, "Москва", true, "chess"), // нет общих интересов
			},
			want: []int{},
		},
		{
			name:       "not availible is skipped",
			target:     target,
			candidates: []domain.Profile{profile(2, "Москва", false, "go")},
			want:       []int{},
		},
		{
			name:       "user doesn't match with himself",
			target:     target,
			candidates: []domain.Profile{profile(1, "Москва", true, "go", "music")},
			want:       []int{},
		},
		{
			name:       "excessive spaces and upper/lower letters don't matter",
			target:     target,
			candidates: []domain.Profile{profile(2, " москва ", true, "  GO ")},
			want:       []int{2},
		},
		{
			name:   "repeating interest counts once",
			target: target,
			candidates: []domain.Profile{
				profile(2, "Москва", true, "go", "go"),    // 1 общий интерес
				profile(3, "Москва", true, "go", "music"), // 2 общих интереса
			},
			want: []int{3, 2},
		},
		{
			name:   "output list sort by ammount of interests, stable sort",
			target: profile(1, "Москва", true, "go", "music", "chess"),
			candidates: []domain.Profile{
				profile(2, "Москва", true, "go"),                   // 1
				profile(3, "Москва", true, "go", "music", "chess"), // 3
				profile(4, "Москва", true, "music"),                // 1
				profile(5, "Москва", true, "go", "music"),          // 2
			},
			want: []int{3, 5, 2, 4},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := MatchCandidates(tc.target, tc.candidates)

			if got == nil {
				t.Fatal("shouldn't be nil even if no matches")
			}
			if !slices.Equal(ids(got), tc.want) {
				t.Fatalf("got %v, want %v", ids(got), tc.want)
			}
		})
	}
}

func TestMatchCandidates_DoesNotModifyInput(t *testing.T) {
	target := profile(1, "Москва", true, "go", "music")
	candidates := []domain.Profile{
		profile(2, "Москва", true, "go"),
		profile(3, "Москва", true, "go", "music"),
		profile(4, "Казань", true, "go"),
	}
	targetBefore := cloneProfiles([]domain.Profile{target})[0]
	candidatesBefore := cloneProfiles(candidates)

	got := MatchCandidates(target, candidates)

	if !reflect.DeepEqual(target, targetBefore) {
		t.Error("function changed target profile")
	}
	if !reflect.DeepEqual(candidates, candidatesBefore) {
		t.Error("function changed input list")
	}

	got[0].Interests[0] = "changed"
	if !reflect.DeepEqual(candidates, candidatesBefore) {
		t.Error("input data was affected")
	}
}
