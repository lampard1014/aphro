package Session
import (
	"time"
	"math/rand"
	"strconv"
	"github.com/lampard1014/aphro/PersistentStore/Redis"
	"strings"
	"github.com/lampard1014/aphro/Gateway/error"

	"github.com/lampard1014/aphro/Encryption"
)

const (
	tokenDuration = 24 * 3600 * time.Second //1 day
)

const (
	//  common biz error //
	SessionServiceError = iota + 200
	//session过期
	SessionServiceError_SessionExpired
)

func FetchSessionTokenValue(sessionToken string) (uid string, merchantID string, err error) {
	var returnErr error = nil

	token,_,err := QuerySessionToken(sessionToken)

	if err == nil && token != "" {
		sessionTokenValue := token
		splitValue := strings.Split(sessionTokenValue, "#")
		uidAndMerchantID := strings.Split(splitValue[0],"@")
		uid = uidAndMerchantID[0]
		merchantID = uidAndMerchantID[1]
	} else {
		returnErr = AphroError.New(AphroError.BizError,"session 过期 请重新登录")
	}
	return uid,merchantID,returnErr
}

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
		redis.Connect()
		defer redis.Close()
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

func DecodeSessionTokenRequestStr(sessionTokenRequestStr string)(string ,error) {
	data , err := Encryption.Base64Decode(sessionTokenRequestStr)
	if err == nil {
		rsadata,err := Encryption.RsaDecryption(data)
		return string(rsadata),err
	} else {
		return "",err
	}
}

func CreateSessionToken(sessionTokenRequestStr string,uid uint32, merchantID uint32) (token string, ttl int64,  err error) {
	_uid := strconv.FormatUint(uint64(uid),36)
	_merchantID := strconv.FormatUint(uint64(merchantID),36)

	//general key
	t :=  time.Now().Unix()
	var tokenKey string
	tokenKeyPrefix := _uid
	tokenKey = strconv.FormatUint(uint64(t),36)
	tokenKeySuffix := _merchantID
	tokenKey = tokenKeyPrefix + tokenKey + tokenKeySuffix + sessionTokenRequestStr //uid + timestamp + mid + sessionTokenRequestStr
	tokenKey = Encryption.PswEncryption(tokenKey) //sha256 encryption
	//xxteaTokenKey,err := Encryption.XxteaEncryption(_encryptionkey,tokenKey)
	if err != nil {
		return
	}
	//general value
	tokenValue := _uid + "@" + _merchantID + "#" + sessionTokenRequestStr

	redis ,err := Redis.NewAPSRedis(nil)
	if err != nil{
		return "",0,err
	}  else {
		redis.Connect()
		defer redis.Close()
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
		redis.Connect()
		defer redis.Close()
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
		redis.Connect()
		defer redis.Close()
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
		redis.Connect()
		defer redis.Close()
		isExists,err := redis.IsExists(sessionToken)
		//res, err := c.IsExists(ctx, &redisPb.IsExistsRequest{Key:token})
		return isExists,err
	}
}

