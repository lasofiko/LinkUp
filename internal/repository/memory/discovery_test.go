package memory

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestDiscoverySources(t *testing.T) {
	ctx := context.Background()

	likes := NewLikes()
	matches := NewMatches()

	likes.Add(1, 4)
	likes.Add(1, 2)
	likes.Add(1, 2)

	if err := matches.Save(ctx, 1, 3); err != nil {
		t.Fatal(err)
	}

	if err := matches.Save(ctx, 3, 1); err != nil {
		t.Fatal(err)
	}

	if err := matches.Save(ctx, 5, 1); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		read func(context.Context, int) ([]int, error)
		id   int
		want []int
	}{
		{
			name: "outgoing likes are unique and sorted",
			read: likes.GetLikedUserIDs,
			id:   1,
			want: []int{2, 4},
		},
		{
			name: "like is directed",
			read: likes.GetLikedUserIDs,
			id:   2,
			want: []int{},
		},
		{
			name: "matches include both participant positions",
			read: matches.GetMatchedUserIDs,
			id:   1,
			want: []int{3, 5},
		},
		{
			name: "match is visible to the other participant",
			read: matches.GetMatchedUserIDs,
			id:   3,
			want: []int{1},
		},
		{
			name: "unknown user has no matches",
			read: matches.GetMatchedUserIDs,
			id:   999,
			want: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.read(ctx, tt.id)
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

			again, err := tt.read(ctx, tt.id)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(again, tt.want) {
				t.Fatal("caller changed repository data")
			}
		})
	}
}

func TestDiscoverySourcesCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	likes := NewLikes()
	matches := NewMatches()

	tests := []struct {
		name string
		read func(context.Context, int) ([]int, error)
	}{
		{
			name: "likes",
			read: likes.GetLikedUserIDs,
		},
		{
			name: "matches",
			read: matches.GetMatchedUserIDs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.read(ctx, 1)

			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expected context.Canceled, got %v", err)
			}

			if got != nil {
				t.Fatalf("expected nil result on error, got %v", got)
			}
		})
	}
}
