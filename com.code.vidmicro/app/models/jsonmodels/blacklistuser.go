package jsonmodels

type BlackListUser struct {
	Id int `json:"id" form:"id" validate:"required,id" field:"blackListUserId"`
}
