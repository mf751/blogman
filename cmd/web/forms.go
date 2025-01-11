package main

import "github.com/mf751/blogman/interanl/validator"

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}
