package sql_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"context"

	"gorm.io/gorm"
)

type SQLTransaction struct {
	tx *gorm.DB
}

func (t *SQLTransaction) Commit(ctx context.Context) error {
	const op errors.Op = "sql_repo.SQLTransaction.Commit"

	if err := t.tx.WithContext(ctx).Commit().Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (t *SQLTransaction) Rollback() error {
	const op errors.Op = "sql_repo.SQLTransaction.Rollback"

	if err := t.tx.Rollback().Error; err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (t *SQLTransaction) Conn() any {
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
		return tx.Conn().(*gorm.DB)
	}

	return sr.db
}
