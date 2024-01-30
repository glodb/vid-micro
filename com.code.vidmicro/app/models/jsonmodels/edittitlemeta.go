package jsonmodels

type EditTitleMetaData struct {
	TitleId         int     `form:"title_id" validate:"required,gt=0" field:"title_id"`
	AlternativeName string  `json:"alternative_name" form:"alternative_name" validate:"omitempty,min=3" field:"alternative_name"`
	Sequence        int     `form:"sequence" validate:"omitempty,oneof=1" field:"sequence"`
	TypeId          int     `form:"type_id" validate:"required,gt=0" field:"type_id"`
	Score           float64 `form:"score" validate:"omitempty,gt=0,lte=10" field:"score"`
}
