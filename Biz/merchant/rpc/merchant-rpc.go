package  main

import (

	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    "time"
    "database/sql"

    _ "github.com/go-sql-driver/mysql"
    merchantServicePB "github.com/lampard1014/aphro/Biz/merchant/pb"
    "github.com/lampard1014/aphro/CommonBiz/Response/PB"
    "github.com/lampard1014/aphro/CommonBiz/Response"

    "github.com/lampard1014/aphro/gateway/error"

    "strconv"
    "github.com/lampard1014/aphro/Encryption"
    "github.com/lampard1014/aphro/PersistentStore/Redis"
    "github.com/lampard1014/aphro/PersistentStore/MySQL"
    "fmt"
    "github.com/lampard1014/aphro/CommonBiz/Session"
)

const (
	port  = ":10089"
    //redisRPCAddress = "192.168.140.23:10101"
    //sessionRPCAddress = "127.0.0.1:10088"
    //encyptRPCAddress = "127.0.0.1:10087"
    //mysqlDSN = "root:123456@tcp(192.168.140.23:3306)/iris_db"
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
    var operatorSuccess bool = false
    var returnErr error = nil
    //检查session 是否合法
    isTokenVaildate, err := Session.IsSessionTokenVailate(sessionToken)
    returnErr = err
    if err == nil {
        if isTokenVaildate {
            base64Decode,err := Encryption.Base64Decode(newPsw)
            returnErr = err
            if err  == nil {
                rawData, RSADecryptionErr := Encryption.RsaDecryption(base64Decode)
                returnErr = RSADecryptionErr
                if RSADecryptionErr == nil {
                    newPsw := string(rawData)
                    if newPsw != "" {
                        verifyCodeRes,err := MerchantVerifyCode(cellphone,scene,verifyCode)
                        

                        verifyCodeRes ,verifyCodeResErr := sessionRPCConnClient.MerchantVerifyCode(sessionRPCConnClientCtx,&sessionServicePB.MerchantVerifyCodeRequest{Cellphone:cellphone,Scene:scene,SmsCode:verifyCode})
                        if verifyCodeRes.Successed {
                            db, dbOpenErr := sql.Open("mysql", mysqlDSN)
                            defer db.Close()
                            dbOpenErr = db.Ping()
                            if dbOpenErr == nil {
                                stmtIns, stmtInsErr := db.Prepare("update `merchant_account` set psw = ? where cellphone = ? limit 1") // ? = placeholder
                                if stmtInsErr == nil {
                                    updateResult, updateErr := stmtIns.Exec(pswEncryption(newPsw),cellphone)
                                    if updateErr == nil {
                                        afftectedRow, afftectedRowErr := updateResult.RowsAffected()
                                        if afftectedRow != 1 || afftectedRowErr != nil {
                                            returnErr = afftectedRowErr
                                        } else {
                                            //成功
                                            operatorSuccess = true
                                        }
                                    } else {
                                        returnErr = updateErr
                                    }
                                } else {

                                    returnErr = stmtInsErr
                                }
                                defer stmtIns.Close()
                            } else {
                                returnErr = dbOpenErr
                            }
                        } else {
                            returnErr = verifyCodeResErr
                        }
                    } else {
                        returnErr = AphroError.New(AphroError.BizError,"解析密码错误")
                    }
                }

            }
        } else {
            returnErr = AphroError.New(AphroError.BizError,"session 过期 请重新登录")
        }
    } else {
        returnErr = isSessionTokenVailateErr
    }


    return &merchantServicePB.MerchantChangePswResponse{Successed:operatorSuccess},returnErr
}

func (s *merchantService) MerchantLogin(ctx context.Context, in *merchantServicePB.MerchantLoginRequest) (*merchantServicePB.MerchantLoginResponse, error) {
    cellphone, psw, err:= parseUsernameAndPsw(in.Key)
    tokenRequest := in.TokenRequest
    merchantID := in.MerchantID
    var returnErr error = err
    var sessionToken string = "";
    //查询数据库
    var (
        uid uint32
        name string
        role int
    )
    db, dbOpenErr := sql.Open("mysql", mysqlDSN)
    defer db.Close()
    dbOpenErr = db.Ping()
    if dbOpenErr == nil {
        queryRowErr := db.QueryRow("SELECT id as uid, name ,role,merchant_id FROM merchant_account WHERE cellphone = ? AND psw = ? AND merchantID = ? LIMIT 1", cellphone, psw,merchantID).Scan(&uid,&name,&role,&merchantID)
        if queryRowErr == nil {
            //制作 令牌
            sessionRPCConn,sessionRPCErr := grpc.Dial(sessionRPCAddress,grpc.WithInsecure())
            if sessionRPCErr == nil {
                c := sessionServicePB.NewSessionServiceClient(sessionRPCConn)

                ctx, cancel := context.WithTimeout(context.Background(), time.Second)

                defer cancel()

                createSessionTokenRes, err := c.CreateSessionToken(ctx, &sessionServicePB.SessionTokenCreateRequest{SessionTokenRequestStr:tokenRequest,Uid:uid,MerchantID:merchantID})
                if err == nil {
                    sessionToken = createSessionTokenRes.SessionToken

                } else {
                    returnErr = err
                }
            } else {
                returnErr = sessionRPCErr
            }
            defer sessionRPCConn.Close()
        } else if queryRowErr == sql.ErrNoRows {
            //没有记录
            returnErr = AphroError.New(AphroError.BizError,"密码或者用户名错误")
        } else {
            returnErr = queryRowErr
        }

    } else {
        returnErr = dbOpenErr
    }
    return &merchantServicePB.MerchantLoginResponse{SessionToken:sessionToken,Successed:returnErr == nil},returnErr
}

func (s *merchantService) MerchantInfo(ctx context.Context, in *merchantServicePB.MerchantInfoRequest) (*merchantServicePB.MerchantInfoResponse, error) {
    //验证token合法性
    uid, merchantID, sessionTokenError := fetchSessionTokenValue(in.Token)
    var returnErr error = nil

    var (
        merchantName string
        name string
        role int
    )
    if sessionTokenError == nil {

        db, dbOpenErr := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        dbOpenErr = db.Ping()
        if dbOpenErr == nil {
            queryRowErr := db.QueryRow("SELECT m.merchant_name,ma.role,ma.name FROM merchant_account AS ma, merchant AS m WHERE ma.merchant_id = ? and ma.ID = ? limit 1;", merchantID, uid).Scan(&merchantName,&role,&name)
            if queryRowErr == nil {
                //制作 令牌
            } else if queryRowErr == sql.ErrNoRows {
                //没有记录
               returnErr = AphroError.New(AphroError.BizError,"没用商户信息")
            } else {
                returnErr = queryRowErr
            }

        } else {
            returnErr = dbOpenErr
        }
    } else {
        returnErr = sessionTokenError
    }

    return &merchantServicePB.MerchantInfoResponse{MerchantName:merchantName,MerchantAccount:&merchantServicePB.InnerMerchantAccount{Role:uint32(role),Name:name}},returnErr
}


func (s *merchantService) MerchantRoomInfo(ctx context.Context, in *merchantServicePB.MerchantRoomInfoRequest) (*merchantServicePB.MerchantRoomInfoResponse, error) {
    return nil,nil
}
//  新增商户服务信息
func (s *merchantService) MerchantWaiterCreate(ctx context.Context, in *merchantServicePB.MerchantWaiterCreateRequest) (*Aphro.CommonBiz.Response, error) {
    token := in.Token
    merchantID := in.MerchantID
    name := in.Waiter.Name
    imageID := in.Waiter.ImageID
    //验证token合法性
    uid, merchantID, sessionTokenError := fetchSessionTokenValue(in.Token)

    fmt.Println(token,merchantID,name,imageID,uid);

    var returnErr error = nil

    var (
        success bool 
    )
    
    if sessionTokenError == nil {
        db, dbOpenErr := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        dbOpenErr = db.Ping()
        if dbOpenErr == nil {
            queryRowErr := db.QueryRow("").Scan()
            if queryRowErr == nil {
                success = true
            } else if queryRowErr == sql.ErrNoRows{
                //没有记录
                success = false
               returnErr = AphroError.New(AphroError.BizError,"没用商户信息")
            } else {
                success = false
                returnErr = queryRowErr
            }
        } else {
            success = false
            returnErr = dbOpenErr

        }

    } else {
        success = false
        returnErr =  sessionTokenError

    }

    return &merchantServicePB.MerchantWaiterCreateRespone{Successed:success},returnErr
}
// 删除商户服务信息
func (s *merchantService) MerchantWaiterDelete(ctx context.Context, in *merchantServicePB.MerchantWaiterDeleteRequest) (*merchantServicePB.MerchantWaiterDeleteRespone, error) {
    token := in.Token
    merchantID := in.MerchantID
    waiterid := in.Waiterid
    //验证token合法性
    uid, merchantID, sessionTokenError := fetchSessionTokenValue(in.Token)
    fmt.Println(token,merchantID,waiterid,uid);
    var returnErr error = nil

    var (
        success bool 
    )
    
    if sessionTokenError == nil {
        db, dbOpenErr := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        dbOpenErr = db.Ping()
        if dbOpenErr == nil {
            queryRowErr := db.QueryRow("").Scan()
            if queryRowErr == nil {
                success = true
            } else if queryRowErr == sql.ErrNoRows{
                //没有记录
                success = false
               returnErr = AphroError.New(AphroError.BizError,"没用商户信息")
            } else {
                success = false
                returnErr = queryRowErr
            }
        } else {
            success = false
            returnErr = dbOpenErr

        }

    } else {
        success = false
        returnErr =  sessionTokenError

    }

    return &merchantServicePB.MerchantWaiterDeleteRespone{Successed:success},returnErr
}

func (s *merchantService) MerchantVerifyCode(ctx context.Context, in *merchantServicePB.MerchantVerifyCodeRequest) (*Aphro_CommonBiz.Response, error) {

    conn,err := grpc.Dial(redisRPCAddress,grpc.WithInsecure())
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

    return &merchantServicePB.MerchantVerifyCodeResponse{
        Successed:queryRes.Value == smsCode,
    },err
}

func (s *merchantService) MerchantSendCode(ctx context.Context, in *merchantServicePB.MerchantSendCodeRequest) (*merchantServicePB.MerchantSendCodeResponse, error) {

    conn,err := grpc.Dial(redisRPCAddress,grpc.WithInsecure())
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

    return &merchantServicePB.MerchantSendCodeResponse{
        Successed:setRes.Successed,
    },err
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
