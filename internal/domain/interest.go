package domain

import "strings"

type Interest string

func NewInterest(value string) (Interest, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", ErrEmptyInterest
	}

	return Interest(value), nil
}
