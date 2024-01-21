package models

type LanguageMeta struct {
	Id         string `db:"id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL"`
	TitlesId   int    `db:"titles_id Integer"`
	LanguageId string `db:"language_id VARCHAR(255)" form:"languageId"`
	StatusId   int    `db:"status_id Integer" form:"statusId"`
}
