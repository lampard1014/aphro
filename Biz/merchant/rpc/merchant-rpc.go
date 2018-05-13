package  main

import (

	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    "time"
    _ "github.com/go-sql-driver/mysql"
    merchantServicePB "github.com/lampard1014/aphro/Biz/merchant/pb"
    "github.com/lampard1014/aphro/CommonBiz/Response/PB"
    "github.com/lampard1014/aphro/CommonBiz/Response"

    "github.com/lampard1014/aphro/Gateway/error"

    "strconv"
    "github.com/lampard1014/aphro/Encryption"
    "github.com/lampard1014/aphro/PersistentStore/Redis"
    "github.com/lampard1014/aphro/PersistentStore/MySQL"
    "fmt"
    "github.com/lampard1014/aphro/CommonBiz/Session"
)

const (
	port  = ":10089"
    Scene_ChangePsw = 2
    Scene_MerchantRegister = 1
    verifyCodeDuration = 60*30
)

type merchantService struct{}



func (s *merchantService) MerchantOpen(ctx context.Context, in *merchantServicePB.MerchantOpenRequest) (*Aphro_CommonBiz.Response, error) {
    merchantName := in.Name
    cellphone := in.Cellphone
    address := in.Address
    paymentBit := in.PaymentBit

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    mysql,err := MySQL.NewAPSMySQL(nil)
    if err != nil {
        returnErr = err
    } else {
       m,ok := mysql.Connect().(*MySQL.APSMySQL)
        if ok {
            querySQL := "INSERT INTO merchant (`merchant_name`,`merchant_address`,`payment_type`,`cellphone`) VALUES( ?, ?, ?, ?)"

            lastInsertID , err := m.Query(querySQL,merchantName, address, paymentBit, cellphone).LastInsertId()
            returnErr = err
            if err == nil && lastInsertID > 0 {
                res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantOpenResponse{Success:true})
            }
            defer m.Close()
        } else {
            returnErr = AphroError.New(AphroError.BizError,"类型断言错误")
        }
    }
    return res,returnErr
}

func (s *merchantService) MerchantRegister(ctx context.Context, in *merchantServicePB.MerchantRegisterRequest) (*Aphro_CommonBiz.Response, error) {
    cellphone, psw, err := Encryption.ParseUsernameAndPsw(in.Key)
    name := in.Name
    role := in.Role
    verifyCode := in.VerifyCode
    merchantID := in.MerchantID

    var returnErr error = err
    var res *Aphro_CommonBiz.Response = nil

    //step1 验证码检查

    Redis.NewAPSRedis(nil)
    redis ,err := Redis.NewAPSRedis(nil)
    returnErr = err
    if err != nil{
        returnErr = err
    } else {
        checkKey := "_verify_sms_"+ cellphone + "@" + strconv.Itoa(Scene_MerchantRegister)
        redis.Connect()
        queryRes, queryRedisErr := redis.Query(checkKey)
        hasError := queryRedisErr != nil

        if !hasError && queryRes!= "" && queryRes ==  verifyCode {
            mysql,err := MySQL.NewAPSMySQL(nil)
            if err != nil {
                returnErr = err
            } else {
                m,ok := mysql.Connect().(*MySQL.APSMySQL)
                if ok {
                    if role == 1 {
                        querySQL := "INSERT INTO `merchant` (`merchant_name`,`merchant_address`,`payment_type`,`cellphone`)VALUES(?,?,?,?)"
                        _ , err := m.Query(querySQL,"@", "@", 0, cellphone).LastInsertId()
                        returnErr = err
                    }
                    if returnErr == nil {
                        //成功新建商户 继续 新建操作员，todo 事务回滚
                        querySQL := "INSERT INTO `merchant_account` (`name`,`cellphone`,`psw`,`role`,`merchant_id`)VALUES(?,?,?,?,?)"
                        lastInsertID , err := m.Query(querySQL,name, cellphone, Encryption.PswEncryption(psw), role,merchantID).LastInsertId()
                        returnErr = err
                        if err == nil && lastInsertID > 0 {
                            res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantRegisterResponse{Success:true})
                        }
                    }

                    defer m.Close()
                } else {
                    returnErr = AphroError.New(AphroError.BizError,"类型断言错误")
                }
            }
        } else {
            returnErr = AphroError.New(AphroError.BizError,"验证码验证错误")
        }
    }
    return res,returnErr
}

func (s *merchantService) MerchantChangePsw(ctx context.Context, in *merchantServicePB.MerchantChangePswRequest) (*Aphro_CommonBiz.Response, error) {

    newPsw := in.NewPsw
    sessionToken := in.SessionToken
    verifyCode := in.VerifyCode
    cellphone := in.Cellphone
    scene := uint32(Scene_ChangePsw)
    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil

    //检查session 是否合法
    isTokenVaildate, err := Session.IsSessionTokenVailate(sessionToken)
    returnErr = err
    if err == nil {
        if isTokenVaildate {
            base64Decode,err := Encryption.Base64Decode(newPsw)
            if err  == nil {
                rawData, RSADecryptionErr := Encryption.RsaDecryption(base64Decode)
                err = RSADecryptionErr
                if RSADecryptionErr == nil {
                    newPsw := string(rawData)
                    if newPsw != "" {
                        verifyCodeRes,err := s.MerchantVerifyCode(ctx,&merchantServicePB.MerchantVerifyCodeRequest{Cellphone:cellphone,Scene:scene,SmsCode:verifyCode})
                        var vcr *merchantServicePB.MerchantVerifyCodeResponse = nil
                        err = Response.UnmarshalAny(verifyCodeRes.Result,vcr)
                        if err == nil && vcr.Success{
                            mysql,err := MySQL.NewAPSMySQL(nil)
                            if err == nil {
                                m, ok := mysql.Connect().(*MySQL.APSMySQL)
                                if ok {
                                    querySQL := "update `merchant_account` set psw = ? where cellphone = ? limit 1"

                                    _ , err := m.Query(querySQL,Encryption.PswEncryption(newPsw),cellphone).RowsAffected()
                                    if err == nil  {
                                        res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantChangePswResponse{Success:err == nil})
                                    }
                                    defer m.Close()
                                } else {
                                    err = AphroError.New(AphroError.BizError,"mysql类型断言错误")
                                }
                            }
                        }
                    } else {
                        err = AphroError.New(AphroError.BizError,"解析密码错误")
                    }
                }
            }
            returnErr = err
        } else {
            returnErr = AphroError.New(AphroError.BizError,"session 过期 请重新登录")
        }
    } else {
        returnErr = err
    }
    return res,returnErr
}

func (s *merchantService) MerchantLogin(ctx context.Context, in *merchantServicePB.MerchantLoginRequest) (*Aphro_CommonBiz.Response, error) {
    cellphone, psw, err:= Encryption.ParseUsernameAndPsw(in.Key)
    tokenRequest := in.TokenRequest
    merchantID := in.MerchantID
    var returnErr error = err
    var sessionToken string = ""
    var (
        uid uint32
        name string
        role int
    )
    var res *Aphro_CommonBiz.Response = nil
    mysql,err := MySQL.NewAPSMySQL(nil)
    if err == nil {
        m, ok := mysql.Connect().(*MySQL.APSMySQL)
        if ok {
            querySQL := "SELECT id as uid, name ,role,merchant_id FROM merchant_account WHERE cellphone = ? AND psw = ? AND merchantID = ? LIMIT 1"
            err := m.Query(querySQL,cellphone, psw,merchantID).FetchRow(&uid,&name,&role,&merchantID)
            if err == nil  {
                token, _, err := Session.CreateSessionToken(tokenRequest,uid,merchantID)
                if err == nil {
                    sessionToken = token
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantLoginResponse{SessionToken:sessionToken,Success:true})
                }
                returnErr = err
            }
            defer m.Close()
        } else {
            returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
        }
    } else {
        returnErr = err
    }
    return res,returnErr
}

func (s *merchantService) MerchantInfo(ctx context.Context, in *merchantServicePB.MerchantInfoRequest) (*Aphro_CommonBiz.Response, error) {

    uid, merchantID, sessionTokenError := Session.FetchSessionTokenValue(in.Token)
    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil

    var (
        merchantName string
        name string
        role int
    )
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "SELECT m.merchant_name,ma.role,ma.name FROM merchant_account AS ma, merchant AS m WHERE ma.merchant_id = ? and ma.ID = ? limit 1"
                err := m.Query(querySQL,merchantID, uid).FetchRow(&merchantName,&role,&name)
                if err == nil {
                    //制作 令牌
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantInfoResponse{MerchantName:merchantName,MerchantAccount:&merchantServicePB.InnerMerchantAccount{Role:uint32(role),Name:name}})
                } else if MySQL.ISErrorNoRows(err) {
                    //没有记录
                    returnErr = AphroError.New(AphroError.BizError,"没有商户信息")
                } else {
                    returnErr = err
                }
                defer m.Close()
            } else {
                returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        } else {
            returnErr = err
        }
    } else {
        returnErr = sessionTokenError
    }

    return res,returnErr
}

func (s *merchantService) MerchantRoomInfo(ctx context.Context, in *merchantServicePB.MerchantRoomInfoRequest) (*Aphro_CommonBiz.Response, error) {
    return nil,nil
}
//  新增商户服务信息
func (s *merchantService) MerchantWaiterCreate(ctx context.Context, in *merchantServicePB.MerchantWaiterCreateRequest) (*Aphro_CommonBiz.Response, error) {
    token := in.Token
    merchantID := in.MerchantID
    name := in.Name
    reserve := in.Reserve
    //验证token合法性
    _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(token)
    var res *Aphro_CommonBiz.Response = nil
    var returnErr error = nil
    
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "insert into merchant_waiter (`name`,`reserve`,`merchant_id`) values (?,?,?)"

                _,err := m.Query(querySQL,name,reserve,merchantID).LastInsertId()
                if err == nil {
                    //制作 令牌
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantWaiterCreateResponse{Success:true})
                } else {
                    returnErr = err
                }
                defer m.Close()
            } else {
                returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        } else {
            returnErr = err
        }
    } else {
        returnErr =  sessionTokenError

    }
    return res,returnErr
}

// 删除商户服务信息
func (s *merchantService) MerchantWaiterDelete(ctx context.Context, in *merchantServicePB.MerchantWaiterDeleteRequest) (*Aphro_CommonBiz.Response, error) {
    token := in.Token
    waiterid := in.Waiterid
    //验证token合法性
    isVaildate, sessionTokenError := Session.IsSessionTokenVailate(token)
    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    if isVaildate {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "DELETE FROM merchant_waiter where id = ? "
                _,err := m.Query(querySQL,waiterid).RowsAffected()
                if err == nil {
                    //制作 令牌
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantWaiterDeleteResponse{Success:true})
                } else {
                    returnErr = err
                }
                defer m.Close()
            } else {
                returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        } else {
            returnErr = err
        }
    } else {
        returnErr =  sessionTokenError
    }

    return res,returnErr
}

func (s *merchantService) MerchantVerifyCode(ctx context.Context, in *merchantServicePB.MerchantVerifyCodeRequest) (*Aphro_CommonBiz.Response, error) {
    redis ,err := Redis.NewAPSRedis(nil)
    var res *Aphro_CommonBiz.Response = nil

    if err != nil{
        return res,err
    } else {
        redis.Connect()
        defer redis.Close()

        cellphone :=  in.Cellphone
        scene :=  strconv.Itoa(int(in.Scene))
        smsCode := in.SmsCode
        checkKey := "_verify_sms_"+ cellphone + "@" + scene
        value,err := redis.Query(checkKey)
        isVaildate := value == smsCode
        res,err = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantVerifyCodeResponse{Success:isVaildate})
        return res,err
    }
}

func (s *merchantService) MerchantSendCode(ctx context.Context, in *merchantServicePB.MerchantSendCodeRequest) (*Aphro_CommonBiz.Response, error) {
    redis ,err := Redis.NewAPSRedis(nil)
    var res *Aphro_CommonBiz.Response = nil

    if err != nil{
        return res,err
    } else {
        redis.Connect()
        defer redis.Close()
        cellphone :=  in.Cellphone
        scene :=  strconv.Itoa(int(in.Scene))
        checkKey := "_verify_sms_"+ cellphone + "@" + scene
        smsCodeTmp := "123456"
        ttl := uint64(verifyCodeDuration * time.Second)//60秒过期
        success,err :=redis.Set(checkKey,smsCodeTmp,int64(ttl))
        res,err = Response.NewCommonBizResponse(0,err.Error(),&merchantServicePB.MerchantSendCodeResponse{Success:success})
        return res,err
    }
}

func deferFunc() {
    if err := recover(); err != nil {
        fmt.Println("error happend:")
        fmt.Println(err)
    }
}

// // auth 验证Token
// func auth(ctx context.Context) error {
//     md, ok := metadata.FromContext(ctx)
//     if !ok {
//         return grpc.Errorf(codes.Unauthenticated, "无Token认证信息")
//     }

//     var (
//         appid  string
//         appkey string
//     )

//     if val, ok := md["appid"]; ok {
//         appid = val[0]
//     }

//     if val, ok := md["appkey"]; ok {
//         appkey = val[0]
//     }

//     if appid != "101010" || appkey != "i am key" {
//         return grpc.Errorf(codes.Unauthenticated, "Token认证信息无效: appid=%s, appkey=%s", appid, appkey)
//     }

//     return nil
// }


func main() {
    defer deferFunc()
    lis, err := net.Listen("tcp", port)
    if err != nil {
        log.Fatal(err)
    }

    // var opts []grpc.ServerOption //签名 和 验签

    // // 注册interceptor
    // var interceptor grpc.UnaryServerInterceptor
    // interceptor = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
    //     err = auth(ctx)
    //     if err != nil {
    //         return
    //     }
    //     // 继续处理请求
    //     return handler(ctx, req)
    // }
    // opts = append(opts, grpc.UnaryInterceptor(interceptor))


    s := grpc.NewServer(grpc.UnaryInterceptor(Response.UnaryServerInterceptor))//opts...)
    merchantServicePB.RegisterMerchantServiceServer(s, new(merchantService))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }
}

//func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
//    log.Printf("before handling. Info: %+v", info)
//    resp, err := handler(ctx, req)
//
//    fmt.Println("reflect", reflect.TypeOf(resp))
//
//    CommonBiz.NewCommonBizResponse(0,err.Error(),resp.(*proto.Message))
//
//    log.Printf("after handling. resp: %+v", resp)
//    return resp, err
//}
//// StreamServerInterceptor is a gRPC server-side interceptor that provides Prometheus monitoring for Streaming RPCs.
//func StreamServerInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
//    log.Printf("before handling. Info: %+v", info)
//    err := handler(srv, ss)
//    log.Printf("after handling. err: %v", err)
//    return err
//}
