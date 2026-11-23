package profilemock

type profileResponse struct {
	UserID    int      `json:"user_id"`
	City      string   `json:"city"`
	Interests []string `json:"interests"`
	Available bool     `json:"available"`
}

func demoProfile(userID int) (profileResponse, bool) {
	switch userID {
	case 1:
		return profileResponse{
			UserID:    1,
			City:      "Moscow",
			Interests: []string{"Go", "Music"},
			Available: true,
		}, true

	case 2:
		return profileResponse{
			UserID:    2,
			City:      "Moscow",
			Interests: []string{"Go"},
			Available: true,
		}, true

	case 3:
		return profileResponse{
			UserID:    3,
			City:      "Moscow",
			Interests: []string{},
			Available: false,
		}, true

	case 4, 5:
		return profileResponse{
			UserID:    userID,
			City:      "Moscow",
			Interests: []string{},
			Available: true,
		}, true

	default:
		return profileResponse{}, false
	}
}
