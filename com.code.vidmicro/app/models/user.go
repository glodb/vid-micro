package models

import "time"

type User struct {
	Id          int       `db:"id SERIAL PRIMARY KEY" form:"id"`
	Username    string    `db:"username VARCHAR(255) NOT NULL UNIQUE" json:"username" form:"username"`
	Name        string    `db:"name VARCHAR(255)" json:"name" form:"name"`
	Email       string    `db:"email VARCHAR(255) NOT NULL UNIQUE" json:"email" form:"email"`
	Password    string    `db:"password VARCHAR(50) NOT NULL" json:"password,omitempty" form:"password,omitempty"`
	AvatarUrl   string    `db:"avatar_url VARCHAR(255)" json:"avatar_url,omitempty" form:"avatar_url,omitempty"`
	IsVerified  bool      `db:"is_verified BOOLEAN NOT NULL DEFAULT FALSE" json:"is_verified,omitempty" form:"is_verified,omitempty"`
	BlackListed bool      `db:"black_listed BOOLEAN NOT NULL DEFAULT FALSE" json:"black_listed,omitempty"`
	Salt        []byte    `db:"salt BYTEA"`
	Role        int       `db:"role SMALLINT" json:"role" form:"role"`
	CreatedAt   time.Time `db:"createdAt TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt   time.Time `db:"updatedAt TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt   time.Time `db:"deletedAt TIMESTAMPTZ" json:"deletedAt"`
}
