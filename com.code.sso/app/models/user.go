package models

type User struct {
	Id               int    `db:"id,INT,AUTO_INCREMENT,PRIMARY KEY"`
	Email            string `db:"email,VARCHAR(255),NOT NULL,UNIQUE" json:"email"`
	Phone            string `db:"phone,VARCHAR(255),NOT NULL" json:"phone"`
	Password         string `db:"password,VARCHAR(50),NOT NULL" json:"password,omitempty"`
	FirstName        string `db:"firstName,VARCHAR(50)" json:"firstName"`
	LastName         string `db:"lastName,VARCHAR(50)" json:"lastName"`
	Salt             []byte `db:"salt,VARBINARY(16)"`
	RegistrationType int    `db:"registrationType,INT" json:"registrationType"`
	CreatedAt        int    `db:"createdAt,INT" json:"createdAt"`
	UpdatedAt        int    `db:"updatedAt,INT" json:"updatedAt"`
}
