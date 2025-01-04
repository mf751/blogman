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

type userModel struct {
	DB *sql.DB
}

func (model *userModel) Insert(user User) (uuid.UUID, error) {
	sqlStatment := `INSERT INTO users (id ,name, email, hashed_password, created) VALUES(
  $1, $2, $3, $4, $5)
  RETURNING id;
  )`
	var id uuid.UUID
	err := model.DB.QueryRow(sqlStatment,
		uuid.New().String(),
		user.Name,
		user.Email,
		user.HashedPassword,
		time.Now().UTC(),
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return uuid.UUID{}, ErrNoRecord
	}
	return id, err
}

func (model *userModel) Get(id uuid.UUID) (User, error) {
	sqlStatment := `SELECT * FROM users WHERE id=$1;`
	user := User{}
	err := model.DB.QueryRow(sqlStatment, id.String()).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.HashedPassword,
		&user.Created,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return user, ErrNoRecord
	}
	return user, err
}

func (model *userModel) Authenticate(user User, password string) error {
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

func (model *userModel) ChangePassword(user User, newPassword string) error {
	sqlStatment1 := `SELECT hashed_password From users WHERE id=$1`
	sqlStatment2 := `UPDATE users SET hashed_password=$1 WHERE id=$2`
	var password string
	err := model.DB.QueryRow(sqlStatment1, user.ID).Scan(&password)
	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrWrongCredintials
		}
		return err
	}
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	var id uuid.UUID
	err = model.DB.QueryRow(sqlStatment2, string(newHashedPassword), user.ID).Scan(&id)
	return err
}
