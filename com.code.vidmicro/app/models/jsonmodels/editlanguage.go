package jsonmodels

import "github.com/bytedance/sonic"

type EditLanguage struct {
	Id   int    `json:"id" form:"id" validate:"required,gt=0"`
	Name string `json:"name" form:"name" validate:"required,min=3" field:"name"`
	Code string `json:"code" form:"code" validate:"required,min=3" field:"code"`
}

func (ts *EditLanguage) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *EditLanguage) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
