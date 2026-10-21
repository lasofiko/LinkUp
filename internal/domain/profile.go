package domain

type Profile struct {
	UserID    int
	City      string
	Interests []Interest
	Available bool
}
