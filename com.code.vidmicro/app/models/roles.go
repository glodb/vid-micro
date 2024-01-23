package models

type Roles struct {
	Id   int    `db:"id PRIMARY KEY"`
	Name string `db:"name VARCHAR(255)" json:"name"`
	Slug string `db:"slug VARCHAR(255) NOT NULL UNIQUE" json:"slug"`
}
