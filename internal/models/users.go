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
				if strings.Contains(pgError.Message, "users_uc_username") {
					return uuid.New(), ErrRepeatedUserName
				} else if strings.Contains(pgError.Message, "users_uc_email") {
					return uuid.New(), ErrRepeatedEmail
				}
			} else if errors.Is(err, sql.ErrNoRows) {
				return uuid.New(), ErrNoRecord
			}
		}
	}
	return user.ID, err
}

func (model *UsersModel) Get(id uuid.UUID) (*User, error) {
	sqlStatment := `SELECT name, username, email, created, hashed_password FROM users WHERE id=$1;`
	user := User{ID: id}
	err := model.DB.QueryRow(sqlStatment, id.String()).Scan(
		&user.Name,
		&user.UserName,
		&user.Email,
		&user.Created,
		&user.HashedPassword,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNoRecord
	}
	return &user, err
}

func (model *UsersModel) Authenticate(email, password string) (*uuid.UUID, error) {
	sqlStatment := `SELECT id,hashed_password FROM users WHERE email=$1`
	user := User{}
	err := model.DB.QueryRow(sqlStatment, email).Scan(&user.ID, &user.HashedPassword)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrWrongCredintials
		}
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return nil, ErrWrongCredintials
		}
		return nil, err
	}
	return &user.ID, nil
}

func (model *UsersModel) ChangePassword(user *User, currentPassword, newPassword string) error {
	sqlStatment := `UPDATE users SET hashed_password=$1 WHERE id=$2 RETURNING id;`

	if err := bcrypt.CompareHashAndPassword(user.HashedPassword, []byte(currentPassword)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrWrongCredintials
		}
		return err
	}
	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}
	err = model.DB.QueryRow(sqlStatment, string(newHashedPassword), user.ID).Scan(&user.ID)
	return err
}

func (model *UsersModel) GetByUsername(username string) (*User, error) {
	sqlStatment := `SELECT id, name, username, email, hashed_password, created FROM users WHERE username=$1`
	user := User{}
	err := model.DB.QueryRow(sqlStatment, username).Scan(
		&user.ID,
		&user.Name,
		&user.UserName,
		&user.Email,
		&user.HashedPassword,
		&user.Created,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		}
		return nil, err
	}
	return &user, nil
}
