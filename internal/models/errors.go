package models

import "errors"

var (
	ErrNoRecord         = errors.New("No record was found")
	ErrWrongCredintials = errors.New("Wrong credintials")
	ErrRepeatedUserName = errors.New("A user with this username Already Exists")
	ErrRepeatedEmail    = errors.New("A user with this email Already Exists")
)
