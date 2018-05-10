package Redis
import (

	"time"
	r "github.com/go-redis/redis"

	"github.com/lampard1014/aphro/PersistentStore"
)

const (
	//port  = ":10101"
)

const (
	kConfigKey_Addr = "ConfigKey_Addr"
	kConfigKey_PSW = "ConfigKey_PSW"
	kConfigKey_DB = "ConfigKey_DB"

	vConfigKey_Addr = "127.0.0.1:6379"
	vConfigKey_PSW = ""
	vConfigKey_DB = 0
)

//
//func createClient()*r.Client {
//	redisCli := r.NewClient(&r.Options{
//		Addr:     "127.0.0.1:6379",
//		Password: "", // no password set
//		DB:       0,  // use default DB
//	})
//	pong, err := redisCli.Ping().Result()
//	fmt.Println(pong, err,redisCli)
//	return redisCli
//}

type APSRedisClientConfiguration struct {
	config map[string]interface{}
}

func (this *APSRedisClientConfiguration)SetOptions(nc map[string]interface{}){
	this.config = nc
}

func (this *APSRedisClientConfiguration)GetOptions()(map[string]interface{}) {
	return this.config
}

func NewAPSRedisClientConfiguration(config map[string]interface{}) *APSRedisClientConfiguration {
	this := &APSRedisClientConfiguration{}
	this.SetOptions(config)
	return this
}

type APSRedisClient struct{
	config *APSRedisClientConfiguration
	redisClient *r.Client
}


func (this APSRedisClient)FetchClient()(interface{}) {
	return this.redisClient
}

func (this APSRedisClient)FetchConfiguration()(PersistentStore.IAphroPersistentStoreClientConfiguration) {
	return this.config
}

func (this APSRedisClient)SetConfiguration(c PersistentStore.IAphroPersistentStoreClientConfiguration)(error) {
	var returnErr error = nil
	d,b := c.(*APSRedisClientConfiguration)
	if !b {
		this.config = d
	} else {
		returnErr = PersistentStore.NewPSErrC(PersistentStore.ConfigurationErr)
	}
	return returnErr
}

func NewAPSRedisClient(config *APSRedisClientConfiguration) (*APSRedisClient,error) {
	var returnErr error = nil
	c := &APSRedisClient{}
	returnErr = c.SetConfiguration(config)
	return c,returnErr
}


type APSRedis struct {

	client *APSRedisClient
	lastError error
	result interface{}
}

func NewAPSRedis(userConfig map[string]interface{} ) (*APSRedis,error) {
	var apsRedis *APSRedis = &APSRedis{}

	var addr string = vConfigKey_Addr
	var psw string = vConfigKey_PSW
	var db int = vConfigKey_DB

	if taddr,ok := userConfig[kConfigKey_Addr]; ok {
		addr = taddr.(string)
	}

	if tpsw,ok := userConfig[kConfigKey_PSW]; ok {
		psw = tpsw.(string)
	}

	if tdb,ok := userConfig[kConfigKey_DB]; ok {
		db = tdb.(int)
	}

	c := map[string]interface{} {
		kConfigKey_Addr:addr,
		kConfigKey_PSW:psw,
		kConfigKey_DB:db,
	}

	apsConfig := NewAPSRedisClientConfiguration(c)
	client, err := NewAPSRedisClient(apsConfig)
	apsRedis.client = client
	return apsRedis, err
}


func (this *APSRedis) Reset()(PersistentStore.IAphroPersistentStore) {
	this.result = nil
	this.lastError = nil
	return this
}

func (this *APSRedis) Connect() (PersistentStore.IAphroPersistentStore) {
	c := this.client.config.GetOptions()
	redisCli := r.NewClient(&r.Options{
		Addr:     c[kConfigKey_Addr].(string),
		Password: c[kConfigKey_PSW].(string), // no password set
		DB:       c[kConfigKey_DB].(int),  // use default DB
	})
	_, err := redisCli.Ping().Result()
	//fmt.Println(pong, err,redisCli)
	this.client.redisClient = redisCli
	this.lastError = err
	this.Reset()
	return this
}

func (this *APSRedis) Close() (PersistentStore.IAphroPersistentStore) {
	return this
}

func (this *APSRedis)IsExists(key string)(isExists bool,err error) {
	cli := this.client.redisClient
	isSuccess,err := cli.Exists(key).Result()
	isExists = isSuccess == 1
	this.lastError = err
	this.Close()
	return
}

func (this *APSRedis)ExpireAt(key string, ttl int64)(success bool,err error) {
	isSuccess,err := this.client.redisClient.ExpireAt(key,time.Unix(int64(ttl),0)).Result()
	this.lastError = err
	this.Close()
	return isSuccess,err
}

func (this *APSRedis)QueryTTL(key string)(ttl int64,err error) {
	res,err := this.client.redisClient.TTL(key).Result()
	this.lastError = err
	this.Close()
	return int64(res.Seconds()), err
}


func (this *APSRedis)Query(key string)(value string,err error) {
	val, err := this.client.redisClient.Get(key).Result()
	this.lastError = err
	if val != "" && err != nil {
		this.lastError = PersistentStore.NewPSErrC(PersistentStore.NoKeyExisted)
	}
	this.Close()
	return val,this.lastError
}

func (this *APSRedis)Delete(key string)(success bool,err error) {
	isSuccess,err := this.client.redisClient.Del(key).Result()
	this.lastError = err
	this.Close()
	return isSuccess == 1,err
}

func (this *APSRedis)Set(key string ,value string ,ttl int64)(success bool,err error) {
	_,cmderr := this.client.redisClient.Set(key,value,time.Duration(ttl)).Result()
	this.lastError = cmderr
	this.Close()
	return cmderr == nil ,cmderr
}


