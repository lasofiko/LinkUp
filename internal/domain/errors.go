package domain

import "errors"

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrProfileNotFound = errors.New("profile not found")
	ErrSelfLike        = errors.New("cannot like yourself")
	ErrEmptyUserID     = errors.New("user id must not be empty")
)
