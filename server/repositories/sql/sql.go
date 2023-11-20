package sql_repo

import (
	"10-typing/repositories"

	"gorm.io/gorm"
)

type SQLTransaction struct {
	tx *gorm.DB
}

func (t *SQLTransaction) Commit() {
	t.tx.Commit()
}

func (t *SQLTransaction) Rollback() {
	t.tx.Rollback()
}

type SQLRepository struct {
	db *gorm.DB
}

func NewSQLRepository(db *gorm.DB) *SQLRepository {
	return &SQLRepository{db}
}

func (sr *SQLRepository) BeginTx() repositories.Transaction {
	tx := sr.db.Begin()
	return &SQLTransaction{tx}
}
