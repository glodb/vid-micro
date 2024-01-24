package models

import "github.com/bytedance/sonic"

type LanguageMeta struct {
	Id         string `db:"id VARCHAR(255) PRIMARY KEY UNIQUE NOT NULL"`
	TitlesId   int    `db:"titles_id Integer" form:"titles_id" json:"titles_id"`
	LanguageId int    `db:"language_id Integer" form:"language_id" json:"language_id"`
	StatusId   int    `db:"status_id Integer" form:"status_id" json:"status_id"`
}

type LanguageMetaDetails struct {
	LanguageId   int    `json:"language_id"`
	LanguageName string `json:"language_name"`
	LanguageCode string `json:"language_code"`
	StatusId     int    `json:"status_id"`
	StatusName   string `json:"status_name"`
}

func (ts *LanguageMetaDetails) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *LanguageMetaDetails) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
