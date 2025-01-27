package service

import "errors"

var (
	ErrEmailRegistrated = errors.New("email already registrated")
	ErrEmailNotRegistrated = errors.New("email not registrated")
	ErrNoPermission = errors.New("no permission")
	ErrPromoNotFound = errors.New("promo id not fount")
	ErrInvalidMaxCount = errors.New("max count for unique is 1")
)
