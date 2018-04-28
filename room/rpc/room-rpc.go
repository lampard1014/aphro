package  main

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    "time"
    "fmt"
    "strings"
    "errors"
    "database/sql"
    _ "github.com/go-sql-driver/mysql"
    roomServicePB "github.com/lampard1014/aphro/room/pb"
    sessionServicePB "github.com/lampard1014/aphro/session/pb"
)

const (
	port  = ":10090"
    sessionRPCAddress = "127.0.0.1:10101"
    mysqlDSN = "root:@tcp(127.0.0.1:3306)/iris_db"
)

type roomService struct{}

func checkSessionVaildate (sessionToken string) (isVailate, error) {
    var isSessionTokenOK bool = false
    var returnErr error = nil
    //先检查session 是否过期
    sessionRPCConn,sessionRPCErr := grpc.Dial(sessionRPCAddress,grpc.WithInsecure())
    if sessionRPCErr == nil {
        sessionRPCConnClient := sessionServicePB.NewSessionServiceClient(sessionRPCConn)

        sessionRPCConnClientCtx, sessionRPCConnClientCancel := context.WithTimeout(context.Background(), time.Second)

        defer sessionRPCConnClientCancel()
        var isSessionTokenOK bool = false
        isSessionTokenVailate, isSessionTokenVailateErr := sessionRPCConnClient.IsSessionTokenVailate(sessionRPCConnClientCtx, &sessionServicePB.IsSessionTokenVailateRequest{SessionToken:sessionToken})
        if isSessionTokenVailateErr == nil {
            isSessionTokenOK := isSessionTokenVailate.Successed

            if isSessionTokenOK {

            } else {
                returnErr = errors.New("session expired, pls login again")
            }
        } else {
            returnErr = isSessionTokenVailateErr
        }
    } else {
        returnErr = sessionRPCErr
    }
    defer sessionRPCConn.Close()
    return isSessionTokenOK, returnErr
}



func (s *roomService) TerminalBind(ctx context.Context, in *roomServicePB.RSTerminalBindRequest) (*roomServicePB.RSTerminalBindResponse, error) {

    sessionToken := in.SessionToken
    terminalCode := in.TerminalCode
    location := in.Location
    roomID := in.RoomID

    var returnErr error = nil

    isVailate, checkSessionError := checkSessionVaildate(sessionToken)

    if isVailate {
        
        db, dbOpenErr := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        // Open doesn't open a connection. Validate DSN data:
        dbOpenErr = db.Ping()
        if (dbOpenErr == nil) {
            stmtIns, stmtInsErr := db.Prepare("INSERT INTO merchant (`merchant_name`,`merchant_address`,`payment_type`,`cellphone`) VALUES( ?, ?, ?, ?)") // ? = placeholder
            if stmtInsErr == nil {
                insertResult, insertErr := stmtIns.Exec(merchantName, address, paymentBit, cellphone) 
                if insertErr == nil {
                    afftectedRow, afftectedRowErr := insertResult.RowsAffected()
                    if afftectedRow != 1 || afftectedRowErr == nil {
                        returnErr := afftectedRowErr
                    }
                } else {
                    returnErr = insertErr
                }
            } else {
                returnErr = stmtInsErr
            }
            defer stmtIns.Close()
        } else {
            returnErr = dbOpenErr        
        }


    } else {
        returnErr = checkSessionError
    }

    return &merchantServicePB.MerchantOpenResponse{Successed:returnErr == nil},returnErr

}

func (s *roomService) TerminalUnbind(ctx context.Context, in *roomServicePB.RSTerminalUnbindRequest) (*roomServicePB.RSTerminalUnindResponse, error) {
}

func (s *roomService) Create(ctx context.Context, in *roomServicePB.RSCreateRequest) (*roomServicePB.RSCreateResponse, error) {
}

func (s *roomService) Update(ctx context.Context, in *roomServicePB.RSUpdateRequest) (*roomServicePB.RSUpdateResponse, error) {
}

func (s *roomService) Delete(ctx context.Context, in *roomServicePB.RSDeleteRequest) (*roomServicePB.RSDeleteResponse, error) {
}

func (s *roomService) Query(ctx context.Context, in *roomServicePB.RSQueryRequest) (*roomServicePB.RSQueryResponse, error) {
}

func (s *merchantService) MerchantOpen(ctx context.Context, in *merchantServicePB.MerchantOpenRequest) (*merchantServicePB.MerchantOpenResponse, error) {

    merchantName := in.Name
    cellphone := in.Cellphone
    address := in.Address
    paymentBit := in.PaymentBit

    var returnErr error = nil

    db, dbOpenErr := sql.Open("mysql", mysqlDSN)
    defer db.Close()
    // Open doesn't open a connection. Validate DSN data:
    dbOpenErr = db.Ping()
    if (dbOpenErr == nil) {
        stmtIns, stmtInsErr := db.Prepare("INSERT INTO merchant (`merchant_name`,`merchant_address`,`payment_type`,`cellphone`) VALUES( ?, ?, ?, ?)") // ? = placeholder
        if stmtInsErr == nil {
            insertResult, insertErr := stmtIns.Exec(merchantName, address, paymentBit, cellphone) 
            if insertErr == nil {
                afftectedRow, afftectedRowErr := insertResult.RowsAffected()
                if afftectedRow != 1 || afftectedRowErr == nil {
                    returnErr := afftectedRowErr
                }
            } else {
                returnErr = insertErr
            }
        } else {
            returnErr = stmtInsErr
        }
        defer stmtIns.Close()
    } else {
        returnErr = dbOpenErr        
    }
    return &merchantServicePB.MerchantOpenResponse{Successed:returnErr == nil},returnErr
}

func (s *merchantService) MerchantRegister(ctx context.Context, in *merchantServicePB.MerchantRegisterRequest) (*merchantServicePB.MerchantRegisterResponse, error) {
    cellphone, psw, err:= parseUsernameAndPsw(in.Key)
    name := in.Name
    role := in.Role
    verifyCode := in.VerifyCode
    merchantID := in.MerchantID

    var operatorSuccess bool = false
    var returnErr error = nil

    //step1 验证码检查
    conn,redisRPCError := grpc.Dial(redisRPCAddress,grpc.WithInsecure())
    if redisRPCError == nil {
        c := redisPb.NewRedisServiceClient(conn)
        ctxRPC, cancel := context.WithTimeout(context.Background(), time.Second)
        defer cancel() 
        checkKey := "_verify_sms_"+ cellphone + "@1" 
        queryRes, queryRedisErr := c.Query(ctxRPC, &redisPb.QueryRedisRequest{Key: checkKey})

        hasError := queryRedisErr != nil

        if !hasError && queryRes!= nil && queryRes.Value ==  verifyCode {

            //插入数据库
            db, dbOpenErr := sql.Open("mysql", mysqlDSN)
            defer db.Close()
            dbOpenErr = db.Ping()
            if dbOpenErr == nil {
                stmtIns, stmtInsErr := db.Prepare("INSERT INTO `merchant_account` (`name`,`cellphone`,`psw`,`role`,`merchant_id`)VALUES(?,?,?,?,?)") // ? = placeholder
                if stmtInsErr == nil {

                    insertResult, insertErr := stmtIns.Exec(name, cellphone, pswEncryption(psw), role,merchantID) 

                    if insertErr == nil {
                        afftectedRow, afftectedRowErr := insertResult.RowsAffected()
                        if afftectedRow != 1 || afftectedRowErr == nil {
                            returnErr := afftectedRowErr
                        } else {
                            //成功
                            operatorSuccess = true
                        }
                    } else {
                        returnErr = insertErr
                    }
                } else {

                    returnErr = stmtInsErr
                }
                defer stmtIns.Close()
            } else {
                returnErr = dbOpenErr
            }
        } else {
           returnErr = errors.New("验证码验证错误")
        }
    } else {
        returnErr = redisRPCError
    }
    defer conn.Close()
    return &merchantServicePB.MerchantRegisterResponse{Successed:operatorSuccess},returnErr
}

func (s *merchantService) MerchantChangePsw(ctx context.Context, in *merchantServicePB.MerchantChangePswRequest) (*merchantServicePB.MerchantChangePswResponse, error) {

    newPsw := in.NewPsw
    sessionToken := in.SessionToken
    verifyCode := in.VerifyCode
    cellphone := in.Cellphone
    scene := 2
    var operatorSuccess bool = false
    var returnErr error = nil
    //检查session 是否合法

    sessionRPCConn,sessionRPCErr := grpc.Dial(sessionRPCAddress,grpc.WithInsecure())
    if sessionRPCErr == nil {
        sessionRPCConnClient := sessionServicePB.NewSessionServiceClient(sessionRPCConn)

        sessionRPCConnClientCtx, sessionRPCConnClientCancel := context.WithTimeout(context.Background(), time.Second)

        defer sessionRPCConnClientCancel()
        var isSessionTokenOK bool = false
        isSessionTokenVailate, isSessionTokenVailateErr := sessionRPCConnClient.IsSessionTokenVailate(sessionRPCConnClientCtx, &sessionServicePB.IsSessionTokenVailateRequest{SessionToken:sessionToken})
        if isSessionTokenVailateErr == nil {
            isSessionTokenOK := isSessionTokenVailate.Successed

            if isSessionTokenOK {
                encyptRPCConn,encyptRPCErr := grpc.Dial(encyptRPCAddress,grpc.WithInsecure())
                var operatorSuccess bool = false
                var returnErr error = nil
                if encyptRPCErr == nil {
                    encyptRPCClient := encryptionServicePB.NewEncryptionServiceClient(encyptRPCConn)
                    encyptRPCClientCtx, encyptRPCClientCancel := context.WithTimeout(context.Background(), time.Second)
                    defer encyptRPCClientCancel()
                    base64DecodeRes, base64DecodeErr := c.Base64Decode(ctx,&encryptionServicePB.EncryptionBase64DecodeRequest{DecodedStr:newPsw})
                    if base64DecodeErr == nil {
                        rawData, RSADecryptionErr := c.RsaDecryption(ctx, &encryptionServicePB.DecryptionRSARequest{EncryptedStr:base64DecodeRes.RawValue})
                        if RSADecryptionErr == nil {
                            newPsw := string(rawData.RawValue)
                            if nil != newPsw {
                                    //获取到最新的密码 验证 验证码是否正常
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
                                                    afftectedRow, afftectedRowErr := insertResult.RowsAffected()
                                                    if afftectedRow != 1 || afftectedRowErr == nil {
                                                        returnErr := afftectedRowErr
                                                    } else {
                                                        //成功
                                                        operatorSuccess = true
                                                    }
                                                } else {
                                                    returnErr = insertErr
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
                                returnErr := errors.New("解析密码错误")
                            }
                        } else {
                           returnErr = RSADecryptionErr
                        }
                    } else {
                        returnErr = base64DecodeErr
                    }
                } else {
                    returnErr = encyptRPCErr
                }
                defer encyptRPCConn.Close()

            } else {
                returnErr = errors.New("session expired, pls login again")
            }
        } else {
            returnErr = isSessionTokenVailateErr
        }
    } else {
        returnErr = sessionRPCErr
    }
    defer sessionRPCConn.Close()

    return &merchantServicePB.MerchantChangePswResponse{Successed:operatorSuccess},returnErr
}

func (s *merchantService) MerchantLogin(ctx context.Context, in *merchantServicePB.MerchantLoginRequest) (*merchantServicePB.MerchantLoginResponse, error) {
    cellphone, psw, err:= parseUsernameAndPsw(in.Key)
    tokenRequest := in.TokenRequest
    merchantID := in.MerchantID
    var operatorSuccess bool = false
    var returnErr error = nil
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
                    sessionToken := createSessionTokenRes.SessionToken

                } else {
                    returnErr = err
                }
            } else {
                returnErr = sessionRPCErr
            }
            defer sessionRPCConn.Close()
        } else if queryRowErr == sql.ErrNoRows {
            //没有记录
            returnErr = errors.New("密码或者用户名错误")
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
    uid, merchantID, sessionTokenError := fetchSessionTokenValue(in.SessionToken)
    var returnErr error = nil
    if sessionTokenError == nil {
        var (
            merchantName string
            name string
            role int
        )
        db, dbOpenErr := sql.Open("mysql", mysqlDSN)
        defer db.Close()
        dbOpenErr = db.Ping()
        if dbOpenErr == nil {
            queryRowErr := db.QueryRow("SELECT m.merchant_name,ma.role,ma.name FROM merchant_account AS ma, merchant AS m WHERE ma.merchant_id = ? and ma.ID = ? limit 1;", merchantID, uid).Scan(&merchantName,,&role,&name)
            if queryRowErr == nil {
                //制作 令牌
            } else if queryRowErr == sql.ErrNoRows {
                //没有记录
                returnErr = errors.New("没用商户信息")
            } else {
                returnErr = queryRowErr
            }

        } else {
            returnErr = dbOpenErr
        }
    } else {
        returnErr = sessionTokenError
    }

    return &merchantServicePB.MerchantInfoResponse{MerchantName:merchantName,&merchantServicePB.InnerMerchantAccount{Role:uint32(role),Name:name}},returnErr
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
    merchantServicePB.RegisterMerchantServiceServer(s, new(merchantService))
    err = s.Serve(lis)
    if err != nil {
        log.Fatal(err)
    }
}
