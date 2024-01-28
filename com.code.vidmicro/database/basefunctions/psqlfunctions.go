package basefunctions

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"com.code.vidmicro/com.code.vidmicro/database/baseconnections"
	"com.code.vidmicro/com.code.vidmicro/database/basetypes"
	"github.com/lib/pq"
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
		tagValue := field.Tag.Get("db")

		if tagValue == "" {
			continue
		}
		tags := strings.Split(tagValue, ",")

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
		tagVal := field.Tag.Get("db")
		if strings.Contains(strings.ToUpper(tagVal), "SERIAL") {
			continue
		}
		tag := strings.Split(tagVal, " ")[0]

		if tag == "" {
			continue
		}

		value := dataValue.Field(i).Interface()

		if strings.Contains(tagVal, "[]") {
			values = append(values, pq.Array(value))

		} else {
			values = append(values, value)
		}

		columns = append(columns, tag)
		placeholders = append(placeholders, "$"+strconv.FormatInt(int64(placeholderCount), 10))
		placeholderCount++
	}

	query += "(" + strings.Join(columns, ", ") + ")"
	query += " VALUES(" + strings.Join(placeholders, ", ") + ")"
	if scan {
		query += " RETURNING id"
	}

	var insertedID int64
	row := conn.QueryRow(
		query,
		values...,
	)
	if row.Err() != nil {
		return -1, row.Err()
	}
	var err error
	if scan {
		err = row.Scan(&insertedID)
	}
	if err != nil {
		return -1, err
	}
	return insertedID, err
}

func (u *PSqlFunctions) AddMany(dbName basetypes.DBName, collectionName basetypes.CollectionName, dataArray []interface{}, scan bool) ([]int64, error) {
	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	query := "INSERT INTO " + string(collectionName)

	var columns []string
	values := make([]interface{}, 0)
	placeholderCount := 1

	for i, data := range dataArray {
		dataValue := reflect.ValueOf(data)
		dataType := dataValue.Type()

		if dataType.Kind() != reflect.Struct {
			return nil, errors.New("required a struct for data")
		}

		var placeholders []string

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
		if i == 0 {
			query += "(" + strings.Join(columns, ", ") + ")"
			query += " VALUES(" + strings.Join(placeholders, ", ") + ")"
		} else {
			query += ", (" + strings.Join(placeholders, ", ") + ")"
		}
	}

	query += " RETURNING id"

	var insertedID []int64
	row := conn.QueryRow(
		query,
		values...,
	)
	if row.Err() != nil {
		return nil, row.Err()
	}
	var err error
	if scan {
		err = row.Scan(&insertedID)
	}
	if err != nil {
		return nil, err
	}
	return insertedID, nil
	// return insertedID, err
}

func (u *PSqlFunctions) Find(dbName basetypes.DBName, collectionName basetypes.CollectionName, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool) (*sql.Rows, error) {
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

func (u *PSqlFunctions) Paginate(dbName basetypes.DBName, collectionName basetypes.CollectionName, keys string, condition map[string]interface{}, result interface{}, useOr bool, appendQuery string, addParenthesis bool, pageSize int, page int) (*sql.Rows, int64, error) {
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

	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", string(collectionName)) + whereClause

	var count int64
	row, err := conn.Query(countQuery, values...)
	if err != nil {
		return nil, -1, err
	}

	countQueryRows := 0
	for row.Next() {
		countQueryRows++
		err = row.Scan(&count)
		if err != nil {
			return nil, -1, err
		}
	}

	query += whereClause + appendQuery

	query = fmt.Sprintf(query+" LIMIT %d", pageSize)

	// If there is a skip offset specified, append the OFFSET clause to the query.
	skip := (page - 1) * pageSize
	if skip > 0 {
		query = fmt.Sprintf(query+" OFFSET %d", skip)
	}

	rows, err := conn.Query(query, values...)

	return rows, count, err
}

func (u *PSqlFunctions) Count(dbName basetypes.DBName, collectionName basetypes.CollectionName, condition map[string]interface{}) (int64, error) {
	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	whereClause := ""
	values := make([]interface{}, 0)

	placeholderCount := 1

	for key, val := range condition {
		if whereClause != "" {
			whereClause += " AND "
		} else {
			whereClause += " WHERE "
		}
		whereClause += key + "= $" + strconv.FormatInt(int64(placeholderCount), 10) + " "
		placeholderCount++
		values = append(values, val)
	}
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", string(collectionName)) + whereClause

	var count int64
	row, err := conn.Query(countQuery, values...)
	if err != nil {
		return -1, err
	}

	countQueryRows := 0
	for row.Next() {
		countQueryRows++
		err = row.Scan(&count)
		if err != nil {
			return -1, err
		}
	}
	return count, nil
}

func (u *PSqlFunctions) UpdateOne(dbName basetypes.DBName, collectionName basetypes.CollectionName, query string, data []interface{}, upsert bool) error {
	if !strings.Contains(query, "UPDATE") {
		return errors.New("format of query seems in correct")
	}
	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	_, err := conn.Exec(query, data...)
	return err
}

func (u *PSqlFunctions) RawQuery(dbName basetypes.DBName, collectionName basetypes.CollectionName, query string, data []interface{}) (*sql.Rows, error) {
	conn := baseconnections.GetInstance().GetConnection(basetypes.PSQL).GetDB(basetypes.PSQL).(*sql.DB)
	rows, err := conn.Query(query, data...)
	return rows, err
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
