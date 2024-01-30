package jsonmodels

type ResetPassword struct {
	Email string `json:"email" form:"email" validate:"required,email" field:"email"`
}
