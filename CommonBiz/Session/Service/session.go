package Service

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	// "github.com/xxtea/xxtea-go/xxtea"
	sessionPb "github.com/lampard1014/aphro/CommonBiz/Session/PB"
	redisPb "github.com/lampard1014/aphro/redis/pb"
	"fmt"
	"time"
	"math/rand"
	"strconv"
)

const (
	port  = ":10088"
	redisRpcAddress = "192.168.140.23:10101"
	tokenDuration = 24 * 3600 * time.Second //1 day
	verifyCodeDuration = 60*30
)

type sessionService struct{}

func (s *sessionService) QuerySessionToken(ctx context.Context, in *sessionPb.SessionTokenQueryRequest) (*sessionPb.SessionTokenQueryResponse, error) {
	conn,err := grpc.Dial(redisRpcAddress,grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := redisPb.NewRedisServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	queryRes, err1 := c.Query(ctx, &redisPb.QueryRedisRequest{Key: in.SessionToken})
	QueryTtlRes,err := c.QueryTtl(ctx, &redisPb.QueryTtlRedisRequest{Key:in.SessionToken})

	hasError := err != nil || err1 != nil
	sessionValue := ""
	if queryRes!= nil  {
		fmt.Println("bingo",queryRes.Value)
		sessionValue = queryRes.Value
	}
	if QueryTtlRes.Ttl >= 0 &&  hasError{
		panic(err)
	}


	fmt.Println(sessionValue)

	return &sessionPb.SessionTokenQueryResponse{
		SessionToken:sessionValue,
		Ttl:QueryTtlRes.Ttl,
		Successed:!hasError,
	},err
}


func (s *sessionService) CreateSessionToken(ctx context.Context, in *sessionPb.SessionTokenCreateRequest) (*sessionPb.SessionTokenCreateResponse, error) {
	conn,err := grpc.Dial(redisRpcAddress,grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := redisPb.NewRedisServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	uid := strconv.FormatUint(uint64(in.Uid),36)
	merchantID := strconv.FormatUint(uint64(in.MerchantID),36)
	encryptionkey := in.SessionTokenRequestStr
	//general key
	t :=  time.Now().Unix()
	tokenKeyPrefix := uid
	tokenKey := strconv.FormatUint(uint64(t),36)
	tokenKeySuffix := merchantID
	tokenKey = tokenKeyPrefix + tokenKey + tokenKeySuffix
	//general value
	tokenValue := uid + "@" + merchantID + "#" + encryptionkey

	setRes, err := c.Set(ctx, &redisPb.SetRedisRequest{Key:tokenKey,Value:tokenValue,Ttl:uint64(tokenDuration)})
	if err != nil {
		panic(err)
	}
	// queryRes, err2 := c.Query(ctx, &redisPb.QueryRedisRequest{Key: token})

	// fmt.Println(token,in.SessionTokenRequestStr,err,time.Now(),int64(tokenDuration))

	// fmt.Println(queryRes,err2)

	return &sessionPb.SessionTokenCreateResponse{
		SessionToken:tokenKey,
		Ttl:int64(tokenDuration),
		Successed:setRes.Successed,
	},err
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


func (s *sessionService) DeleteSessionToken(ctx context.Context, in *sessionPb.DeleteSessionTokenRequest) (*sessionPb.DeleteSessionTokenResponse, error) {
	conn,err := grpc.Dial(redisRpcAddress,grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := redisPb.NewRedisServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// ttl := time.Now().Add(tokenDuration)
	delRes, err := c.Delete(ctx, &redisPb.DeleteRedisRequest{Key:in.SessionToken})
	if err != nil {
		panic(err)
	}
	return &sessionPb.DeleteSessionTokenResponse{
		Successed:delRes.Successed,
	},err

}

func (s *sessionService) RenewSessionToken(ctx context.Context, in *sessionPb.RenewSessionTokenRequest) (*sessionPb.RenewSessionTokenResponse, error) {
	conn,err := grpc.Dial(redisRpcAddress,grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := redisPb.NewRedisServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	ttl := int64(time.Now().Add(tokenDuration).Unix())
	token := in.SessionToken
	res, err := c.ExpireAt(ctx, &redisPb.ExpireAtRequest{Key:token,Ttl:ttl})
	if err != nil {
		panic(err)
	}
	return &sessionPb.RenewSessionTokenResponse{
		Ttl:ttl,
		Successed:res.Successed,
	},err
}

func (s *sessionService) IsSessionTokenVailate(ctx context.Context, in *sessionPb.IsSessionTokenVailateRequest) (*sessionPb.IsSessionTokenVailateResponse, error) {
	conn,err := grpc.Dial(redisRpcAddress,grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := redisPb.NewRedisServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	token := in.SessionToken
	res, err := c.IsExists(ctx, &redisPb.IsExistsRequest{Key:token})
	fmt.Println("res ", res,err)
	if err != nil {
		panic(err)
	}
	return &sessionPb.IsSessionTokenVailateResponse{
		Successed:res.IsExists,
	},err
}

func (s *sessionService) MerchantVerifyCode(ctx context.Context, in *sessionPb.MerchantVerifyCodeRequest) (*sessionPb.MerchantVerifyCodeResponse, error) {

	conn,err := grpc.Dial(redisRpcAddress,grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := redisPb.NewRedisServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cellphone :=  in.Cellphone
	scene :=  strconv.Itoa(int(in.Scene))
	smsCode := in.SmsCode
	checkKey := "_verify_sms_"+ cellphone + "@" + scene
	queryRes, err := c.Query(ctx, &redisPb.QueryRedisRequest{Key: checkKey})

	fmt.Println("xxx",queryRes, err)

	if err != nil {
		panic(err)
	}

	return &sessionPb.MerchantVerifyCodeResponse{
		Successed:queryRes.Value == smsCode,
	},err
}

func (s *sessionService) MerchantSendCode(ctx context.Context, in *sessionPb.MerchantSendCodeRequest) (*sessionPb.MerchantSendCodeResponse, error) {

	conn,err := grpc.Dial(redisRpcAddress,grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	c := redisPb.NewRedisServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	cellphone :=  in.Cellphone
	scene :=  strconv.Itoa(int(in.Scene))

	// token := in.Token
	checkKey := "_verify_sms_"+ cellphone + "@" + scene
	smsCodeTmp := "123456"
	ttl := uint64(verifyCodeDuration * time.Second)//60秒过期

	fmt.Println("scene,checkKey, smsCode :",scene,checkKey,smsCodeTmp)
	setRes, err := c.Set(ctx, &redisPb.SetRedisRequest{Key:checkKey,Value:smsCodeTmp,Ttl:ttl})
	if err != nil {
		panic(err)
	}

	return &sessionPb.MerchantSendCodeResponse{
		Successed:setRes.Successed,
	},err
}

func deferFunc() {
	if err := recover(); err != nil {
		fmt.Println("error happend:")
		fmt.Println(err)
	}
}

func main() {
	defer deferFunc()
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	s := grpc.NewServer()
	sessionPb.RegisterSessionServiceServer(s, new(sessionService))
	err = s.Serve(lis)
	if err != nil {
		log.Fatal(err)
	}
}
