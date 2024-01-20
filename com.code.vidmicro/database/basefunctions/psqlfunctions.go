package basefunctions

import (
	"database/sql"
	"errors"
	"reflect"
	"strconv"
	"strings"

	"com.code.vidmicro/com.code.vidmicro/database/baseconnections"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
)

type PSqlFunctions struct {
}

func (u *PSqlFunctions) GetFunctions() BaseFucntionsInterface {
	return u
}

func (u *PSqlFunctions) EnsureIndex(dbName basetypes.DBName, collectionName basetypes.CollectionName, data interface{}) error {
	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := `CREATE TABLE IF NOT EXISTS ` + string(collectionName) + ` (`
	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return errors.New("required a struct for data")
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

func (u *PSqlFunctions) Add(dbName basetypes.DBName, collectionName basetypes.CollectionName, data interface{}, scan bool) (int64, error) {
	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	query := "INSERT INTO " + string(collectionName)

	dataValue := reflect.ValueOf(data)
	dataType := dataValue.Type()

	if dataType.Kind() != reflect.Struct {
		return -1, errors.New("required a struct for data")
	}

	var columns []string
	var placeholders []string
	values := make([]interface{}, 0)

	placeholderCount := 1

	for i := 0; i < dataType.NumField(); i++ {
		field := dataType.Field(i)
		if strings.Contains(strings.ToUpper(field.Tag.Get("db")), "SERIAL") {
			continue
		}
		tag := strings.Split(field.Tag.Get("db"), " ")[0]

		if tag == "" {
			continue
		}

		value := dataValue.Field(i).Interface()
		values = append(values, value)

		columns = append(columns, tag)
		placeholders = append(placeholders, "$"+strconv.FormatInt(int64(placeholderCount), 10))
		placeholderCount++
	}

	query += "(" + strings.Join(columns, ", ") + ")"
	query += " VALUES(" + strings.Join(placeholders, ", ") + ") RETURNING id"

	var insertedID int64
	row := conn.QueryRow(
		query,
		values...,
	)
	var err error
	if scan {
		err = row.Scan(&insertedID)
	}
	if err != nil {
		return -1, err
	}
	return insertedID, err
}

func (u *PSqlFunctions) FindOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool) (*sql.Rows, error) {
	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	query := "SELECT * FROM " + string(collectionName)

	if keys != "" {
		query = "SELECT " + keys + " FROM " + string(collectionName)
	}

	whereClause := ""
	values := make([]interface{}, 0)

	placeholderCount := 1

	for key, val := range condition {
		if whereClause != "" {
			if !useOr {
				whereClause += " AND "
			} else {
				whereClause += " OR "
			}
		} else {
			whereClause += " WHERE "
			if addParenthesis {
				whereClause = whereClause + "("
			}
		}
		whereClause += key + "= $" + strconv.FormatInt(int64(placeholderCount), 10) + " "
		placeholderCount++
		values = append(values, val)
	}
	if addParenthesis {
		whereClause = whereClause + ")"
	}

	query += whereClause + appendQuery
	rows, err := conn.Query(query, values...)

	return rows, err
}

func (u *PSqlFunctions) UpdateOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query string, data []interface{}, upsert bool) error {
	if !strings.Contains(query, "UPDATE") {
		return errors.New("format of query seems in correct")
	}
	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	_, err := conn.Exec(query, data...)
	return err
}

func (u *PSqlFunctions) DeleteOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, condition map[string]interface{}, useOr bool, addParenthesis bool) error {
	if len(condition) == 0 {
		errors.New("delete can't run with out conditions")
	}

	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.MYSQL).(*sql.DB)
	query := "DELETE FROM " + string(collectionName)

	whereClause := ""
	values := make([]interface{}, 0)

	placeholderCount := 1

	for key, val := range condition {
		if whereClause != "" {
			if !useOr {
				whereClause += " AND "
			} else {
				whereClause += " OR "
			}
		} else {
			whereClause += " WHERE "
			if addParenthesis {
				whereClause = whereClause + "("
			}
		}
		whereClause += key + "= $" + strconv.FormatInt(int64(placeholderCount), 10) + " "
		placeholderCount++
		values = append(values, val)
	}
	if addParenthesis {
		whereClause = whereClause + ")"
	}

	query += whereClause
	_, err := conn.Exec(query, values...)
	return err
}
