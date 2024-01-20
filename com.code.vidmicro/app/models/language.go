package models

type Language struct {
	Id   string `db:"id VARCHAR(255) NOT NULL UNIQUE" json:"id" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name"`
	Code string `db:"code VARCHAR(255) NOT NULL UNIQUE" json:"code" form:"code"`
}
