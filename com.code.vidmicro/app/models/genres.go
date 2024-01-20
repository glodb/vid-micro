package models

type Genres struct {
	Id   int    `db:"id SERIAL PRIMARY KEY" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name"`
}
