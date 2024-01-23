package models

import "github.com/bytedance/sonic"

type ContentType struct {
	Id   int    `db:"id SERIAL PRIMARY KEY" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name"`
}

func (ts *ContentType) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *ContentType) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
