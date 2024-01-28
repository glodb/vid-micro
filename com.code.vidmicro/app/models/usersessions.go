package models

type UserSessions struct {
	UserId     int    `db:"user_id INTEGER"`
	SessionId  string `db:"session_id VARCHAR(255)" json:"session_id"`
	CreatedAt  int64  `db:"created_at INTEGER" json:"created_at"`
	ExpiringAt int64  `db:"expiring_at INTEGER" json:"expiring_at"`
}
