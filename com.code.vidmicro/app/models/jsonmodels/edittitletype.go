package jsonmodels

import "github.com/bytedance/sonic"

type EditTitleType struct {
	Id   int    `form:"id" validate:"required,gt=0" field:"id"`
	Name string `json:"name" form:"name" validate:"required,min=3" field:"name"`
	Slug string `json:"slug" form:"slug" validate:"required,min=2" field:"slug"`
}

func (ts *EditTitleType) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *EditTitleType) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
