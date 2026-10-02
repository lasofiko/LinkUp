package domain

type User struct {
	ID        int
	Name      string
	Email     string
	City      string
	Interests []string
	Available bool
}
