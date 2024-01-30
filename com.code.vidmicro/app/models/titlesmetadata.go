package models

type TitleMetaData struct {
	Id              int     `db:"id SERIAL PRIMARY KEY" form:"id"`
	TitleId         int     `db:"title_id INTEGER UNIQUE" form:"title_id" validate:"required,gt=0" field:"title_id"`
	Title           string  `db:"title VARCHAR(255) NOT NULL" json:"title" form:"title"`
	AlternativeName string  `db:"alternative_name VARCHAR(255)" json:"alternative_name" form:"alternative_name" validate:"omitempty,min=3" field:"alternative_name"`
	Sequence        int32   `db:"sequence INTEGER DEFAULT NULL" form:"sequence" validate:"omitempty,oneof=1" field:"sequence"`
	TypeId          int     `db:"type_id INTEGER DEFAULT NULL" form:"type_id" validate:"required,gt=0" field:"type_id"`
	Year            int     `db:"year INTEGER" json:"year" form:"year"`
	Score           float64 `db:"score REAL" json:"score" form:"score" validate:"omitempty,gt=0,lte=10" field:"score"`
	Genres          []int   `db:"genres INTEGER[]" json:"genres" form:"genres" validate:"gt=0,dive,required" field:"genres"`
}
