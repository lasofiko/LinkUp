package memory

import (
	"reflect"
	"testing"
)

func TestDiscoveryRepository(t *testing.T) {
	repo := NewDiscoveryRepository()

	repo.AddLike(1, 2)
	repo.AddLike(1, 2)
	repo.AddMatch(1, 3)
	repo.AddMatch(1, 3)

	tests := []struct {
		name string
		read func(int) ([]int, error)
		id   int
		want []int
	}{
		{"outgoing like", repo.GetLikedUserIDs, 1, []int{2}},
		{"like is directed", repo.GetLikedUserIDs, 2, nil},
		{"match forward", repo.GetMatchedUserIDs, 1, []int{3}},
		{"match reverse", repo.GetMatchedUserIDs, 3, []int{1}},
		{"unknown user", repo.GetMatchedUserIDs, 999, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.read(tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}

			if len(got) == 0 {
				return
			}

			got[0] = 999

			again, err := tt.read(tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(again, tt.want) {
				t.Fatal("caller changed repository data")
			}
		})
	}
}
