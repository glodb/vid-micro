package baseconnections

import (
	"database/sql"

	_ "github.com/lib/pq"

	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
)

// Keeping it open MysqlConnectionfor multiple db or own db connections in microservices
type PsqlConnection struct {
	db *sql.DB
}

func (u *PsqlConnection) CreateConnection() (ConntectionInterface, error) {
	dsn := "postgres://" + configmanager.GetInstance().Database.Username + ":" + configmanager.GetInstance().Database.Password + "@" + configmanager.GetInstance().Database.Host + ":" + configmanager.GetInstance().Database.Port + "/" + configmanager.GetInstance().Database.DBName + "?sslmode=disable"

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	u.db = db
	return u, nil
}

func (u *PsqlConnection) GetDB(dbType basetypes.DbType) interface{} {
	return u.db
}
