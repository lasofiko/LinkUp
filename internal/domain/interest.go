package domain

import (
	"errors"
	"strings"
)

var ErrEmptyInterest = errors.New("interest must not be empty")

type Interest string

func NewInterest(value string) (Interest, error) {
	value = strings.TrimSpace(value)

	if value == "" {
		return "", ErrEmptyInterest
	}

	return Interest(value), nil
}
