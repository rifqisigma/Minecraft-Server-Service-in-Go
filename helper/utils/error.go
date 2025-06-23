package utils

import "errors"

var (
	ErrNotCreator   = errors.New("kau bukan kreator")
	ErrUserNotFound = errors.New("user not found")
	ErrInvalidEmail = errors.New("invalid email")
	ErrUnauhorized  = errors.New("you unauthorized for this action")
)
