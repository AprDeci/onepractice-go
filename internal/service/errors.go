package service

import "errors"

var ErrDatabaseDisabled = errors.New("database disabled: MYSQL_DSN is empty")
