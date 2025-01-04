package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             uuid.UUID
	Name           string
	Email          string
	Created        time.Time
	HashedPassword []byte
}

type usersModel struct {
	DB *sql.DB
}

func (model *usersModel) Insert(user User) (User, error) {
	sqlStatment := `INSERT INTO users (id ,name, email, hashed_password, created) VALUES(
  $1, $2, $3, $4, $5)
  RETURNING id, created;
  )`
	err := model.DB.QueryRow(sqlStatment,
		uuid.New().String(),
		user.Name,
		user.Email,
		user.HashedPassword,
		time.Now().UTC(),
	).Scan(&user.ID, &user.Created)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrNoRecord
	}
	return user, err
}

func (model *usersModel) Get(id uuid.UUID) (User, error) {
	sqlStatment := `SELECT name, email, created FROM users WHERE id=$1;`
	user := User{ID: id}
	err := model.DB.QueryRow(sqlStatment, id.String()).Scan(
		&user.Name,
		&user.Email,
		&user.Created,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrNoRecord
	}
	return user, err
}

func (model *usersModel) Authenticate(user User, password string) error {
	sqlStatment := `SELECT id,hashed_password FROM users WHERE email=$1`
	err := model.DB.QueryRow(sqlStatment, user.Email).Scan(&user.ID, &user.HashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrWrongCredintials
		}
		return err
	}
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrWrongCredintials
		}
		return err
	}
	return nil
}

func (model *usersModel) ChangePassword(user User, newPassword string) (User, error) {
	sqlStatment1 := `SELECT hashed_password From users WHERE id=$1`
	sqlStatment2 := `UPDATE users SET hashed_password=$1 WHERE id=$2 RETURNING name, email, created ;`
	var password string
	err := model.DB.QueryRow(sqlStatment1, user.ID).Scan(&password)
	if err != nil {
		return User{}, err
	}

	if err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return User{}, ErrWrongCredintials
		}
		return User{}, err
	}
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return User{}, err
	}
	user.HashedPassword = newHashedPassword
	err = model.DB.QueryRow(sqlStatment2, string(newHashedPassword), user.ID).Scan(
		&user.Name,
		&user.Email,
		&user.Created,
	)
	if err != nil {
		return User{}, err
	}
	return user, nil
}
