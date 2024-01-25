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

func (u *MySqlFunctions) AddMany(dbName basetypes.DBName, collectionName basetypes.CollectionName, data []interface{}, scan bool) ([]int64, error) {
	return nil, errors.New("unimplemented exception")
}

func (u *MySqlFunctions) Paginate(dbName basetypes.DBName, collectionName basetypes.CollectionName, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool, pageSize int, page int) (*sql.Rows, int64, error) {
	return nil, -1, errors.New("unimplemented exception")
}

func (u *MySqlFunctions) Count(dbName basetypes.DBName, collectionName basetypes.CollectionName, condition map[string]interface{}) (int64, error) {
	return -1, nil
}

func (u *MySqlFunctions) Add(dbName basetypes.DBName, collectionName basetypes.CollectionName, data interface{}, scan bool) (int64, error) {
	conn := baseconnections.GetInstance().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := "INSERT INTO " + string(collectionName)

	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return -1, errors.New("Required a struct for data")
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
	return 0, err
}
func (u *MySqlFunctions) Find(dbName basetypes.DBName, collectionName basetypes.CollectionName, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool) (*sql.Rows, error) {
	conn := baseconnections.GetInstance().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)

	query := "SELECT * FROM " + string(collectionName)

	whereClause := ""
	values := make([]interface{}, 0)

	for key, val := range condition {
		if whereClause != "" {
			if !useOr {
				whereClause += " AND "
			} else {
				whereClause += " OR "
			}
		} else {
			whereClause += " WHERE "
		}
		whereClause += key + "= ? "
		values = append(values, val)
	}

	if addParenthesis {
		whereClause = "(" + whereClause + ")"
	}

	query += whereClause + appendQuery
	rows, err := conn.Query(query, values...)

	return rows, err
}
func (u *MySqlFunctions) UpdateOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query string, data []interface{}, upsert bool) error {
	conn := baseconnections.GetInstance().GetConnection(basetypes.MYSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	_, err := conn.Exec(query, data...)
	return err
}

func (u *MySqlFunctions) RawQuery(dbName basetypes.DBName, collectionName basetypes.CollectionName, query string, data []interface{}, upsert bool) error {
	return nil
}
func (u *MySqlFunctions) DeleteOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, condition map[string]interface{}, useOr bool, addParenthesis bool) error {
	log.Println("Unimplemented DeleteOne MySql")
	return nil
}
