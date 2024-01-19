package baseconnections

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
)

//Keeping it open for multiple db or own db connections in microservices
type MysqlConnection struct {
	db *sql.DB
}

func (u *MysqlConnection) CreateConnection() (ConntectionInterface, error) {
	dsn := configmanager.GetInstance().Database.Username + ":" + configmanager.GetInstance().Database.Password + "@tcp(" + configmanager.GetInstance().Database.Host + ":" + configmanager.GetInstance().Database.Port + ")/" + configmanager.GetInstance().Database.DBName
	db, err := sql.Open("mysql", dsn)

	if err != nil {
		return nil, err
	}

	u.db = db
	return u, nil
}

func (u *MysqlConnection) GetDB(dbType basetypes.DbType) interface{} {
	return u.db
}
