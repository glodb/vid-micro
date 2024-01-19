package basefunctions

import (
	"database/sql"
	"errors"
	"log"
	"reflect"
	"strings"

	"com.code.vidmicro/com.code.vidmicro/database/baseconnections"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
)

type MySqlFunctions struct {
}

func (u *MySqlFunctions) GetFunctions() BaseFucntionsInterface {
	return u
}

func (u *MySqlFunctions) EnsureIndex(dbName basetypes.DBName, collectionName basetypes.CollectionName, data interface{}) error {
	conn := baseconnections.GetInstance().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := `CREATE TABLE IF NOT EXISTS ` + string(collectionName) + ` (`
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return errors.New("Required a struct for data")
	}

	columns := ""

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		tags := strings.Split(field.Tag.Get("db"), ",")

		if columns != "" {
			columns += ","
		}

		columns += strings.Join(tags, " ")
	}

	query += columns + ");"
	_, err := conn.Exec(query)
	return err
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
		tag := strings.Split(field.Tag.Get("db"), ",")[0]

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
	return err
}
func (u *MySqlFunctions) FindOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, condition map[string]interface{}, result interface{}) (*sql.Rows, error) {
	conn := baseconnections.GetInstance().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := "SELECT * FROM " + string(collectionName)

	whereClause := ""
	values := make([]interface{}, 0)

	for key, val := range condition {
		if whereClause != "" {
			whereClause += " AND "
		} else {
			whereClause += " WHERE "
		}
		whereClause += key + "= ? "
		values = append(values, val)
	}

	query += whereClause
	rows, err := conn.Query(query, values...)

	return rows, err
}
func (u *MySqlFunctions) UpdateOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query string, data []interface{}, upsert bool) error {
	conn := baseconnections.GetInstance().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	_, err := conn.Exec(query, data...)
	return err
}
func (u *MySqlFunctions) DeleteOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query interface{}) error {
	log.Println("Unimplemented DeleteOne MySql")
	return nil
}
