package models

import (
	"errors"
)

type APIError struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrTrainingDayNotFound = errors.New("training day not found")

	ErrDayOrderValueNotPositive = errors.New("dayOrder value should be positive")

	ErrValueExceededMaximum = errors.New("itemsPerPage exceeded maximum")
	ErrInvalidOffset        = errors.New("invalid offset value")
	ErrInvalidItemsPerPage  = errors.New("invalid itemsPerPage value")
	ErrInvalidSortingColumn = errors.New("invalid sorting column")

	ErrUserWasNotDeleted = errors.New("user was not deleted")

	ErrInvalidTokenClaims = errors.New("invalid token claims: unable to assert to jwt.MapClaims")
	ErrUsernameTooShort   = errors.New("username must be at least 6 characters long")
	ErrPasswordTooShort   = errors.New("password must be at least 6 characters long")
	ErrUsernameExists     = errors.New("user with this name is already exists")

	ErrUnknownError = errors.New("unknown error")
)
