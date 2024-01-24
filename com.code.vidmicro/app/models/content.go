package models

type Contents struct {
	Id              int    `db:"id SERIAL PRIMARY KEY" json:"id" form:"id"`
	Name            string `db:"name VARCHAR(255) NOT NULL" json:"name" form:"name"`
	AlternativeName string `db:"alternative_name VARCHAR(255)" json:"alternative_name" form:"alternative_name"`
	Thumbnail       string `db:"thumbnail VARCHAR(255)" json:"thumbnail" form:"thumbnail"`
	Description     string `db:"description TEXT" json:"description" form:"description"`
	TypeId          int    `db:"type_id INTEGER" json:"type_id" form:"type_id"`
	TypeName        string `json:"type_name"`
	LanguageId      int    `db:"language_id INTEGER" json:"language_id" form:"language_id"`
	LanguageName    string `json:"language_name"`
	LanguageCode    string `json:"language_code"`
	AssociatedTitle int    `db:"associated_title VARCHAR(255)" json:"associated_title" form:"associated_title"`
}
