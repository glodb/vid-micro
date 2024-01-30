package models

import "github.com/bytedance/sonic"

type TitleType struct {
	Id   int    `db:"id SERIAL PRIMARY KEY" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name" validate:"required,min=3" field:"name"`
	Slug string `db:"slug VARCHAR(255) NOT NULL UNIQUE" json:"slug" form:"slug" validate:"required,min=2" field:"slug"`
}

func (ts *TitleType) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *TitleType) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
