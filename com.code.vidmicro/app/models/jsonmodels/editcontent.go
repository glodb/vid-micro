package jsonmodels

type EditContents struct {
	Id              int    `json:"id" form:"id"  validate:"required,gt=0" field:"id"`
	Name            string `json:"name" form:"name" validate:"omitempty,min=3" field:"name"`
	AlternativeName string `json:"alternative_name" form:"alternative_name" validate:"omitempty,min=3" field:"alternative_name"`
	Thumbnail       string `json:"thumbnail" form:"thumbnail" validate:"omitempty,min=3" field:"thumbnail"`
	Description     string `json:"description" form:"description" validate:"omitempty,min=3" field:"description"`
	TypeId          int    `json:"type_id" form:"type_id" validate:"omitempty,gt=0" field:"type_id"`
	LanguageId      int    `json:"language_id" form:"language_id" validate:"omitempty,gt=0" field:"language_id"`
	AssociatedTitle int    `json:"associated_title" form:"associated_title" validate:"required,gt=0" field:"associated_title"`
}
