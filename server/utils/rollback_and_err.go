package utils

import (
	"10-typing/common"
	"10-typing/errors"
)

func RollbackAndErr(op errors.Op, err error, tx common.Transaction) error {
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		return errors.E(op, rollbackErr)
	}

	return err
}
