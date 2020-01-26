package mysql_utils

import (
	"fmt"
	"strings"

	"github.com/apiotrowski312/isOnline-utils-go/rest_errors"
	"github.com/go-sql-driver/mysql"
)

const (
	ErrorNoRow = "no rows in result set"
)

func ParseError(err error) rest_errors.RestErr {
	sqlErr, ok := err.(*mysql.MySQLError)

	if !ok {
		if strings.Contains(err.Error(), ErrorNoRow) {
			return rest_errors.NewInternalServerError(fmt.Sprintf("no record matching given id"))
		}
		return rest_errors.NewInternalServerError("error parsing database response")
	}

	fmt.Println(sqlErr.Number)

	switch sqlErr.Number {
	case 1062:
		return rest_errors.NewInternalServerError("Smth with uniqe field is wrong")
	}
	return rest_errors.NewInternalServerError("error processing request")
}
