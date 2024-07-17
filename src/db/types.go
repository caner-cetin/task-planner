package db

import (
	"database/sql"

	"gorm.io/gorm"
)

var Pool *sql.DB
var DB *gorm.DB
