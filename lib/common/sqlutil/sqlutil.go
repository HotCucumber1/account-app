package sqlutil

import (
	"database/sql"
	"errors"
	"strings"

	mysqldriver "github.com/go-sql-driver/mysql"
)

const errDuplicateEntry = 1062

func IsDuplicate(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == errDuplicateEntry
}

func NotFound(err, notFound error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return notFound
	}
	return err
}

func IsDuplicateKey(err error, key string) bool {
	var mysqlErr *mysqldriver.MySQLError
	return errors.As(err, &mysqlErr) && mysqlErr.Number == errDuplicateEntry && strings.Contains(mysqlErr.Message, key)
}

func Placeholders(count int) string {
	return strings.TrimSuffix(strings.Repeat("?, ", count), ", ")
}

func Rows(rows, columns int) string {
	row := "(" + Placeholders(columns) + ")"
	return strings.TrimSuffix(strings.Repeat(row+", ", rows), ", ")
}
