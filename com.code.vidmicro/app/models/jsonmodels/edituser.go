package jsonmodels

import "database/sql"

type EditUser struct {
	Name      string         `json:"name" form:"name" validate:"omitempty,min=3" field:"name"`
	AvatarUrl sql.NullString `json:"avatar_url,omitempty" form:"avatar_url,omitempty"`
}
