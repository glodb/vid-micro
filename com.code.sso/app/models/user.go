package models

type User struct {
	Email            string `db:"email" json:"email"`
	Password         string `db:"password" json:"password"`
	FirstName        string `db:"firstName" json:"firstName"`
	LastName         string `db:"lastName" json:"lastName"`
	Salt             string `db:"salt"`
	RegistrationType int    `db:"registrationType"`
	CreatedAt        string `db:"createdAt" json:"createdAt"`
	UpdatedAt        string `db:"updatedAt" json:"updatedAt"`
}
