package common

type Transaction interface {
	Commit()
	Rollback()
}
