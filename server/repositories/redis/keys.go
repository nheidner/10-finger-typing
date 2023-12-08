package redis_repo

import "github.com/google/uuid"

const (
	currentGameStatusField = "status"
	currentGameIdField     = "game_id"
	currentGameTextIdField = "text_id"
)

// getCurrentGameKey returns a redis key of the following form: rooms:[room_id]:current_game
//
// The key holds a HASH value with the following fields: game_id, text_id, status
func getCurrentGameKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":current_game"
}

// getCurrentGameUserIdsKey returns a redis key of the following form: rooms:[room_id]:current_game:user_ids
//
// The key holds a SET value of user ids.
// When a user id is in the set, the user id is part of the current game.
// The user ids must be a subset of the user ids held in the key that is returned from getRoomSubscriberIdsKey().
func getCurrentGameUserIdsKey(roomId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":user_ids"
}

// getCurrentGameScoreKey returns a redis key of the following form: rooms:[room_id]:current_game:scores:user_ids
//
// The key holds a SORTED SET value: score:wpm, member:userId.
// The value holds references through the user ids to the scores.
func getCurrentGameScoresUserIdsKey(roomId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":scores:user_ids"
}

// getCurrentGameScoreKey returns a redis key of the following form: rooms:[room_id]:current_game:scores:[user_id]
//
// The key holds the STRINGIFIED JSON representation of a models.Score value
func getCurrentGameScoreKey(roomId, userId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":scores:" + userId.String()
}
