package baseconnections

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"

	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"com.code.vidmicro/com.code.vidmicro/settings/configmanager"
)

// Keeping it open MysqlConnectionfor multiple db or own db connections in microservices
type PsqlConnection struct {
	db *sql.DB
}

func (u *PsqlConnection) CreateConnection() (ConntectionInterface, error) {
	dsn := "postgres://" + configmanager.GetInstance().Database.Username + ":" + configmanager.GetInstance().Database.Password + "@" + configmanager.GetInstance().Database.Host + ":" + configmanager.GetInstance().Database.Port + "?sslmode=disable"

	db, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	defer db.Close()

	dbName := configmanager.GetInstance().Database.DBName
	checkDBQuery := fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname='%s'", dbName)
	var exists int
	err = db.QueryRow(checkDBQuery).Scan(&exists)

	if err != nil && err != sql.ErrNoRows {
		log.Println(err)
	}

	// If the database doesn't exist, create it
	if exists != 1 {
		createDBQuery := fmt.Sprintf("CREATE DATABASE %s", dbName)
		_, err := db.Exec(createDBQuery)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Database %s created successfully\n", dbName)
	} else {
		log.Printf("Database %s already exists\n", dbName)
	}

	dsn = "postgres://" + configmanager.GetInstance().Database.Username + ":" + configmanager.GetInstance().Database.Password + "@" + configmanager.GetInstance().Database.Host + ":" + configmanager.GetInstance().Database.Port + "/" + configmanager.GetInstance().Database.DBName + "?sslmode=disable"

	newdb, err := sql.Open("postgres", dsn)

	if err != nil {
		return nil, err
	}

	u.db = newdb
	return u, nil
}

func (u *PsqlConnection) GetDB(dbType basetypes.DbType) interface{} {
	return u.db
}
