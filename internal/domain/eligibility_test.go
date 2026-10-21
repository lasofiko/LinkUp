package domain

import (
	"reflect"
	"testing"
)

func TestCheckCandidateEligibility(t *testing.T) {
	tests := []struct {
		name         string
		userID       int
		candidate    Profile
		alreadyLiked bool
		want         EligibilityStatus
	}{
		{
			name:   "allowed",
			userID: 1,
			candidate: Profile{
				UserID:    2,
				Available: true,
			},
			alreadyLiked: false,
			want:         EligibilityAllowed,
		},
		{
			name:   "self interaction",
			userID: 1,
			candidate: Profile{
				UserID:    1,
				Available: true,
			},
			alreadyLiked: false,
			want:         EligibilityUnavailable,
		},
		{
			name:   "candidate unavailable",
			userID: 1,
			candidate: Profile{
				UserID:    2,
				Available: false,
			},
			alreadyLiked: false,
			want:         EligibilityUnavailable,
		},
		{
			name:   "already liked",
			userID: 1,
			candidate: Profile{
				UserID:    2,
				Available: true,
			},
			alreadyLiked: true,
			want:         EligibilityAlreadyLiked,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckCandidateEligibility(
				tt.userID,
				tt.candidate,
				tt.alreadyLiked,
			)

			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCheckCandidateEligibilityDoesNotModifyInput(t *testing.T) {
	candidate := Profile{
		UserID:    2,
		City:      "Moscow",
		Available: true,
		Interests: []Interest{"go", "linux"},
	}

	originalInterests := append(
		[]Interest(nil),
		candidate.Interests...,
	)

	CheckCandidateEligibility(1, candidate, false)

	if !reflect.DeepEqual(candidate.Interests, originalInterests) {
		t.Fatalf(
			"candidate interests changed: got %v, want %v",
			candidate.Interests,
			originalInterests,
		)
	}
}
