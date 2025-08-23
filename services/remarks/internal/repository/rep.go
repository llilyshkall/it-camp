package repository

import (
	"database/sql"
)

type StoreInterface interface {
}

type SQLStore struct {
	db *sql.DB
}

func NewSQLStore(db *sql.DB) StoreInterface {
	return &SQLStore{
		db: db,
	}
}
