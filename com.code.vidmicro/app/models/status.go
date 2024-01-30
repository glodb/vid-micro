package models

import (
	"github.com/bytedance/sonic"
)

type Status struct {
	Id   int    `db:"id SERIAL PRIMARY KEY" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name" validate:"required,min=3" field:"name"`
}

func (ts *Status) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *Status) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
