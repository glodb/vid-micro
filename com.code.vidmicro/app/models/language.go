package models

import (
	"github.com/bytedance/sonic"
)

type Language struct {
	Id   int    `db:"id SERIAL PRIMARY KEY" json:"id" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name"`
	Code string `db:"code VARCHAR(255) NOT NULL UNIQUE" json:"code" form:"code"`
}

func (ts *Language) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *Language) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
