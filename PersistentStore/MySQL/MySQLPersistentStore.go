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

	DELIMITER_COLON = ":"
	DELIMITER_COMMA = ","
	DELIMITER_SPACE = " "
	LEFT_BRACKETS = "("
	RIGHT_BRACKETS = ")"

	SELECT_ALL = "*"
	)

func ISErrorNoRows(err error)bool {
	return err == sql.ErrNoRows
}

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
	APSMySQLToken_LIMIT
	APSMySQLToken_FROM
	APSMySQLToken_WHERE
	APSMySQLToken_AS
	APSMySQLToken_ORDERBY
	APSMySQLToken_GROUPBY
	APSMySQLToken_HAVING
)

var APSMySQLTokenMap = map[APSMySQLToken]string{
	APSMySQLToken_SELECT:"SELECT",
	APSMySQLToken_INSERT:"INSERT INTO",
	APSMySQLToken_UPDATE:"UPDATE",
	APSMySQLToken_DELETE:"DELETE",
	APSMySQLToken_LIMIT:"LIMIT",
	APSMySQLToken_FROM:"FROM",
	APSMySQLToken_WHERE:"WHERE",
	APSMySQLToken_AS:"AS",
	APSMySQLToken_ORDERBY:"ORDER BY",
	APSMySQLToken_GROUPBY:"GROUP BY",
	APSMySQLToken_HAVING:"HAVING",
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

	//lastRecordID uint64
	lastError error
	//rowAffected int64
	rawResult interface{}

}

func (this *APSMySQLResult)LastInsertId() (int64, error) {
	d,ok := this.rawResult.(sql.Result)
	var lastInsertId int64 = 0
	if  ok {
		lastInsertId , this.lastError = d.LastInsertId()
	} else {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.ResultTypeErr)
	}

	return lastInsertId,this.lastError
}

func (this *APSMySQLResult)RowsAffected() (int64, error) {

	d,ok := this.rawResult.(sql.Result)
	var rowsAffected int64 = 0
	if  ok {
		rowsAffected , this.lastError = d.RowsAffected()
	} else {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.ResultTypeErr)
	}
	return rowsAffected,this.lastError
}

func (this *APSMySQLResult)FetchRow(dest...interface{})(error) {

	d,ok := this.rawResult.(*sql.Row)
	if ok {
		this.lastError = d.Scan(dest...)
	} else {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.ResultTypeErr)
	}
	return this.lastError
}

//type APSMySQLResultEnumerator func(dest...interface{}){};
// todo checkout : is golang pass by value ?
func (this *APSMySQLResult)FetchAll(callFunc func(outer...interface{}),in...interface{})(error) {
	d,ok := this.rawResult.(*sql.Rows)
	if ok {
		for d.Next() {
			 err := d.Scan(in)
			if err != nil {
				break
			} else {
				if callFunc != nil{
					callFunc(in...)
				}
			}
			this.lastError = err
		}
	} else {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.ResultTypeErr)
	}
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

	limit []string
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

func (this *APSMySQL)Connect()(PersistentStore.IAphroPersistentStore) {
	c := this.client.config.GetOptions()

	db, dbOpenErr := sql.Open(c[KConfigKey_DriverName].(string), c[KConfigKey_DSN].(string))
	dbOpenErr = db.Ping()
	if (dbOpenErr == nil) {
		this.client.mysqlClient = db
	} else {
		this.lastError = dbOpenErr
	}
	//this.Reset()
	return this
}

func (this *APSMySQL)Close()(PersistentStore.IAphroPersistentStore) {
	defer this.Reset()
	defer this.client.mysqlClient.Close()
	return this
}

func (this *APSMySQL)Reset()(PersistentStore.IAphroPersistentStore) {
	this.result = nil
	this.lastError = nil
	this.columns = []*APSField{}
	this.entities = []*APSEntity{}
	this.entitiesJoin = []APSMySQLEntityJoin{}

	this.insertValues = [][]interface{}{}
	this.updateColumns = map[string]interface{}{}
	this.token = 0

	this.limit = []string{}
	this.orderBy = ""

	this.bindValues = []interface{}{}
	this.prepareStatement = ""

	return this
}



func (this *APSMySQL) Select(columnsAs map[string]string)(PersistentStore.IAphroSQLPersistentStore) {
	//this.Reset()
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

func (this *APSMySQL)From(entity string, entityAlias string)(PersistentStore.IAphroSQLPersistentStore) {
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
	//this.Reset()
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
	//this.Reset()
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
	//this.Reset()
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

func (this *APSMySQL)Query(querySQL string, bindsValues...interface{})(PersistentStore.IAphroPersistentStoreResult) {
	//this.Reset()
	this.prepareStatement = querySQL
	this.bindValues = bindsValues

	checkForToken := strings.Split(strings.ToUpper(strings.TrimSpace(this.prepareStatement))," ")[0]
	if checkForToken == APSMySQLTokenMap[APSMySQLToken_SELECT]{
		this.token = APSMySQLToken_SELECT
	} else if checkForToken == APSMySQLTokenMap[APSMySQLToken_DELETE] {
		this.token = APSMySQLToken_DELETE
	} else if checkForToken == APSMySQLTokenMap[APSMySQLToken_UPDATE] {
		this.token = APSMySQLToken_UPDATE
	} else {
		this.token = APSMySQLToken_INSERT
	}
	this.query()
	return this.result
}

func (this *APSMySQL)Execute(bindsValues...interface{})(PersistentStore.IAphroPersistentStoreResult){

	var queryStatment string = ""

	//query Token
	queryToken := APSMySQLTokenMap[this.token]
	queryStatment += queryToken

	//query Fields
	var queryFields string = SELECT_ALL
	if this.columns != nil {

		queryFeildsFormer := []string{}
		for _,cv := range this.columns {
			var fieldToken  string = ""
			fieldName := cv.filedName
			fieldToken += fieldName
			if cv.alias != "" {
				fieldToken += DELIMITER_SPACE + APSMySQLTokenMap[APSMySQLToken_AS] + DELIMITER_SPACE + cv.alias
			}
			queryFeildsFormer = append(queryFeildsFormer,fieldToken)
		}
		queryFields = strings.Join(queryFeildsFormer,DELIMITER_COMMA)
	}

	queryStatment += DELIMITER_SPACE + queryFields + DELIMITER_SPACE

	// From  from a as a innerjoin b as b
	var fromToken string = APSMySQLTokenMap[APSMySQLToken_FROM]
	if this.entities != nil {
		entitiesFormer := []string{}
		for ei,ev := range this.entities {
			var entityToken  string = ""
			entityName := ev.entityName
			entityToken += entityName
			if ev.alias != "" {
				entityToken += DELIMITER_SPACE + APSMySQLTokenMap[APSMySQLToken_AS] + DELIMITER_SPACE + ev.alias
			}

			if this.entitiesJoin != nil {
				entityToken += DELIMITER_SPACE + APSMySQLEntityJoinMap[this.entitiesJoin[ei/2]] + DELIMITER_SPACE
			}

			entitiesFormer = append(entitiesFormer,entityToken)
		}
		fromToken += DELIMITER_SPACE + strings.Join(entitiesFormer,DELIMITER_SPACE)
	} else {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.NoEntitySpecify)
	}

	queryStatment += DELIMITER_SPACE + fromToken + DELIMITER_SPACE
	// where
	queryStatment += DELIMITER_SPACE + this.wheres + DELIMITER_SPACE
	//order by
	queryStatment += DELIMITER_SPACE + this.orderBy + DELIMITER_SPACE
	//limit
	queryStatment += DELIMITER_SPACE + APSMySQLTokenMap[APSMySQLToken_LIMIT] + strings.Join(this.limit,DELIMITER_COMMA) + DELIMITER_SPACE
	// do query
	this.prepareStatement = queryStatment
	this.bindValues = bindsValues
	this.query()
	return this.result
}

func (this *APSMySQL) query()(*APSMySQL) {
	if this.token ==  APSMySQLToken_SELECT {
		//Query Row
		if len(this.limit) == 1 && "1" == this.limit[0] {
			this.result = &APSMySQLResult{}
			rawResult := this.client.mysqlClient.QueryRow(this.prepareStatement,this.bindValues...)
			this.result.rawResult = rawResult
			//queryRowErr :=.Scan(&uid,&name,&role,&merchantID)
		} else {
			//Query All
		}
	} else {
		stmtIns, stmtInsErr := this.client.mysqlClient.Prepare(this.prepareStatement)
		if stmtInsErr != nil {
			this.lastError = stmtInsErr
		} else {
			this.result = &APSMySQLResult{}
			res, err := stmtIns.Exec(this.bindValues...)
			this.result.rawResult = res
			this.lastError = err
			this.result.lastError = err
		}
	}
	return this
}


//[3,4]
func (this *APSMySQL) Limit(s...string)(PersistentStore.IAphroSQLPersistentStore) {
	this.limit = s
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
	Operator APSMySQLOperator
	Operand1 interface{}
	Operand2 interface{}
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
		operator := condition.Operator
		operand1 := condition.Operand1
		operand2 := condition.Operand1

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
		conditionClause = LEFT_BRACKETS + parseO1+ RIGHT_BRACKETS + DELIMITER_SPACE + string(operator) + DELIMITER_SPACE + LEFT_BRACKETS + parseO2 + RIGHT_BRACKETS
	}
	return conditionClause
}


