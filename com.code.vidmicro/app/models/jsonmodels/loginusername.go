package jsonmodels

type LoginUsername struct {
	Username string `db:"username VARCHAR(255) NOT NULL UNIQUE" json:"username" form:"username" validate:"min=3,max=20,required" field:"username"`
	Password string `db:"password VARCHAR(50) NOT NULL" json:"password,omitempty" form:"password,omitempty" validate:"min=8,password,required" field:"password"`
}
