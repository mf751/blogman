package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mf751/blogman/internal/models"
	"github.com/mf751/blogman/internal/validator"
)

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	blogs, err := app.blogs.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	users := []*models.User{}
	for _, blog := range blogs {
		if len(blog.Content) > 500 {
			blog.Content = blog.Content[:500] + "..."
		}
		user, err := app.users.Get(blog.UserID)
		user.Created = time.Time{}
		user.Email = ""
		user.HashedPassword = []byte("")
		if err != nil {
			app.serverError(w, err)
			return
		}
		users = append(users, user)
	}

	data := app.newTemplateData(r)
	data.Blogs = blogs
	data.Users = users
	data.Active = "blogs"

	app.render(w, "home", http.StatusOK, data)
}

func pong(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func (app *application) about(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Active = "about"
	app.render(w, "about", http.StatusOK, data)
}

func (app *application) blogView(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ID, err := strconv.Atoi(id)
	if err != nil || ID < 1 {
		app.notFound(w, r)
		return
	}
	blog, err := app.blogs.Get(ID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}

	user, err := app.users.Get(blog.UserID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Blog = blog
	data.User = user
	data.Active = "None"

	app.render(w, "blog", http.StatusOK, data)
}

func (app *application) userLogin(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userLoginForm{}
	data.Active = "login"
	app.render(w, "login", http.StatusOK, data)
}

func (app *application) userLoginPost(w http.ResponseWriter, r *http.Request) {
	var form userLoginForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a valid email address",
	)
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		data.Active = "login"
		app.render(w, "login", http.StatusUnprocessableEntity, data)
		return
	}
	ID, err := app.users.Authenticate(form.Email, form.Password)
	if err != nil {
		if errors.Is(err, models.ErrWrongCredintials) {
			form.AddNonFieldError("Email or Password is incorrect")
			data := app.newTemplateData(r)
			data.Form = form
			data.Active = "login"
			app.render(w, "login", http.StatusUnprocessableEntity, data)
		} else {
			app.serverError(w, err)
		}
		return
	}
	err = app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Put(r.Context(), "authenticatedUserID", (*ID).String())
	app.sessionManager.Put(r.Context(), "flash", "You've logged in succussfully")
	originalUrl := app.sessionManager.PopString(r.Context(), "original-path")
	if originalUrl != "" {
		http.Redirect(w, r, originalUrl, http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userLogoutPost(w http.ResponseWriter, r *http.Request) {
	err := app.sessionManager.RenewToken(r.Context())
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.sessionManager.Remove(r.Context(), "authenticatedUserID")
	app.sessionManager.Put(r.Context(), "flash", "You've been logged out succussfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) userSignup(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = userSignupForm{}
	data.Active = "signup"
	app.render(w, "signup", http.StatusOK, data)
}

func (app *application) userSignupPost(w http.ResponseWriter, r *http.Request) {
	var form userSignupForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Name), "name", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.UserName), "username", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Email), "email", "This field cannot be blank")
	form.CheckField(validator.NotBlank(form.Password), "password", "This field cannot be blank")
	form.CheckField(
		validator.Matches(form.Email, validator.EmailRX),
		"email",
		"This field must be a vlid email address",
	)
	form.CheckField(
		validator.MinChars(form.Password, 8),
		"password",
		"This field must be 8 charachters at least",
	)
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		data.Active = "signup"
		app.render(w, "signup", http.StatusUnprocessableEntity, data)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(form.Password), 12)
	if err != nil {
		app.serverError(w, err)
		return
	}

	user := models.User{
		Name:           form.Name,
		UserName:       form.UserName,
		Email:          form.Email,
		HashedPassword: hashedPassword,
	}
	ID, err := app.users.Insert(user)
	if err != nil {
		data := app.newTemplateData(r)
		if errors.Is(err, models.ErrRepeatedUserName) {
			form.AddNonFieldError("Username already taken")
		} else if errors.Is(err, models.ErrRepeatedEmail) {
			form.AddNonFieldError("A uesr exists with this email")
		} else {
			app.serverError(w, err)
			return
		}
		data.Form = form
		data.Active = "signup"
		app.render(w, "signup", http.StatusUnprocessableEntity, data)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", ID.String())
	app.sessionManager.Put(r.Context(), "flash", "You've Signed Up succussfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) blogCreate(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Active = "create"
	data.Form = blogCreateForm{}
	app.render(w, "create", http.StatusOK, data)
}

func (app *application) blogCreatePost(w http.ResponseWriter, r *http.Request) {
	var form blogCreateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be empty")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be empty")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Active = "create"
		data.Form = form
		app.render(w, "create", http.StatusUnprocessableEntity, data)
		return
	}

	userId := r.Context().Value(isAuthenticatedKey)
	userID, err := uuid.Parse(userId.(string))
	if err != nil {
		app.serverError(w, err)
		return
	}
	id, err := app.blogs.Insert(form.Title, form.Content, userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Blog was created succussfully")
	http.Redirect(w, r, fmt.Sprintf("/blog/%v", id), http.StatusSeeOther)
}

func (app *application) myBlogs(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(isAuthenticatedKey)
	userID, err := uuid.Parse(userId.(string))
	if err != nil {
		app.serverError(w, err)
		return
	}
	user, err := app.users.Get(userID)
	user.Created = time.Time{}
	user.Email = ""
	user.HashedPassword = []byte("")
	if err != nil {
		app.serverError(w, err)
		return
	}
	blogs, err := app.blogs.ByUser(userID)
	if err != nil && !errors.Is(err, models.ErrNoRecord) {
		app.serverError(w, err)
		return
	}
	for _, blog := range blogs {
		if len(blog.Content) > 500 {
			blog.Content = blog.Content[:500] + "..."
		}
	}
	data := app.newTemplateData(r)
	data.Active = "myBlogs"
	data.User = user
	data.Blogs = blogs
	app.render(w, "myBlogs", http.StatusOK, data)
}

func (app *application) userBlogs(w http.ResponseWriter, r *http.Request) {
	username := r.PathValue("username")
	user, err := app.users.GetByUsername(username)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	blogs, err := app.blogs.ByUser(user.ID)
	if err != nil && err != models.ErrNoRecord {
		app.serverError(w, err)
		return
	}
	user.Created = time.Time{}
	user.Email = ""
	user.HashedPassword = []byte("")
	user.ID = uuid.Nil
	users := []*models.User{}
	for _, blog := range blogs {
		if len(blog.Content) > 500 {
			blog.Content = blog.Content[:500] + "..."
		}
		if err != nil {
			app.serverError(w, err)
			return
		}
		users = append(users, user)
	}

	data := app.newTemplateData(r)
	data.User = user
	data.Active = "None"
	data.Blogs = blogs
	data.Users = users
	app.render(w, "userBlogs", http.StatusOK, data)
}

func (app *application) blogUpdate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ID, err := strconv.Atoi(id)
	if err != nil || ID < 1 {
		app.notFound(w, r)
		return
	}
	blog, err := app.blogs.Get(ID)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
			return
		} else {
			app.serverError(w, err)
			return
		}
	}
	userId := r.Context().Value(isAuthenticatedKey)
	userID, err := uuid.Parse(userId.(string))
	if err != nil {
		app.serverError(w, err)
		return
	}
	if blog.UserID != userID {
		app.sessionManager.Put(r.Context(), "flash", "Cannot Update blog")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := app.newTemplateData(r)
	data.Active = "myBlogs"
	data.Blog = blog
	data.Form = blogCreateForm{}
	app.render(w, "update", http.StatusOK, data)
}

func (app *application) blogUpdatePost(w http.ResponseWriter, r *http.Request) {
	var form blogUpdateForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(validator.NotBlank(form.Title), "title", "This field cannot be empty")
	form.CheckField(validator.NotBlank(form.Content), "content", "This field cannot be empty")

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Active = "myBlogs"
		data.Form = form
		data.Blog = &models.Blog{ID: form.BlogID, Title: form.Title, Content: form.Content}
		app.render(w, "create", http.StatusUnprocessableEntity, data)
		return
	}

	blog, err := app.blogs.Get(form.BlogID)
	if err != nil {
		app.notFound(w, r)
		return
	}
	userId := r.Context().Value(isAuthenticatedKey)
	userID, err := uuid.Parse(userId.(string))
	if err != nil {
		app.serverError(w, err)
		return
	}
	if blog.UserID != userID {
		app.sessionManager.Put(r.Context(), "flash", "Cannot Update blog")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if form.Content == blog.Content && form.Title == blog.Title {
		form.AddNonFieldError("Must update either the title or the content")
	}

	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Active = "myBlogs"
		data.Form = form
		data.Blog = &models.Blog{ID: form.BlogID, Title: form.Title, Content: form.Content}
		app.render(w, "update", http.StatusUnprocessableEntity, data)
		return
	}

	err = app.blogs.UpdateBlog(form.BlogID, form.Title, form.Content)
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Blog was updated succussfully")
	http.Redirect(w, r, fmt.Sprintf("/blog/%v", form.BlogID), http.StatusSeeOther)
}

func (app *application) blogDeletePost(w http.ResponseWriter, r *http.Request) {
	var form struct {
		Id int `form:"id"`
	}
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	userId := r.Context().Value(isAuthenticatedKey)
	userID, err := uuid.Parse(userId.(string))
	if err != nil {
		app.serverError(w, err)
		return
	}

	blog, err := app.blogs.Get(form.Id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	if blog.UserID != userID {
		app.sessionManager.Put(r.Context(), "flash", "Failed to delete blog")
		http.Redirect(w, r, "/blogs", http.StatusSeeOther)
		return
	}

	err = app.blogs.DeleteBlog(form.Id)
	if err != nil {
		if errors.Is(err, models.ErrNoRecord) {
			app.notFound(w, r)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Blog deleted succussfully")
	http.Redirect(w, r, "/blogs", http.StatusSeeOther)
}

func (app *application) userAccount(w http.ResponseWriter, r *http.Request) {
	userId := r.Context().Value(isAuthenticatedKey)
	userID, err := uuid.Parse(userId.(string))
	if err != nil {
		app.serverError(w, err)
		return
	}

	numberOfBlogs, err := app.blogs.GetBlogsNumber(userID)
	if err != nil && !errors.Is(err, models.ErrNoRecord) {
		app.serverError(w, err)
		return
	}
	user, err := app.users.Get(userID)
	if err != nil {
		app.serverError(w, err)
		return
	}

	data := app.newTemplateData(r)
	data.Active = "account"
	data.Form = struct{ NumberOfBlogs int }{NumberOfBlogs: numberOfBlogs}
	data.User = user
	app.render(w, "account", http.StatusOK, data)
}

func (app *application) userChangePassword(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Active = "account"
	data.Form = passwordChangeForm{}
	app.render(w, "changePassword", http.StatusOK, data)
}

func (app *application) userChangePasswordPost(w http.ResponseWriter, r *http.Request) {
	var form passwordChangeForm
	err := app.decodePostForm(r, &form)
	if err != nil {
		app.clientError(w, http.StatusBadRequest)
		return
	}

	form.CheckField(
		validator.NotBlank(form.CurrentPassword),
		"currentPassword",
		"This field cannot be blank",
	)
	form.CheckField(
		validator.NotBlank(form.NewPassword),
		"newPassword",
		"This field cannot be blank",
	)
	form.CheckField(
		validator.NotBlank(form.ConfirmNewPassword),
		"confirmNewPassword",
		"This field cannot be blank",
	)
	form.CheckField(
		validator.MinChars(form.NewPassword, 8),
		"newPassword",
		"This field must be 8 charachters at least",
	)
	form.CheckField(
		form.NewPassword == form.ConfirmNewPassword,
		"confirmNewPassword",
		"This field must match the new password",
	)
	if !form.Valid() {
		data := app.newTemplateData(r)
		data.Form = form
		data.Active = "account"
		app.render(w, "changePassword", http.StatusUnprocessableEntity, data)
		return
	}
	if form.CurrentPassword == form.NewPassword {
		form.AddNonFieldError("New password cannot be the old password")
		data := app.newTemplateData(r)
		data.Form = form
		data.Active = "account"
		app.render(w, "changePassword", http.StatusUnprocessableEntity, data)
		return
	}
	userId := r.Context().Value(isAuthenticatedKey)
	userID, err := uuid.Parse(userId.(string))
	if err != nil {
		app.serverError(w, err)
		return
	}
	user, err := app.users.Get(userID)
	if err != nil {
		// Won't reach here in the first place if the usres.Get in the middle requireAuth failed
		app.serverError(w, err)
		return
	}

	err = app.users.ChangePassword(user, form.CurrentPassword, form.NewPassword)
	if err != nil {
		if errors.Is(err, models.ErrWrongCredintials) {
			form.AddNonFieldError("Incorrect Password")
			data := app.newTemplateData(r)
			data.Active = "account"
			data.Form = form
			app.render(w, "changePassword", http.StatusUnprocessableEntity, data)
		} else {
			app.serverError(w, err)
		}
		return
	}

	app.sessionManager.Put(r.Context(), "flash", "Password Changed succussfully")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) search(w http.ResponseWriter, r *http.Request) {
	searchBy := r.URL.Query().Get("search-by")
	searchKey := r.URL.Query().Get("search-key")
	searchUser := r.URL.Query().Get("search-user")
	blogs, err := app.blogs.SearchBy(searchBy, searchKey, searchUser)
	if err != nil && !errors.Is(err, models.ErrNoRecord) {
		app.serverError(w, err)
		return
	}

	users := []*models.User{}
	for _, blog := range blogs {
		if len(blog.Content) > 500 {
			blog.Content = blog.Content[:500] + "..."
		}
		user, err := app.users.Get(blog.UserID)
		user.Created = time.Time{}
		user.Email = ""
		user.HashedPassword = []byte("")
		if err != nil {
			app.serverError(w, err)
			return
		}
		users = append(users, user)
	}

	data := app.newTemplateData(r)
	data.Active = "blogs"
	data.Blogs = blogs
	data.Users = users
	data.Form = struct{ SearchQuery string }{SearchQuery: searchKey}
	app.render(w, "search", http.StatusOK, data)
}
