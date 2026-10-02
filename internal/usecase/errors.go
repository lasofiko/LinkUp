package usecase

import "errors"

var (
	ErrReadLikes = errors.New("read likes failed")
	ErrReadMatch = errors.New("read match failed")
	ErrSaveMatch = errors.New("save match failed")
)