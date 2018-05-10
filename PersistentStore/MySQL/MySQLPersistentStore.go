package MySQL

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"strconv"
	"strings"
	"github.com/lampard1014/aphro/PersistentStore"
)

const (
	KConfigKey_DriverName = "ConfigKey_DriverName"
	KConfigKey_DSN = "ConfigKey_DSN"

	vConfigKey_DriverName = "mysql"
	vConfigKey_DSN = "root:@tcp(127.0.0.1:3306)/iris_db"
)

//SQL : select a AS a1,b AS b1 from table1 AS  where c=3 and d =4 order by f limit 5
/// mysql := NewAPSMySQL(nil)
//mysql->Select(map[string]string{a:"a1"})->From("table1")->Where()->Limit(1)->OrderBy("a desc")->Execute()

//  &APSMySQLCondition{APSMySQLOperator_AND, &APSMySQLCondition{APSMySQLOperator_EQUAL, "a" ,3} ,&APSMySQLCondition{APSMySQLOperator_EQUAL, "b" ,4}}

//select a From table1

///////////////////////////////////////////////
//field的结构体，实现接口IAphroPersistentStoreField
///////////////////////////////////////////////
type APSField struct {
	filedName string
	alias string
}

func (apsField *APSField) FetchFieldName() (string, error) {
	var returnErr error = nil
	if apsField.filedName == "" {
		returnErr = PersistentStore.NewPSErrC(PersistentStore.NoFieldSpecify)
	}
	return apsField.filedName,returnErr
}

func (apsField *APSField) FetchAlias() (string, error) {
	var returnErr error = nil
	if apsField.alias == "" {
		returnErr = PersistentStore.NewPSErrC(PersistentStore.NoFieldAliasSpecify)
	}
	return apsField.alias,returnErr
}

type APSEntity struct {
	entityName string
	alias string
}

func (this *APSEntity) FetchEntityName() (string, error) {
	var returnErr error = nil
	if this.entityName == "" {
		returnErr = PersistentStore.NewPSErrC(PersistentStore.NoEntitySpecify)
	}
	return this.entityName,returnErr
}

func (this *APSEntity) FetchEntityAlais() (string, error) {
	var returnErr error = nil
	if this.alias == "" {
		returnErr = PersistentStore.NewPSErrC(PersistentStore.NoEntityAliasSpecify)
	}
	return this.alias,returnErr
}

type APSMySQLClientConfiguration struct {
	config map[string]interface{}
}

func (this *APSMySQLClientConfiguration)SetOptions(nc map[string]interface{}){
	this.config = nc
}

func (this *APSMySQLClientConfiguration)GetOptions()(map[string]interface{}) {
	return this.config
}

func NewAPSMySQLClientConfiguration(config map[string]interface{}) *APSMySQLClientConfiguration {
	this := &APSMySQLClientConfiguration{}
	this.SetOptions(config)
	return this
}

type APSMySQLClient struct{
	config *APSMySQLClientConfiguration
	mysqlClient *sql.DB
}


func (this APSMySQLClient)FetchClient()(interface{}) {
	return this.mysqlClient
}

func (this APSMySQLClient)FetchConfiguration()(PersistentStore.IAphroPersistentStoreClientConfiguration) {
	 return this.config
}

func (this APSMySQLClient)SetConfiguration(c PersistentStore.IAphroPersistentStoreClientConfiguration)(error) {
	var returnErr error = nil
	d,b := c.(*APSMySQLClientConfiguration)
	if !b {
		this.config = d
	} else {
		returnErr = PersistentStore.NewPSErrC(PersistentStore.ConfigurationErr)
	}
	return returnErr
}

func NewAPSMySQLClient(config *APSMySQLClientConfiguration) (*APSMySQLClient,error) {
	var returnErr error = nil
	c := &APSMySQLClient{}
	returnErr = c.SetConfiguration(config)
	return c,returnErr
}

type APSMySQLToken int

const (
	_ APSMySQLToken = iota
	APSMySQLToken_SELECT
	APSMySQLToken_INSERT
	APSMySQLToken_UPDATE
	APSMySQLToken_DELETE
)

var APSMySQLTokenMap = map[APSMySQLToken]string{
	APSMySQLToken_SELECT:"SELECT",
	APSMySQLToken_INSERT:"INSERT INTO",
	APSMySQLToken_UPDATE:"UPDATE",
	APSMySQLToken_DELETE:"DELETE",
}


type APSMySQLEntityJoin int
const (
	_ APSMySQLEntityJoin = iota
	APSMySQLEntityJoin_INNERJOIN
	APSMySQLEntityJoin_LEFTJOIN
	APSMySQLEntityJoin_RIGHTJOIN
)

var APSMySQLEntityJoinMap = map[APSMySQLEntityJoin]string{
	APSMySQLEntityJoin_INNERJOIN:"INNER JOIN",
	APSMySQLEntityJoin_LEFTJOIN:"LEFT JOIN",
	APSMySQLEntityJoin_RIGHTJOIN:"RIGHT JOIN",
}


type APSMySQLResult struct {

	lastRecordID uint64
	lastError error
	rowAffected int64

}

func (this *APSMySQLResult)LastInsertId() (uint64, error) {
	return this.lastRecordID,this.lastError
}

func (this *APSMySQLResult)RowsAffected() (int64, error) {
	return this.rowAffected,this.lastError
}

func (this *APSMySQLResult)FetchRow(dest ...interface{})(error) {
	return this.lastError
}

func (this *APSMySQLResult)FetchAll(dest ...interface{})(error) {
	return this.lastError
}


type APSMySQL struct {

	client *APSMySQLClient

	lastError error

	result *APSMySQLResult

	columns []*APSField
	entities []*APSEntity
	entitiesJoin []APSMySQLEntityJoin

	insertValues[][]interface{}
	updateColumns map[string]interface{}

	token APSMySQLToken

	bindValues []interface{}
	wheres string

	limit string
	orderBy string

	prepareStatement string
}

func NewAPSMySQL(userConfig map[string]string ) (*APSMySQL,error) {
	var apsMysql *APSMySQL = &APSMySQL{}
	var drivername string = vConfigKey_DriverName
	var DSN string = vConfigKey_DSN
	var ok bool = false
	if drivername,ok = userConfig[KConfigKey_DriverName]; ok {
		drivername = userConfig[KConfigKey_DriverName]
	}
	if DSN,ok = userConfig[KConfigKey_DSN]; ok {
		DSN = userConfig[KConfigKey_DriverName]
	}
	c := map[string]interface{} {
		KConfigKey_DriverName:drivername,
		KConfigKey_DSN:DSN,
	}

	apsConfig := NewAPSMySQLClientConfiguration(c)
	client, err := NewAPSMySQLClient(apsConfig)
	apsMysql.client = client
	return apsMysql, err
}




func (this *APSMySQL) Connect() (PersistentStore.IAphroPersistentStore) {
	c := this.client.config.GetOptions()

	db, dbOpenErr := sql.Open(c[KConfigKey_DriverName].(string), c[KConfigKey_DSN].(string))
	dbOpenErr = db.Ping()
	if (dbOpenErr == nil) {
		this.client.mysqlClient = db
	} else {
		this.lastError = dbOpenErr
	}
	this.Reset()
	return this
}

func (this *APSMySQL) Close() (PersistentStore.IAphroPersistentStore) {
	defer this.client.mysqlClient.Close()
	return this
}

func (this *APSMySQL) Reset()(PersistentStore.IAphroPersistentStore) {
	this.result = nil
	this.lastError = nil
	this.columns = []*APSField{}
	this.entities = []*APSEntity{}
	this.entitiesJoin = []APSMySQLEntityJoin{}

	this.insertValues = [][]interface{}{}
	this.updateColumns = map[string]interface{}{}
	this.token = 0

	this.limit = ""
	this.orderBy = ""

	this.bindValues = []interface{}{}
	this.prepareStatement = ""

	return this
}


func (this *APSMySQL) Query(querySQL string, bindsValue []interface{})(PersistentStore.IAphroSQLPersistentStore) {
	this.Reset()
	this.prepareStatement = querySQL
	this.bindValues = bindsValue
	return this
}

func (this *APSMySQL) Select(columnsAs map[string]string)(PersistentStore.IAphroSQLPersistentStore) {
	this.Reset()
	this.token = APSMySQLToken_SELECT
	if columnsAs == nil {
		this.columns = nil
	} else if len(columnsAs) > 0 {
		for k,v := range columnsAs {
			c := new(APSField)
			c.filedName = k
			if v != "" {
				c.alias = v
			}
			this.columns = append(this.columns,c)
		}
	}
	return this
}

func (this *APSMySQL) From(entity string, entityAlias string)(PersistentStore.IAphroSQLPersistentStore) {
	var returnErr error = nil
	if entity == "" {
		returnErr = PersistentStore.NewPSErrC(PersistentStore.NoEntitySpecify)
	} else {
		e := new(APSEntity)
		e.entityName = entity
		if entityAlias != "" {
			e.alias = entityAlias
		}
		this.entities = append(this.entities,e)
	}
	this.lastError = returnErr
	return this
}

func (this *APSMySQL)Insert(entity string,columns []string,values [][]interface{})(PersistentStore.IAphroSQLPersistentStore) {
	this.Reset()
	this.token = APSMySQLToken_INSERT
	//entity
	e := new(APSEntity)
	e.entityName = entity
	this.entities = append(this.entities,e)
	//columns
	for _,f := range columns {
		c := new(APSField)
		c.filedName = f
		this.columns = append(this.columns,c)
	}
	//values
	this.insertValues = values

	return this
}


func (this *APSMySQL)Update(entity string,columnValues map[string]interface{})(PersistentStore.IAphroSQLPersistentStore) {
	this.Reset()
	this.token = APSMySQLToken_UPDATE
	//entity
	e := new(APSEntity)
	e.entityName = entity
	this.entities = append(this.entities,e)
	//columnValue
	this.updateColumns = columnValues
	return this
}

func (this *APSMySQL)Delete()(PersistentStore.IAphroSQLPersistentStore) {
	this.Reset()
	this.token = APSMySQLToken_DELETE
	return this
}

//////////////////////////////////////////////////////////////////////////////////////
func (this *APSMySQL)InnerJoin(entity string,entityAlais string) (PersistentStore.IAphroSQLPersistentStore) {
	if entity == "" {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.NoEntitySpecify)
	} else {
		e := new(APSEntity)
		e.entityName = entity
		if entityAlais != "" {
			e.alias = entityAlais
		}
		this.entities = append(this.entities,e)
		this.entitiesJoin = append(this.entitiesJoin,APSMySQLEntityJoin_INNERJOIN)
	}
	return this
}

//////////////////////////////////////////////////////////////////////////////////////
func (this *APSMySQL)LeftJoin(entity string,entityAlais string) (PersistentStore.IAphroSQLPersistentStore) {

	if entity == "" {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.NoEntitySpecify)
	} else {
		e := new(APSEntity)
		e.entityName = entity
		if entityAlais != "" {
			e.alias = entityAlais
		}
		this.entities = append(this.entities,e)
		this.entitiesJoin = append(this.entitiesJoin,APSMySQLEntityJoin_LEFTJOIN)
	}
	return this
}

//////////////////////////////////////////////////////////////////////////////////////
func (this *APSMySQL)RightJoin(entity string,entityAlais string) (PersistentStore.IAphroSQLPersistentStore) {
	if entity == "" {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.NoEntitySpecify)
	} else {
		e := new(APSEntity)
		e.entityName = entity
		if entityAlais != "" {
			e.alias = entityAlais
		}
		this.entities = append(this.entities,e)
		this.entitiesJoin = append(this.entitiesJoin,APSMySQLEntityJoin_RIGHTJOIN)
	}
	return this
}

func (this *APSMySQL)Execute(bindsValue []interface{})(PersistentStore.IAphroSQLPersistentStore){
	//form to query

	var queryStatment string = ""
	this.bindValues = bindsValue

	//query Token
	queryToken := APSMySQLTokenMap[this.token]
	queryStatment += queryToken

	//query Fields
	var queryFields string = "*"
	if this.columns != nil {

		queryFeildsFormer := []string{}
		for _,cv := range this.columns {
			var fieldToken  string = ""
			fieldName := cv.filedName
			fieldToken += fieldName
			if cv.alias != "" {
				fieldToken += " AS " + cv.alias
			}
			queryFeildsFormer = append(queryFeildsFormer,fieldToken)
		}
		queryFields = strings.Join(queryFeildsFormer,",")
	}

	queryStatment += " " + queryFields + " "

	// From  from a as a innerjoin b as b
	var fromToken string = "FROM"
	if this.entities != nil {
		entitiesFormer := []string{}
		for ei,ev := range this.entities {
			var entityToken  string = ""
			entityName := ev.entityName
			entityToken += entityName
			if ev.alias != "" {
				entityToken += " AS " + ev.alias
			}

			if this.entitiesJoin != nil {
				entityToken += " " + APSMySQLEntityJoinMap[this.entitiesJoin[ei/2]] + " "
			}

			entitiesFormer = append(entitiesFormer,entityToken)
		}
		fromToken += " " + strings.Join(entitiesFormer," ")
	} else {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.NoEntitySpecify)
	}

	queryStatment += " " + fromToken + " "
	// where
	queryStatment += " " + this.wheres + " "
	//order by
	queryStatment += " " + this.orderBy + " "
	//limit
	queryStatment += " " + this.limit + " "
	// do query
	this.prepareStatement = queryStatment
	this.bindValues = bindsValue
	//this.Query(queryStatment,bindsValue)

	return this
}

func (this *APSMySQL) query()(*APSMySQL) {
	if this.token ==  APSMySQLToken_SELECT {
		//this.client.mysqlClient.QueryRow()
	}
	return this
}


func (this *APSMySQL) Limit(limitSQL string)(PersistentStore.IAphroSQLPersistentStore) {
	this.limit = limitSQL
	return this
}

func (this *APSMySQL) OrderBy(orderBySQL string)(PersistentStore.IAphroSQLPersistentStore) {
	this.orderBy = orderBySQL
	return this
}

////////////////////////////////whereCondition////////////////////////////////////////

//SQL a=2 =>(=,a,2)

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

type APSMySQLOperator string
const (
	APSMySQLOperator_NONE = ""
	APSMySQLOperator_OR = "OR"
	APSMySQLOperator_AND = "AND"
	APSMySQLOperator_Equal = "="
	APSMySQLOperator_Above = ">"
	APSMySQLOperator_Below = "<"
	APSMySQLOperator_NotEqual1 = "!="
	APSMySQLOperator_NotEqual2 = "<>"
	APSMySQLOperator_AboveEqual = ">="
	APSMySQLOperator_BelowEqual = "<="
	APSMySQLOperator_Not = "NOT"
	APSMySQLOperator_In = "IN"
	APSMySQLOperator_Between = "BETWEEN"
	APSMySQLOperator_Like = "LIKE"
)

type APSMySQLCondition struct {
	operator APSMySQLOperator
	operand1 interface{}
	operand2 interface{}
}

//&aphro_mysql.APSMysqlCondition{operator:"AND",operand1:[],operand2:[]}

func (this *APSMySQL)Where(condition interface{}) (PersistentStore.IAphroSQLPersistentStore) {

	c,ok := condition.(*APSMySQLCondition)
	if ok {
		this.wheres = this.parseWhereCondition(c)
	} else {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.WhereConditionParseErr)
	}
	return this
}

func (this *APSMySQL)parseWhereCondition(condition *APSMySQLCondition) string {
	//todo need  complete all where conditions
	var conditionClause string = "1"
	if condition != nil {
		//提取 operator
		operator := condition.operator
		operand1 := condition.operand1
		operand2 := condition.operand1

		var parseO1 string = ""
		var parseO2 string = ""

		switch  o1:= operand1.(type) {
		case *APSMySQLCondition:
			parseO1 = this.parseWhereCondition(o1)
		case string:
			parseO1 = o1
		default:
			this.lastError = PersistentStore.NewPSErrC(PersistentStore.WhereConditionParseErr)
		}

		switch  o2:= operand2.(type) {
		case *APSMySQLCondition:
			parseO2 = this.parseWhereCondition(o2)
		case string:
			parseO2 = o2
		case int:
			parseO2 = strconv.Itoa(o2)
		case float64:
			parseO2 = strconv.FormatFloat(o2,'f',-1,64)
		default:
			this.lastError = PersistentStore.NewPSErrC(PersistentStore.WhereConditionParseErr)
		}

		conditionClause = "("+parseO1+") "+ string(operator) + " (" +parseO2+")"
	}
	return conditionClause
}


