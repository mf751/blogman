package models

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Blog struct {
	ID      int
	Title   string
	Content string
	UserID  uuid.UUID
	Created time.Time
	Updated time.Time
	Views   int
}

type blogsModel struct {
	DB *sql.DB
}

func (model *blogsModel) Insert(title, content string, userID uuid.UUID) (int, error) {
	sqlStatement := `INSERT INTO blogs(title, content, user_id, created, updated, views) Values ($1, $2, $3, NOW(), NOW(), 0) RETURNING id;`
	var id int
	err := model.DB.QueryRow(sqlStatement, title, content, userID).Scan(&id)
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (model *blogsModel) Get(id int) (Blog, error) {
	sqlStatement1 := `SELECT * FROM blogs WHERE id=$1`
	sqlStatement2 := `UPDATE blogs SET views = views + 1 WHERE id = $1 RETURNING views`
	blog := Blog{}
	err := model.DB.QueryRow(sqlStatement1, id).Scan(
		&blog.ID,
		&blog.Title,
		&blog.Content,
		&blog.UserID,
		&blog.Created,
		&blog.Updated,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Blog{}, ErrNoRecord
		}
		return Blog{}, err
	}
	err = model.DB.QueryRow(sqlStatement2, id).Scan(&blog.Views)
	if err != nil {
		return Blog{}, err
	}
	return blog, nil
}

func (model *blogsModel) Latest() ([]*Blog, error) {
	sqlStatement := `SELECT * FROM blogs ORDER BY id DESC LIMIT 10`
	rows, err := model.DB.Query(sqlStatement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blogs := []*Blog{}
	for rows.Next() {
		blog := &Blog{}
		err := rows.Scan(
			&blog.ID,
			&blog.Title,
			&blog.Content,
			&blog.UserID,
			&blog.Created,
			&blog.Updated,
			&blog.Views,
		)
		if err != nil {
			return nil, err
		}
		blogs = append(blogs, blog)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return blogs, nil
}

func (model *blogsModel) Update(id int, content string) error {
	sqlStatement := `UPDATE blogs SET content=$1, update=NOW() WHERE id=$2 RETURNING id`
	err := model.DB.QueryRow(sqlStatement, content, id).Scan(&id)
	return err
}
