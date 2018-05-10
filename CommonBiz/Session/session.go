package Session
import (
	"time"
	"math/rand"
	"strconv"
	"github.com/lampard1014/aphro/PersistentStore/Redis"
)

const (
	tokenDuration = 24 * 3600 * time.Second //1 day
)

func  GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}
	return string(result)
}

func QuerySessionToken(sessionToken string) (token string, ttl int64,  err error) {
	redis ,err := Redis.NewAPSRedis(nil)
	if err != nil {
		return "",0,err
	} else {
		var returnErr error = nil
		qr,err1 := redis.Query(sessionToken)
		qtr, err2 := redis.QueryTTL(sessionToken)
		var hasErr bool = (err1 !=nil || err2 != nil)
		hasErr = hasErr && qtr > 0

		if hasErr {
			if err1 != nil {
				returnErr = err1
			} else {
				returnErr = err2
			}
		}
		return qr,qtr,returnErr
	}
}

func CreateSessionToken(sessionTokenRequestStr string,uid uint32, merchantID uint32) (token string, ttl int64,  err error) {
	_uid := strconv.FormatUint(uint64(uid),36)
	_merchantID := strconv.FormatUint(uint64(merchantID),36)
	_encryptionkey := sessionTokenRequestStr

	//general key
	t :=  time.Now().Unix()
	tokenKeyPrefix := _uid
	tokenKey := strconv.FormatUint(uint64(t),36)
	tokenKeySuffix := _merchantID
	tokenKey = tokenKeyPrefix + tokenKey + tokenKeySuffix
	//general value
	tokenValue := _uid + "@" + _merchantID + "#" + _encryptionkey

	redis ,err := Redis.NewAPSRedis(nil)
	if err != nil{
		return "",0,err
	}  else {
		_ ,err := redis.Set(tokenKey,tokenValue,int64(tokenDuration))
		if err != nil {
			return "",0,err
		} else {
			return tokenKey,int64(tokenDuration), err
		}
	}
}

func DeleteSessionToken(sessionToken string) (error) {
	var returnErr error = nil
	redis ,err := Redis.NewAPSRedis(nil)
	returnErr = err
	if err == nil{
		_,err := redis.Delete(sessionToken)
		returnErr = err
	}
	return returnErr
}

func RenewSessionToken(sessionToken string) (ttl int64,err error) {
	redis ,err := Redis.NewAPSRedis(nil)
	if err != nil{
		return 0,err
	} else {
		ttl := int64(time.Now().Add(tokenDuration).Unix())
		_,err := redis.ExpireAt(sessionToken,ttl)
		return ttl, err
	}
}

func IsSessionTokenVailate(sessionToken string) (bool,error) {

	redis ,err := Redis.NewAPSRedis(nil)
	if err != nil{
		return false,err
	} else {
		isExists,err := redis.IsExists(sessionToken)
		//res, err := c.IsExists(ctx, &redisPb.IsExistsRequest{Key:token})
		return isExists,err
	}
}

