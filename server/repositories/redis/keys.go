package redis_repo

import "github.com/google/uuid"

// -----ROOM ----

const (
	roomAdminIdField         = "admin_id"
	roomCreatedAtField       = "created_at"
	roomUpdatedAtField       = "updated_at"
	roomGameDurationSecField = "game_duration"
)

// getRoomKey returns a redis key: rooms:[room_id]
//
// The keys holds a HASH value with the following fields: admin_id, created_at, updated_at, game_duration
func getRoomKey(roomId uuid.UUID) string {
	return "rooms:" + roomId.String()
}

// ---- GAME ----

const (
	currentGameStatusField = "status"
	currentGameIdField     = "game_id"
	currentGameTextIdField = "text_id"
)

// getCurrentGameKey returns a redis key: rooms:[room_id]:current_game
//
// The key holds a HASH value with the following fields: game_id, text_id, status
func getCurrentGameKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":current_game"
}

// getCurrentGameUserIdsKey returns a redis key: rooms:[room_id]:current_game:user_ids
//
// The key holds a SET value of user ids.
// When a user id is in the set, the user id is part of the current game.
// The user ids must be a subset of the user ids held in the key that is returned from getRoomSubscriberIdsKey().
func getCurrentGameUserIdsKey(roomId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":user_ids"
}

// getCurrentGameScoreKey returns a redis key: rooms:[room_id]:current_game:scores:user_ids
//
// The key holds a SORTED SET value: score:wpm, member:userId.
// The value holds references through the user ids to the scores.
func getCurrentGameScoresUserIdsKey(roomId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":scores:user_ids"
}

// getCurrentGameScoreKey returns a redis key: rooms:[room_id]:current_game:scores:[user_id].
//
// The key holds the STRINGIFIED JSON representation of a models.Score value.
func getCurrentGameScoreKey(roomId, userId uuid.UUID) string {
	return getCurrentGameKey(roomId) + ":scores:" + userId.String()
}

// ---- ROOM SUBSCRIBER ----

// getRoomSubscriberIdsKey returns a redis key: rooms:[room_id]:subscribers_ids.
//
// The key hold a SET value of user ids.
// The user ids can be used to look up all the room subscribers with the rooms:[room_id]:subscribers:[user_id] key.
func getRoomSubscriberIdsKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers_ids"
}

const (
	roomSubscriberStatusField     = "status"
	roomSubscriberGameStatusField = "game_status"
	roomSubscriberUsernameField   = "username"
)

// getRoomSubscriberKey returns a redis key: rooms:[room_id]:subscribers:[user_id]
//
// The key holds a HASH value with the fields: status (inactive/active), username, game_status (started, finished, unstarted etc.).
func getRoomSubscriberKey(roomId, userId uuid.UUID) string {
	return getRoomKey(roomId) + ":subscribers:" + userId.String()
}

// getRoomSubscriberConnectionKey returns a redis key: rooms:[room_id]:subscribers:[user_id]:conns
//
// The key holds a SORTED SET value: score:expiration time of connection, member:connection uuid
//
// Every time the status of a room subscriber is queried (active/inactive), the expired connections are deleted.
func getRoomSubscriberConnectionKey(roomId, userId uuid.UUID) string {
	return getRoomSubscriberKey(roomId, userId) + ":conns"
}

// ---- USER ----

const (
	userUsernameField     = "username"
	userPasswordHashField = "password_hash"
	userFirstNameField    = "first_name"
	userLastNameField     = "last_name"
	userEmailField        = "email"
	userIsVerifiedField   = "is_verified"
)

// getUserKey returns a redis key: users:[userid]
//
// The key holds a HASH value: username, password_hash, first_name, last_name, email, is_verified
func getUserKey(userId uuid.UUID) string {
	return "users:" + userId.String()
}

func getUserEmailKey(email string) string {
	return "user_emails:" + email
}

// ---- USER NOTIFICATIONS ----

// getRoomSubscriberConnectionKey returns a redis key: users:[userid]:notifications
//
// The key holds a STREAM value: message:stringified JSON representation of models.UserNotification
func getUserNotificationStreamKey(userId uuid.UUID) string {
	return getUserKey(userId) + ":notifications"
}

// ---- TEXT ----

const (
	// text_ids SET: text id UUIDs
	textIdsKey = "text_ids"
)

// ---- SESSION ----

// getSessionKey returns a redis key: users:[tokenhash]
//
// The key holds a STRING value: user id
func getSessionKey(tokenHash string) string {
	return "sessions:" + tokenHash
}

// ---- ROOM STREAM ----

const (
	streamEntryTypeField    = "type"
	streamEntryMessageField = "message"
	streamEntryActionField  = "action"
)

// getSessionKey returns a redis key: rooms:[room_id]:stream
//
// The key holds a STREAM value: type: message/action; message?: stringified JSON representation of models.PushMessage; action: stringified JSON representation of models.StreamActionType
func getRoomStreamKey(roomId uuid.UUID) string {
	return getRoomKey(roomId) + ":stream"
}
