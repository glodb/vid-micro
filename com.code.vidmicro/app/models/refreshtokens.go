package models

type RefreshToken struct {
	UserId       int    `db:"user_id INTEGER UNIQUE NOT NULL"`
	RefreshToken string `db:"refresh_token VARCHAR(255)" json:"refresh_token"`
}
