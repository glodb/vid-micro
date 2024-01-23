package models

import "github.com/bytedance/sonic"

type LanguageContent struct {
	Id   int    `db:"id INTEGER PRIMARY KEY" json:"id" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name"`
	Code string `db:"code VARCHAR(255) NOT NULL UNIQUE" json:"code" form:"code"`
}

func (ts *LanguageContent) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *LanguageContent) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
