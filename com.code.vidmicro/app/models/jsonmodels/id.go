package jsonmodels

type Id struct {
	Id int `form:"id" json:"id" validate:"gt=0,required" field:"id"`
}

type IdEmpty struct {
	Id int `form:"id" json:"id" validate:"omitempty,gt=0" field:"id"`
}

type TitleId struct {
	Id int `form:"title_id" json:"title_id" validate:"gt=0,required" field:"title_id"`
}

type TitleIdEmpty struct {
	Id int `form:"title_id" json:"title_id" validate:"omitempty,gt=0" field:"title_id"`
}

type AssociatedTitle struct {
	Id int `form:"associated_title" json:"associated_title" validate:"gt=0,required" field:"associated_title"`
}

type AssociatedTitleEmpty struct {
	Id int `form:"associated_title" json:"associated_title" validate:"omitempty,gt=0" field:"associated_title"`
}
