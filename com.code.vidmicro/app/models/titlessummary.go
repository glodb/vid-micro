package models

type TitlesSummary struct {
	Id            int    `db:"id INTEGER PRIMARY KEY NOT NULL"`
	OriginalTitle string `db:"original_title VARCHAR(255) NOT NULL" json:"original_title" form:"original_title"`
	Languages     []int  `db:"languages_meta INTEGER[]" json:"languages,omitempty"`
}
