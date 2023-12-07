package sql_repo

import (
	"10-typing/common"

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

func (t *SQLTransaction) Db() any {
	return t.tx
}

type SQLRepository struct {
	db *gorm.DB
}

func NewSQLRepository(db *gorm.DB) *SQLRepository {
	return &SQLRepository{db}
}

func (sr *SQLRepository) BeginTx() common.Transaction {
	tx := sr.db.Begin()
	return &SQLTransaction{tx}
}

func (sr *SQLRepository) dbConn(tx common.Transaction) *gorm.DB {
	if tx != nil {
		return tx.Db().(*gorm.DB)
	}

	return sr.db
}
