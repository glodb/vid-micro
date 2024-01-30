package jsonmodels

import "github.com/bytedance/sonic"

type EditStatus struct {
	Id   int    `form:"id" field:"id" validate:"required,gt=0"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name" validate:"required,min=3" field:"name"`
}

func (ts *EditStatus) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *EditStatus) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
