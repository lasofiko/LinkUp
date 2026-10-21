package domain

type EligibilityStatus string

const (
	EligibilityAllowed      EligibilityStatus = "allowed"
	EligibilityAlreadyLiked EligibilityStatus = "already_liked"
	EligibilityUnavailable  EligibilityStatus = "unavailable"
)

func CheckCandidateEligibility(
	userID int,
	candidate Profile,
	alreadyLiked bool,
) EligibilityStatus {
	if userID == candidate.UserID {
		return EligibilityUnavailable
	}

	if !candidate.Available {
		return EligibilityUnavailable
	}

	if alreadyLiked {
		return EligibilityAlreadyLiked
	}

	return EligibilityAllowed
}
