package jsonmodels

type LoginEmail struct {
	Email    string `db:"email VARCHAR(255) NOT NULL UNIQUE" json:"email" form:"email" validate:"required,email,validateExclusive" field:"email"`
	Password string `db:"password VARCHAR(50) NOT NULL" json:"password,omitempty" form:"password,omitempty" validate:"min=8,password,required" field:"password"`
}
