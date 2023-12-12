package redis_repo

import (
	"10-typing/common"
	"10-typing/errors"
	"context"

	"github.com/redis/go-redis/v9"
)

const retries = 100

type RedisRepository struct {
	redisClient *redis.Client
}

func NewRedisRepository(redisClient *redis.Client) *RedisRepository {
	return &RedisRepository{redisClient}
}

type RedisPipeline struct {
	pipe redis.Pipeliner
}

func (p *RedisPipeline) Conn() any {
	return p.pipe
}

func (p *RedisPipeline) Commit(ctx context.Context) error {
	const op errors.Op = "redis_repo.RedisPipeline.Commit"

	cmds, err := p.pipe.Exec(ctx)
	if err == nil {
		return nil
	}

	err = nil
	for _, cmd := range cmds {
		if cmd.Err() != nil {
			err = errors.Join(err, cmd.Err())
		}
	}

	return errors.E(op, err)
}

func (p *RedisPipeline) Rollback() error {
	p.pipe.Discard()

	return nil
}

type RedisTransaction struct {
	pipe redis.Pipeliner
}

func (t *RedisTransaction) Conn() any {
	return t.pipe
}

func (t *RedisTransaction) Commit(ctx context.Context) error {
	const op errors.Op = "redis_repo.RedisTransaction.Commit"

	cmds, err := t.pipe.Exec(ctx)
	if err == nil {
		return nil
	}

	err = nil
	for _, cmd := range cmds {
		if cmd.Err() != nil {
			err = errors.Join(err, cmd.Err())
		}
	}

	return errors.E(op, err)
}

func (t *RedisTransaction) Rollback() error {
	t.pipe.Discard()

	return nil
}

func (repo *RedisRepository) BeginPipeline() common.Transaction {
	return &RedisPipeline{pipe: repo.redisClient.Pipeline()}
}

func (repo *RedisRepository) BeginTx() common.Transaction {
	return &RedisTransaction{pipe: repo.redisClient.TxPipeline()}
}

func (repo *RedisRepository) cmdable(tx common.Transaction) redis.Cmdable {
	if tx != nil {
		return tx.Conn().(redis.Pipeliner)
	}

	return repo.redisClient
}

func (repo *RedisRepository) beginPipelineIfNoOuterTransactionExists(outerTx common.Transaction) (cmd redis.Cmdable, innerTx common.Transaction) {
	if outerTx != nil {
		cmd = repo.cmdable(outerTx)
	} else {
		innerTx = repo.BeginPipeline()
		cmd = repo.cmdable(innerTx)
	}

	return cmd, innerTx
}
