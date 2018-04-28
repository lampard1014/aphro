package aphro_mysql

import (
	"fmt"
	"errors"
	"database/sql"
    _ "github.com/go-sql-driver/mysql"
)

/////// interface //////




type AphroPersistentStoreResult interface {}

///////////////////////////////////////////////
//field的结构体，实现接口IAphroPersistentStoreField
///////////////////////////////////////////////
type IAphroPersistentStoreField interface {
	func FetchFieldName() (string,error)
	func FetchAlias() (string,error)
}
type APSField struct {
	filedName string
	alias string
}

func (apsField *APSField) FetchFieldName (string, error) {
	var returnErr error = nil
	if fetchFieldName == nil {
		returnErr = errors.New("没有列名")
	}
	return apsField.filedName,returnErr
}

func (apsField *APSField) FetchAlias (string, error) {
	var returnErr error = nil
	if alias == nil {
		returnErr = errors.New("没有列的别名")
	}
	return apsField.alias,returnErr
}
///////////////////////////////////////////////
//entity的结构体，实现接口IAphroPersistentStoreEntity
///////////////////////////////////////////////
type IAphroPersistentStoreEntity interface {
	FetchEntityName()(string,error)
	FetchEntityAlais()(string,error)
}

type APSEntity struct {
	entityName string
	alias string
}

func (apsEntity *APSEntity) FetchEntityName (string, error) {
	var returnErr error = nil
	if entityName == nil {
		returnErr = errors.New("没有实体名")
	}
	return apsEntity.entityName,returnErr
}

func (apsEntity *APSEntity) FetchEntityAlais (string, error) {
	var returnErr error = nil
	if alias == nil {
		returnErr = errors.New("没有实体名的别名")
	}
	return apsEntity.alias,returnErr
}
////////////////////////////////////////
type IAphroPersistentStore interface {
	Query()(AphroPersistentStoreResult,error)
	Insert()
	Update()
	Delete()
}

type APSEntityJoin int 
const (
	_ APSEntityJoin = iota
	InnerJoin
	LeftJoin
	RightJoin
)

type APSMysql struct {
	lastRecordID uint64
	lastError error

	columns [] *APSField 	//nil == * 
	entities [] *APSEntity 	
	entitiesJoin []APSEntityJoin

	bindValues []string
	wheres []string 
	wheresOperator []string

	perpareStatement string 

}

//////////////////////////////////////////////////////////////////////////////////////
func (apsMysql *APSMysql)Select(columnsAs [string]string) (apsMysql *APSMysql,error) {
	var returnErr error = nil
	var returnInstance *APSMysql = nil
	if columnsAs == nil {
		apsMysql.columns == nil
	} else if len(columnsAs) > 0 {
		for k,v := range columnsAs {
			c := new(APSField)
			c.filedName = k
			if v != nil {
				c.alias = v
			}
		}
		returnInstance = apsMysql
	} else {
		returnErr = errors.New("SELECT column can't be empty")
	}
	return returnInstance,returnErr
}

//////////////////////////////////////////////////////////////////////////////////////
func (apsMysql *APSMysql)From(entity string,entityAlais string) (apsMysql *APSMysql,error) {
	var returnErr error = nil
	var returnInstance *APSMysql = nil

	if entity == nil {
		returnErr = errors.New("empty FROM table entity ")
	} else {
		e := new(APSEntity)
		e.entityName = entity
		if entityAlais != nil {
			e.alias = entityAlais
		}
		apsMysql.entities = append(apsMysql.entities,e)
		returnInstance = apsMysql

	}
	return returnInstance,returnErr
}

//////////////////////////////////////////////////////////////////////////////////////
func (apsMysql *APSMysql)InnerJoin(entity string,entityAlais string) (apsMysql *APSMysql,error) {
	var returnErr error = nil
	var returnInstance *APSMysql = nil

	if entity == nil {
		returnErr = errors.New("empty InnerJoin table entity ")
	} else {
		e := new(APSEntity)
		e.entityName = entity
		if entityAlais != nil {
			e.alias = entityAlais
		}
		apsMysql.entities = append(apsMysql.entities,e)
		apsMysql.entitiesJoin = append(apsMysql.entitiesJoin,InnerJoin)
		returnInstance = apsMysql
	}
	return returnInstance,returnErr
}

//////////////////////////////////////////////////////////////////////////////////////
func (apsMysql *APSMysql)LeftJoin(entity string,entityAlais string) (apsMysql *APSMysql,error) {
	var returnErr error = nil
	var returnInstance *APSMysql = nil

	if entity == nil {
		returnErr = errors.New("empty LeftJoin table entity ")
	} else {
		e := new(APSEntity)
		e.entityName = entity
		if entityAlais != nil {
			e.alias = entityAlais
		}
		apsMysql.entities = append(apsMysql.entities,e)
		apsMysql.entitiesJoin = append(apsMysql.entitiesJoin,LeftJoin)
		returnInstance = apsMysql
	}
	return returnInstance,returnErr
}

//////////////////////////////////////////////////////////////////////////////////////
func (apsMysql *APSMysql)RightJoin(entity string,entityAlais string) (apsMysql *APSMysql,error) {
	var returnErr error = nil
	var returnInstance *APSMysql = nil

	if entity == nil {
		returnErr = errors.New("empty RightJoin table entity ")
	} else {
		e := new(APSEntity)
		e.entityName = entity
		if entityAlais != nil {
			e.alias = entityAlais
		}
		apsMysql.entities = append(apsMysql.entities,e)
		apsMysql.entitiesJoin = append(apsMysql.entitiesJoin,RightJoin)
		returnInstance = apsMysql
	}
	return returnInstance,returnErr
}


////////////////////////////////whereCondition////////////////////////////////////////

//a == b OR c == d
//SQL:`id=1 AND id=2`
//$cond = ['and', 'id=1', 'id=2']

//SQL:`type=1 AND (id=1 OR id=2)`
//$cond = ['and', 'type=1', ['or', 'id=1', 'id=2']]

//SQL:`type=1 AND (id=1 OR id=2)` //此写法'='可以换成其他操作符，例：in like != >=等
// $cond = [
//     'and',
//     ['=', 'type', 1],
//     [
//         'or',
//         ['=', 'id', '1'],
//         ['=', 'id', '2'],
//     ]
//]

type APSMysqlOperator string 
const (
	APSMysqlOperator_OR = "OR"
	APSMysqlOperator_AND = "AND"
	APSMysqlOperator_Equal = "="
	APSMysqlOperator_Above = ">"
	APSMysqlOperator_Below = "<"
	APSMysqlOperator_NotEqual1 = "!="
	APSMysqlOperator_NotEqual2 = "<>"
	APSMysqlOperator_AboveEqual = ">="
	APSMysqlOperator_BelowEqual = "<="
	APSMysqlOperator_Not = "NOT"
	APSMysqlOperator_In = "IN"
	APSMysqlOperator_Between = "BETWEEN"
	APSMysqlOperator_Like = "LIKE"

)

type APSMysqlExpressionOperator string 
const (
	APSMysqlExpressionOperator_Equal = "="
	APSMysqlExpressionOperator_Above = ">"
	APSMysqlExpressionOperator_Below = "<"
)


type APSMysqlCondition struct {
	operator ConditionOperator 
	operand1 []interface{}
	operand2 []interface{}
}

//&aphro_mysql.APSMysqlCondition{operator:"AND",operand1:[],operand2:[]}

func (apsMysql *APSMysql)Where(condition APSMysqlCondition,bindValues []) (apsMysql *APSMysql,error) {

	if condition == nil {
		append(wheres,"1")
	} else {

	}
}

func (apsMysql *APSMysql)Update() (apsMysql *APSMysql,error) {
}

func (apsMysql *APSMysql)Insert() (apsMysql *APSMysql,error) {
}

const (
	port  = ":10102"
    mysqlDSN = "root:@tcp(127.0.0.1:3306)/iris_db"
)

func (apsMysql *APSMysql) Open() (apsMysql *APSMysql, error) {

	var returnErr error = nil
	var returnInstance *APSMysql = nil


	db, dbOpenErr := sql.Open("mysql", mysqlDSN)
    defer db.Close()
    // Open doesn't open a connection. Validate DSN data:
    dbOpenErr = db.Ping()
    if (dbOpenErr == nil) {
    	returnInstance = apsMysql
    } else {
        returnErr = dbOpenErr        
    }
    return returnInstance,returnErr
}

func (am *mysql) Query() (Result,error) {

}

func (am *mysql) QueryRow() (Result,error) {

}

func (am *mysql) Update() (Result, error) {

}

func (am *mysql) Delete() (Result, error) {

}

func (am *mysql) Insert() (Result, error) {

}