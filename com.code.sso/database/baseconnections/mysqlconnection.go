package baseconnections

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"com.code.sso/com.code.sso/config"
	"com.code.sso/com.code.sso/database/basetypes"
)

//Keeping it open for multiple db or own db connections in microservices
type MysqlConnection struct {
	dbName string
	db     *sql.DB
}

func (u *MysqlConnection) CreateConnection() (ConntectionInterface, error) {
	dsn := config.GetInstance().Database.Username + ":" + config.GetInstance().Database.Password + "@tcp(" + config.GetInstance().Database.Host + ":" + config.GetInstance().Database.Port + ")/" + config.GetInstance().Database.DBName
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
