package internal

import "errors"

var (
	ErrInvalidProtocol = errors.New("NORMALIZER - provided url must be in http/https protocol")
	ErrBadFormat       = errors.New("NORMALIZER - provided url is not in standard format.")
	ErrEmptyHost       = errors.New("NORMALIZER - url's host is empty")
	ErrEmptyInput      = errors.New("NORMALIZER - provided url is empty")
	ErrMaxUrlReached   = errors.New("ENCODER - there is no space for a new url because of counter")
	ErrUrlNotFound     = errors.New("STORE - url not found ")

	ErrInvalidURL = errors.New("URL is not valid")

	ErrUrlNoTime = errors.New("URL does not have creation time (server internal problem)")
)
