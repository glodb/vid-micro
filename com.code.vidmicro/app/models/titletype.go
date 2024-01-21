package models

type TitleType struct {
	Id   int    `db:"id SERIAL PRIMARY KEY" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name"`
	Slug string `db:"slug VARCHAR(255) NOT NULL UNIQUE" json:"slug" form:"slug"`
}
