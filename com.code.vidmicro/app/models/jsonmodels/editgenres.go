package jsonmodels

import "github.com/bytedance/sonic"

type EditGenres struct {
	Id   int    `form:"id" field:"id" validate:"required,gt=0" field:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name" validate:"required,min=3" field:"name"`
}

func (ts *EditGenres) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *EditGenres) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
