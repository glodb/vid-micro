package basefunctions

import (
	"database/sql"
	"errors"
	"log"
	"reflect"
	"strings"

	"com.code.sso/com.code.sso/database/baseconnections"
	"com.code.sso/com.code.sso/database/basetypes"
)

type MySqlFunctions struct {
}

func (u *MySqlFunctions) GetFunctions() BaseFucntionsInterface {
	return u
}

func (u *MySqlFunctions) EnsureIndex(dbName basetypes.DBName, collectionName basetypes.CollectionName, unique bool) error {
	return nil
}

func (u *MySqlFunctions) Add(dbName basetypes.DBName, collectionName basetypes.CollectionName, data interface{}) error {
	conn := baseconnections.GetInstance().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := "INSERT INTO " + string(collectionName)

	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return errors.New("Required a struct for data")
	}

	var columns []string
	var placeholders []string
	values := make([]interface{}, 0)

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		tag := field.Tag.Get("db")

		if tag == "" {
			continue
		}

		value := dataValue.Field(i).Interface()
		values = append(values, value)

		columns = append(columns, tag)
		placeholders = append(placeholders, "?")
	}

	query += "(" + strings.Join(columns, ", ") + ")"
	query += " VALUES(" + strings.Join(placeholders, ", ") + ")"

	_, err := conn.Exec(query, values...)
	log.Println(query, conn, values)
	return err
}
func (u *MySqlFunctions) FindOne(basetypes.DBName, basetypes.CollectionName, interface{}, interface{}) error {
	log.Println("FindOne MySql")
	return nil
}
func (u *MySqlFunctions) UpdateOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query interface{}, data interface{}, upsert bool) error {
	log.Println("UpdateOne MySql")
	return nil
}
func (u *MySqlFunctions) DeleteOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query interface{}) error {
	log.Println("DeleteOne MySql")
	return nil
}
