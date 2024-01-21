package models

type TitlesLanguage struct {
	LanguageId string `json:"lang_id"`
	StatusId   int    `json:"status_id"`
}

type Titles struct {
	Id            int      `db:"id SERIAL PRIMARY KEY"`
	OriginalTitle string   `db:"original_title VARCHAR(255) NOT NULL" json:"original_title" form:"original_title"`
	Year          int      `db:"year INTEGER" json:"year" form:"year"`
	CoverUrl      string   `db:"cover_url VARCHAR(255)" json:"cover_url" form:"cover_url"`
	Languages     string   `json:"languages" form:"languages"`
	LanguagesMeta []string `db:"languages_meta VARCHAR(50)[]"`
}
