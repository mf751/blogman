package models

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type blogs struct {
	Title   string
	Content string
	ID      uuid.UUID
	UserID  uuid.UUID
	Created time.Time
	Updated time.Time
	Views   int
}

type blogsModel struct {
	DB *sql.DB
}

func (model *blogsModel) Insert(title, content string, userID uuid.UUID)
