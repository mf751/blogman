package main

import "github.com/mf751/blogman/internal/validator"

type userLoginForm struct {
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type userSignupForm struct {
	Name                string `form:"name"`
	UserName            string `form:"username"`
	Email               string `form:"email"`
	Password            string `form:"password"`
	validator.Validator `form:"-"`
}

type blogCreateForm struct {
	Title               string `form:"title"`
	Content             string `form:"content"`
	validator.Validator `form:"-"`
}

type blogUpdateForm struct {
	blogCreateForm
	BlogID              int `form:"blog_id"`
	validator.Validator `form:"-"`
}

type passwordChangeForm struct {
	CurrentPassword    string `form:"current_password"`
	NewPassword        string `form:"new_password"`
	ConfirmNewPassword string `form:"confirm_new_password"`
  validator.Validator
}
