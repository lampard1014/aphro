package MerchantServiceIMP

import (

	"golang.org/x/net/context"
    "time"
    _ "github.com/go-sql-driver/mysql"
    merchantServicePB "github.com/lampard1014/aphro/Biz/merchant/pb"
    "github.com/lampard1014/aphro/CommonBiz/Response/PB"
    "github.com/lampard1014/aphro/CommonBiz/Response"

    "strconv"
    "github.com/lampard1014/aphro/Encryption"
    "github.com/lampard1014/aphro/PersistentStore/Redis"
    "github.com/lampard1014/aphro/PersistentStore/MySQL"
    "github.com/lampard1014/aphro/CommonBiz/Session"
    "strings"
    "github.com/lampard1014/aphro/Gateway/error"
)

const (
	Port  = ":10089"
    Scene_ChangePsw = 2
    Scene_MerchantRegister = 1
    verifyCodeDuration = 30
)

type MerchantServiceIMP struct{}

func (s *MerchantServiceIMP) MerchantOpen(ctx context.Context, in *merchantServicePB.MerchantOpenRequest) (res *Aphro_CommonBiz.Response,returnErr error) {
    merchantName := in.Name
    cellphone := in.Cellphone
    address := in.Address
    paymentBit := in.PaymentBit

    mysql,returnErr := MySQL.NewAPSMySQL(nil)

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
        returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
    }
    if returnErr != nil {
        res,returnErr = Response.NewCommonBizResponseWithError(returnErr,nil)
    }
    return
}

func (s *MerchantServiceIMP) MerchantRegister(ctx context.Context, in *merchantServicePB.MerchantRegisterRequest) (res *Aphro_CommonBiz.Response, returnErr error) {
    cellphone, psw, returnErr := Encryption.ParseUsernameAndPsw(in.Key)
    name := in.Name
    role := in.Role
    verifyCode := in.VerifyCode
    merchantID := in.MerchantID
    tokenRequest := in.TokenRequest

    //step1 验证码检查
    redis ,returnErr := Redis.NewAPSRedis(nil)
    checkKey := "_verify_sms_"+ cellphone + "@" + strconv.Itoa(Scene_MerchantRegister)
    redis.Connect()
    queryRes, returnErr := redis.Query(checkKey)

    if returnErr == nil  && queryRes ==  verifyCode {
        m,returnErr := MySQL.NewAPSMySQL(nil)
        m,ok := m.Connect().(*MySQL.APSMySQL)
        if ok {
            var mid int64 = int64(merchantID)
            if role == 1 {
                querySQL := "INSERT INTO `merchant` (`merchant_name`,`merchant_address`,`payment_type`,`cellphone`)VALUES(?,?,?,?)"
                mid , returnErr = m.Query(querySQL,"", "", 0, cellphone).LastInsertId()
            }
            if returnErr == nil {
                //成功新建商户 继续 新建操作员，todo 事务回滚
                querySQL := "INSERT INTO `merchant_account` (`name`,`cellphone`,`psw`,`role`,`merchant_id`)VALUES(?,?,?,?,?)"

                lastInsertID , returnErr := m.Query(querySQL,name, cellphone, Encryption.PswEncryption(psw), role,mid).LastInsertId()
                if returnErr == nil && lastInsertID > 0 {
                    res,returnErr = Response.NewCommonBizResponseWithCodeWithError(0,returnErr,&merchantServicePB.MerchantRegisterResponse{true,""})

                    //获取新的session
                    if tokenRequest  != "" {
                        rsa,returnErr := Encryption.RsaEncryption([]byte(strings.Join([]string{cellphone,psw},Encryption.Delimiter_PSW_USERNAME)))
                        if returnErr == nil {
                            requestKey ,returnErr := Encryption.Base64Encode(rsa)
                            if returnErr == nil {
                                merchantLoginResponse,returnErr := s.MerchantLogin(ctx,&merchantServicePB.MerchantLoginRequest{requestKey,tokenRequest,""})
                                if returnErr == nil {
                                    merchantLoginResponseResult := merchantLoginResponse.Result
                                    var loginResp merchantServicePB.MerchantLoginResponse = merchantServicePB.MerchantLoginResponse{}
                                    returnErr := Response.UnmarshalAny(merchantLoginResponseResult,&loginResp)
                                    if returnErr == nil {
                                        res,returnErr = Response.NewCommonBizResponseWithCodeWithError(0,returnErr,&merchantServicePB.MerchantRegisterResponse{true,loginResp.SessionToken})

                                    }
                                }
                            }
                        }
                    }
                }
            }
            defer m.Close()
        } else {
            returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
        }
    } else {
        returnErr = AphroError.New(Response.BizError,"验证码验证错误")
    }
    if returnErr != nil {
        res,returnErr = Response.NewCommonBizResponseWithError(returnErr,nil)
    }
    return
}

func (s *MerchantServiceIMP) MerchantChangePsw(ctx context.Context, in *merchantServicePB.MerchantChangePswRequest) (res *Aphro_CommonBiz.Response, returnErr error) {

    cellphone, newPsw, returnErr := Encryption.ParseUsernameAndPsw(in.Key)
    //sessionToken := in.SessionToken
    verifyCode := in.VerifyCode
    scene := uint32(Scene_ChangePsw)

    //检查session 是否合法
    //isTokenVaildate, err := Session.IsSessionTokenVailate(sessionToken)
    if returnErr == nil {
        if newPsw != "" {
            verifyCodeRes,returnErr := s.MerchantVerifyCode(ctx,&merchantServicePB.MerchantVerifyCodeRequest{Cellphone:cellphone,Scene:scene,SmsCode:verifyCode})
            var vcr *merchantServicePB.MerchantVerifyCodeResponse = &merchantServicePB.MerchantVerifyCodeResponse{}
            returnErr = Response.UnmarshalAny(verifyCodeRes.Result,vcr)
            if returnErr == nil && vcr.Success {
                mysql,returnErr := MySQL.NewAPSMySQL(nil)
                if returnErr == nil {
                    m, ok := mysql.Connect().(*MySQL.APSMySQL)
                    if ok {
                        querySQL := "update `merchant_account` set psw = ? where cellphone = ? limit 1"
                        psw := Encryption.PswEncryption(newPsw)
                        _ , returnErr := m.Query(querySQL,psw,cellphone).RowsAffected()
                        if returnErr == nil  {
                            res,returnErr = Response.NewCommonBizResponseWithCodeWithError(0,returnErr,&merchantServicePB.MerchantChangePswResponse{true})
                        }
                        defer m.Close()
                    } else {
                        returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
                    }
                }
            }
        } else {
            returnErr = AphroError.New(Response.BizError,"解析密码错误")
        }
    }
    if returnErr != nil {
        res,returnErr = Response.NewCommonBizResponseWithError(returnErr,nil)
    }
    return
}

func (s *MerchantServiceIMP) MerchantLogin(ctx context.Context, in *merchantServicePB.MerchantLoginRequest) (res *Aphro_CommonBiz.Response, err error) {
    tokenRequest := in.TokenRequest
    inSessionToken := in.SessionToken
    var (
        uid uint32
        name string
        role int
        merchantID uint32
    )
    if inSessionToken != ""{
        ok,err := Session.IsSessionTokenVailate(inSessionToken)

        if ok {
           _,err = Session.RenewSessionToken(inSessionToken)
        }
        res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&merchantServicePB.MerchantLoginResponse{SessionToken:inSessionToken,Success:ok && err == nil})
    } else {
        cellphone, psw, err:= Encryption.ParseUsernameAndPsw(in.Key)

        //merchantID := in.MerchantID
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "SELECT id as uid, name ,role,merchant_id FROM merchant_account WHERE cellphone = ? AND psw = ? LIMIT 1"
                pswEncryption := Encryption.PswEncryption(psw)
                err := m.Query(querySQL,cellphone, pswEncryption).FetchRow(&uid,&name,&role,&merchantID)
                if err == nil && uid > 0  {

                    encryptionkey,err := Session.DecodeSessionTokenRequestStr(tokenRequest)
                    if err == nil {
                        token, _, err := Session.CreateSessionToken(encryptionkey,uid,merchantID)
                        if err == nil {
                            encryptionToken,err := Encryption.XxteaEncryption(encryptionkey,token)
                            if err == nil {
                                if base64EncryptionToken , err :=Encryption.Base64Encode(encryptionToken); err == nil {
                                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&merchantServicePB.MerchantLoginResponse{base64EncryptionToken,true})
                                }
                            }
                        }
                    }
                } else {
                    err = AphroError.New(AphroError.BizError,"账号密码错误")
                }
                defer m.Close()
            } else {
                err = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        }
    }
    if err != nil {
        res,err = Response.NewCommonBizResponseWithError(err,nil)
    }
    return
}

func (s *MerchantServiceIMP) MerchantInfo(ctx context.Context, in *merchantServicePB.MerchantInfoRequest) (res *Aphro_CommonBiz.Response, err error) {

    uid, merchantID, err := Session.FetchSessionTokenValue(in.SessionToken)

    var (
        merchantName string
        name string
        role int
    )
    if err == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "SELECT m.merchant_name,ma.role,ma.name FROM merchant_account AS ma, merchant AS m WHERE ma.merchant_id = ? and ma.ID = ? limit 1"
                err := m.Query(querySQL,merchantID, uid).FetchRow(&merchantName,&role,&name)
                if err == nil {
                    //制作 令牌
                    mid ,err := strconv.Atoi(merchantID)
                    if err == nil {
                        res,err = Response.NewCommonBizResponseWithCodeWithError(
                            0,
                            err,
                            &merchantServicePB.MerchantInfoResponse{
                                merchantName,
                                uint32(mid),
                                &merchantServicePB.InnerMerchantAccount{
                                    uint32(role),
                                    name},
                            },
                        )
                    }
                } else if MySQL.ISErrorNoRows(err) {
                    err = AphroError.New(AphroError.BizError,"没有商户信息")
                }
                defer m.Close()
            } else {
                err = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        }
    }
    if err != nil {
        res,err = Response.NewCommonBizResponseWithError(err,nil)
    }
    return
}

func (s *MerchantServiceIMP) MerchantRoomInfo(ctx context.Context, in *merchantServicePB.MerchantRoomInfoRequest) (res *Aphro_CommonBiz.Response,err  error) {
    return nil,nil
}
//  新增商户服务信息
func (s *MerchantServiceIMP) MerchantWaiterCreate(ctx context.Context, in *merchantServicePB.MerchantWaiterCreateRequest) (res *Aphro_CommonBiz.Response, err error) {
    token := in.Token
    merchantID := in.MerchantID
    name := in.Name
    reserve := in.Reserve
    //验证token合法性
    _, merchantID, err = Session.FetchSessionTokenValue(token)

    if err == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "insert into merchant_waiter (`name`,`reserve`,`merchant_id`) values (?,?,?)"

                _,err := m.Query(querySQL,name,reserve,merchantID).LastInsertId()
                if err == nil {
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&merchantServicePB.MerchantWaiterCreateResponse{true})
                }
                defer m.Close()
            } else {
                err = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        }
    }
    if err != nil {
        res,err = Response.NewCommonBizResponseWithError(err,nil)
    }
    return
}

// 删除商户服务信息
func (s *MerchantServiceIMP) MerchantWaiterDelete(ctx context.Context, in *merchantServicePB.MerchantWaiterDeleteRequest) (*Aphro_CommonBiz.Response, error) {
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
                    res,returnErr = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantWaiterDeleteResponse{Success:true})
                } else {
                    returnErr = err
                }
                defer m.Close()
            } else {
                res,returnErr = Response.NewCommonBizResponse(Response.BizError,"mysql类型断言错误",nil)
            }
        } else {
            returnErr = err
        }
    } else {
        returnErr =  sessionTokenError
    }

    return res,returnErr
}

func (s *MerchantServiceIMP) MerchantVerifyCode(ctx context.Context, in *merchantServicePB.MerchantVerifyCodeRequest) (*Aphro_CommonBiz.Response, error) {
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
        isVaildate := strings.ToUpper(value) == strings.ToUpper(smsCode)
        var errMsg string
        if err != nil {
            errMsg = err.Error()
        }
        if errMsg == "" && !isVaildate {
            errMsg = "验证码验证错误"
        }
        res,err = Response.NewCommonBizResponse(0,errMsg,&merchantServicePB.MerchantVerifyCodeResponse{Success:isVaildate})
        return res,err
    }
}

func (s *MerchantServiceIMP) MerchantSendCode(ctx context.Context, in *merchantServicePB.MerchantSendCodeRequest) (*Aphro_CommonBiz.Response, error) {
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
        smsCodeTmp := Encryption.RandNumberBytesMaskImprSrc(4)
        ttl := uint64(verifyCodeDuration * time.Minute)//60秒过期
        success,err :=redis.Set(checkKey,smsCodeTmp,int64(ttl))

        res,err = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantSendCodeResponse{Success:success})
        return res,err
    }
}

func (s *MerchantServiceIMP) MerchantAccountCellphoneUnquie(ctx context.Context, in *merchantServicePB.MerchantAccountCellphoneUnquieReqeuest) (*Aphro_CommonBiz.Response, error) {
    cellphone := in.Cellphone
    role := in.Roles
    var res *Aphro_CommonBiz.Response = nil
    var returnErr error = nil

    mysql,err := MySQL.NewAPSMySQL(nil)
    if err == nil {
        m, ok := mysql.Connect().(*MySQL.APSMySQL)
        if ok {

            var binds []interface{} = []interface{}{}
            binds = append(binds,cellphone)
            var roleCondition []string = []string{}
            for _,d := range role{
                binds = append(binds,d)
                roleCondition = append(roleCondition,"?")
            }

            roleC :="role in (" + strings.Join(roleCondition,",") + ")"

            querySQL := "SELECT 1 FROM `merchant_account` WHERE `cellphone` = ? AND " + roleC
            //fmt.Println(querySQL)
            var bingo uint32 = 0
            err := m.Query(querySQL,binds...).FetchRow(&bingo)
            if err == nil {
                res,returnErr = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantAccountCellphoneUnquieResponse{IsExisted:bingo > 0})
            } else {
                returnErr = err
            }
            defer m.Close()
        } else {
            res,returnErr = Response.NewCommonBizResponse(Response.BizError,"mysql类型断言错误",nil)
        }
    } else {
        returnErr = err
    }
    return res,returnErr
}