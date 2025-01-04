package models

import "errors"

var (
	ErrNoRecord         = errors.New("No record was found")
	ErrWrongCredintials = errors.New("Wrong credintials")
)
