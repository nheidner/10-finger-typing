package models

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/uuid"
)

// rooms:[roomId]:current_game:user_ids
// rooms:[roomId]:current_game text_id status
func (gs *GameService) SetNewCurrentGame(ctx context.Context, textId, roomId uuid.UUID, userIds ...uuid.UUID) error {
	if len(userIds) == 0 {
		return fmt.Errorf("at least one user id must be specified")
	}

	currentGameKey := getCurrentGameKey(roomId)
	statusStr := strconv.Itoa(int(UnstartedGameStatus))
	currentGameValue := map[string]string{
		"text_id": textId.String(),
		"status":  statusStr,
	}
	if err := gs.RDB.HSet(ctx, currentGameKey, currentGameValue).Err(); err != nil {
		return err
	}

	currentGameUseridsKey := getCurrentGameUserIdsKey(roomId)
	userIdStrs := make([]interface{}, 0, len(userIds))
	for _, userId := range userIds {
		userIdStrs = append(userIdStrs, userId.String())
	}
	if err := gs.RDB.SAdd(ctx, currentGameUseridsKey, userIdStrs...).Err(); err != nil {
		return err
	}

	return nil
}
