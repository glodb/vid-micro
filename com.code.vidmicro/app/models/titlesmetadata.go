package models

type TitleMetaData struct {
	Id              int     `db:"id SERIAL PRIMARY KEY" form:"id"`
	TitleId         int     `db:"title_id INTEGER UNIQUE" form:"title_id"`
	Title           string  `db:"title VARCHAR(255) NOT NULL" json:"title" form:"title"`
	AlternativeName string  `db:"alternative_name VARCHAR(255)" json:"alternative_name" form:"alternative_name"`
	Sequence        int     `db:"sequence INTEGER DEFAULT NULL" form:"sequence"`
	TypeId          int     `db:"type_id INTEGER DEFAULT NULL" form:"type_id"`
	Year            int     `db:"year INTEGER" json:"year" form:"year"`
	Score           float64 `db:"score REAL" json:"score" form:"score"`
	Genres          []int   `db:"genres INTEGER[]" json:"genres" form:"genres"`
}
