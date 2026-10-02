package matching

import (
	"reflect"
	"testing"

	"github.com/lasofiko/LinkUp/internal/domain"
)

func TestFilterRestrictedCandidates(t *testing.T) {
	tests := []struct {
		name       string
		candidates []domain.User
		likedIDs   []int
		matchedIDs []int
		wantIDs    []int
	}{
		{
			name: "available candidate remains",
			candidates: []domain.User{
				{ID: 2, Available: true},
			},
			wantIDs: []int{2},
		},
		{
			name: "current user is excluded",
			candidates: []domain.User{
				{ID: 1, Available: true},
			},
			wantIDs: []int{},
		},
		{
			name: "unavailable candidate is excluded",
			candidates: []domain.User{
				{ID: 2, Available: false},
			},
			wantIDs: []int{},
		},
		{
			name: "liked candidate is excluded",
			candidates: []domain.User{
				{ID: 2, Available: true},
			},
			likedIDs: []int{2},
			wantIDs:  []int{},
		},
		{
			name: "matched candidate is excluded",
			candidates: []domain.User{
				{ID: 2, Available: true},
			},
			matchedIDs: []int{2},
			wantIDs:    []int{},
		},
		{
			name: "all candidates are excluded",
			candidates: []domain.User{
				{ID: 1, Available: true},
				{ID: 2, Available: false},
				{ID: 3, Available: true},
				{ID: 4, Available: true},
			},
			likedIDs:   []int{3},
			matchedIDs: []int{4},
			wantIDs:    []int{},
		},
		{
			name:    "empty input",
			wantIDs: []int{},
		},
		{
			name: "remaining candidates keep their order",
			candidates: []domain.User{
				{ID: 5, Available: true},
				{ID: 3, Available: true},
				{ID: 2, Available: true},
			},
			likedIDs: []int{3},
			wantIDs:  []int{5, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterRestrictedCandidates(
				1,
				tt.candidates,
				tt.likedIDs,
				tt.matchedIDs,
			)

			if got == nil {
				t.Fatal("expected a non-nil result slice")
			}

			gotIDs := make([]int, 0, len(got))
			for _, user := range got {
				gotIDs = append(gotIDs, user.ID)
			}

			if !reflect.DeepEqual(gotIDs, tt.wantIDs) {
				t.Fatalf("got IDs %v, want %v", gotIDs, tt.wantIDs)
			}
		})
	}
}

func TestFilterRestrictedCandidatesDoesNotMutateInput(t *testing.T) {
	candidates := []domain.User{
		{ID: 1, Available: true, Interests: []string{"Music"}},
		{ID: 2, Available: true, Interests: []string{"Go", "Books"}},
		{ID: 3, Available: true, Interests: []string{"Art"}},
	}

	wantCandidates := []domain.User{
		{ID: 1, Available: true, Interests: []string{"Music"}},
		{ID: 2, Available: true, Interests: []string{"Go", "Books"}},
		{ID: 3, Available: true, Interests: []string{"Art"}},
	}

	likedIDs := []int{3, 8}
	matchedIDs := []int{9, 7}

	got := FilterRestrictedCandidates(
		1,
		candidates,
		likedIDs,
		matchedIDs,
	)

	if !reflect.DeepEqual(candidates, wantCandidates) {
		t.Fatal("candidates changed during filtering")
	}

	if !reflect.DeepEqual(likedIDs, []int{3, 8}) {
		t.Fatal("liked IDs changed during filtering")
	}

	if !reflect.DeepEqual(matchedIDs, []int{9, 7}) {
		t.Fatal("matched IDs changed during filtering")
	}

	if len(got) != 1 || got[0].ID != 2 {
		t.Fatalf("unexpected result: %#v", got)
	}

	got[0].Name = "Changed"
	got[0].Interests[0] = "Changed"

	if !reflect.DeepEqual(candidates, wantCandidates) {
		t.Fatal("result shares mutable data with input")
	}
}
