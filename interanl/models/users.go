package models

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             uuid.UUID
	Name           string
	UserName       string
	Email          string
	Created        time.Time
	HashedPassword []byte
}

type UsersModel struct {
	DB *sql.DB
}

func (model *UsersModel) Insert(user User) (uuid.UUID, error) {
	sqlStatment := `INSERT INTO users (id ,name,username, email, hashed_password, created) VALUES(
  $1, $2, $3, $4, $5, $6)
  RETURNING id;
 `
	err := model.DB.QueryRow(sqlStatment,
		uuid.New().String(),
		user.Name,
		user.UserName,
		user.Email,
		user.HashedPassword,
		time.Now().UTC(),
	).Scan(&user.ID)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == "23505" {
				if strings.Contains(pgError.Message, "users_uc_email") {
					return uuid.New(), ErrRepeatedEmail
				} else if strings.Contains(pgError.Message, "users_uc_username") {
					return uuid.New(), ErrRepeatedUserName
				}
			} else if errors.Is(err, sql.ErrNoRows) {
				return uuid.New(), ErrNoRecord
			}
		}
	}
	return user.ID, err
}

func (model *UsersModel) Get(id uuid.UUID) (*User, error) {
	sqlStatment := `SELECT name, username, email, created FROM users WHERE id=$1;`
	user := User{ID: id}
	err := model.DB.QueryRow(sqlStatment, id.String()).Scan(
		&user.Name,
		&user.UserName,
		&user.Email,
		&user.Created,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRecord
	}
	return &user, err
}

func (model *UsersModel) Authenticate(user User, password string) error {
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

func (model *UsersModel) ChangePassword(user User, newPassword string) error {
	sqlStatment1 := `SELECT hashed_password From users WHERE id=$1`
	sqlStatment2 := `UPDATE users SET hashed_password=$1 WHERE id=$2 RETURNING id;`
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
	err = model.DB.QueryRow(sqlStatment2, string(newHashedPassword), user.ID).Scan(&user.ID)
	return err
}
