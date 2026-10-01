package domain

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrProfileNotFound = errors.New("profile not found")
	ErrEmptyUserID     = errors.New("user id must not be empty")
)
