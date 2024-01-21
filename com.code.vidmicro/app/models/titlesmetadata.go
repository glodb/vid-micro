package models

type TitleMetaData struct {
	Id              int    `db:"id SERIAL PRIMARY KEY" form:"id"`
	TitleId         int    `db:"title_id INTEGER" form:"title_id"`
	Title           string `db:"title VARCHAR(255) NOT NULL" json:"title" form:"title"`
	AlternativeName string `db:"alternnative_name VARCHAR(255)" json:"alternnative_name" form:"alternnative_name"`
	Sequence        int    `db:"sequence INTEGER DEFAULT 0" form:"sequence"`
	TypeId          int    `db:"type_id INTEGER DEFAULT NOT NULL" form:"type_id"`
	Year            int    `db:"year INTEGER" json:"year" form:"year"`
	Score           int    `db:"score REAL" json:"score" form:"score"`
	Genres          []int  `db:"genres INTEGER[]" json:"genres" form:"genres"`
}
