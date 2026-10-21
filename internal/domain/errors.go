package domain

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrProfileNotFound     = errors.New("profile not found")
	ErrEmptyUserID         = errors.New("user id must not be empty")
	ErrSelfMatch           = errors.New("cannot create match with yourself")
	ErrEmptyInterest       = errors.New("interest must not be empty")
	ErrSelfReaction        = errors.New("cannot react to yourself")
	ErrInvalidReactionType = errors.New("invalid reaction type")
)
