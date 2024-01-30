package jsonmodels

type EditTitles struct {
	Id            int    `db:"id SERIAL PRIMARY KEY" form:"id" validate:"required,gt=0" field:"id"`
	OriginalTitle string `db:"original_title VARCHAR(255) NOT NULL" json:"original_title" form:"original_title" validate:"required,min=3" field:"original_title"`
	Year          int    `db:"year INTEGER" json:"year" form:"year" validate:"required,gt=1900" field:"year"`
	CoverUrl      string `db:"cover_url VARCHAR(255)" json:"cover_url" form:"cover_url"`
}
