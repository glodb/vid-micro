package models

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
	Score            int                   `json:"score"`
	Genres           []string              `json:"genres"`
}
