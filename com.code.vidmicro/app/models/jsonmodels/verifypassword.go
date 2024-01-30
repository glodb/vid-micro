package jsonmodels

type VerifyPassword struct {
	PasswordHash string `json:"password_hash" form:"password_hash" validate:"required" field:"password_hash"`
	NewPassword  string `json:"new_password" form:"new_password" validate:"min=8,password,required" field:"new_password"`
}
