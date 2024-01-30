package jsonmodels

type Login struct {
	Username string `json:"username" form:"username" validate:"required_without=Email,omitempty,min=3,max=20" field:"username"`
	Email    string `json:"email" form:"email" validate:"required_without=Username,omitempty,email" field:"email"`
	Password string `json:"password,omitempty" form:"password,omitempty" validate:"min=8,password,required" field:"password"`
}
