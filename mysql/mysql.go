package aphro_mysql

import (
	"fmt"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
)

type Mysql struct {
	lastRecordID uint64
	lastError error
}

type QueryResult struct {

}

type InsertResult struct {

}

type DeleteResult struct {

}

func (am *mysql) Query (Result,error){

}

func (am *mysql) QueryRow (Result,) {

}