package models

import "github.com/bytedance/sonic"

type MeilisearchTitle struct {
	Id               int                   `json:"id"`
	OriginalTitle    string                `json:"original_title"`
	Year             int                   `json:"year"`
	CoverUrl         string                `json:"cover_url"`
	LanguagesDetails []LanguageMetaDetails `json:"languages_details,omitempty"`
	AlternativeName  string                `json:"alternnative_name"`
	Sequence         int                   `json:"sequence"`
	TypeId           int                   `json:"type_id"`
	TypeName         string                `json:"type_name"`
	Score            float64               `json:"score"`
	Genres           []string              `json:"genres"`
	GenresObject     map[int]string        `json:"genres_ids"`
}

func (ts *MeilisearchTitle) EncodeRedisData() []byte {
	buf, _ := sonic.Marshal(ts)
	return buf
}

func (ts *MeilisearchTitle) DecodeRedisData(data []byte) {
	sonic.Unmarshal(data, &ts)
}
