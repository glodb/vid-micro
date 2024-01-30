package jsonmodels

import "github.com/bytedance/sonic"

type EditContentType struct {
	Id   int    `form:"id" validate:"required,gt=0"`
	Name string `json:"name" form:"name" validate:"required,min=3" field:"name"`
}

func (ts *EditContentType) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *EditContentType) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
