package domain

type Profile struct {
	UserId    int
	City      string
	Interests []Interest
	Available bool
}
