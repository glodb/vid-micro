package models

import "github.com/bytedance/sonic"

type Genres struct {
	Id   int    `db:"id SERIAL PRIMARY KEY" form:"id"`
	Name string `db:"name VARCHAR(255) NOT NULL UNIQUE" json:"name" form:"name"`
}

func (ts *Genres) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *Genres) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}

type GenresData struct {
	GenreId int `form:"genre_id" json:"genre_id"`
	TitleId int `form:"title_id" json:"title_id"`
}
