package MerchantServiceIMP

import (

	"golang.org/x/net/context"
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
    "github.com/lampard1014/aphro/CommonBiz/Session"
    "strings"
)

const (
	Port  = ":10089"
    Scene_ChangePsw = 2
    Scene_MerchantRegister = 1
    verifyCodeDuration = 60*30
)

type MerchantServiceIMP struct{}



func (s *MerchantServiceIMP) MerchantOpen(ctx context.Context, in *merchantServicePB.MerchantOpenRequest) (*Aphro_CommonBiz.Response, error) {
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

func (s *MerchantServiceIMP) MerchantRegister(ctx context.Context, in *merchantServicePB.MerchantRegisterRequest) (*Aphro_CommonBiz.Response, error) {
    cellphone, psw, err := Encryption.ParseUsernameAndPsw(in.Key)
    name := in.Name
    role := in.Role
    verifyCode := in.VerifyCode
    merchantID := in.MerchantID

    var returnErr error = err
    var res *Aphro_CommonBiz.Response = nil

    //step1 验证码检查
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
            m,err := MySQL.NewAPSMySQL(nil)
            if err != nil {
                returnErr = err
            } else {
                m,ok := m.Connect().(*MySQL.APSMySQL)
                if ok {
                    var mid int64 = int64(merchantID)
                    if role == 1 {
                        querySQL := "INSERT INTO `merchant` (`merchant_name`,`merchant_address`,`payment_type`,`cellphone`)VALUES(?,?,?,?)"
                        mid , returnErr = m.Query(querySQL,"@", "@", 0, cellphone).LastInsertId()
                    }
                    if returnErr == nil {
                        //成功新建商户 继续 新建操作员，todo 事务回滚
                        querySQL := "INSERT INTO `merchant_account` (`name`,`cellphone`,`psw`,`role`,`merchant_id`)VALUES(?,?,?,?,?)"

                        lastInsertID , err := m.Query(querySQL,name, cellphone, Encryption.PswEncryption(psw), role,mid).LastInsertId()
                        returnErr = err
                        if err == nil && lastInsertID > 0 {
                            res,returnErr = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantRegisterResponse{Success:true})
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

func (s *MerchantServiceIMP) MerchantChangePsw(ctx context.Context, in *merchantServicePB.MerchantChangePswRequest) (*Aphro_CommonBiz.Response, error) {

    cellphone, newPsw, err := Encryption.ParseUsernameAndPsw(in.Key)
    sessionToken := in.SessionToken
    verifyCode := in.VerifyCode
    scene := uint32(Scene_ChangePsw)
    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil

    //检查session 是否合法
    isTokenVaildate, err := Session.IsSessionTokenVailate(sessionToken)
    returnErr = err
    if err == nil {
        if isTokenVaildate {
            if newPsw != "" {
                verifyCodeRes,err := s.MerchantVerifyCode(ctx,&merchantServicePB.MerchantVerifyCodeRequest{Cellphone:cellphone,Scene:scene,SmsCode:verifyCode})
                var vcr *merchantServicePB.MerchantVerifyCodeResponse = &merchantServicePB.MerchantVerifyCodeResponse{}
                err = Response.UnmarshalAny(verifyCodeRes.Result,vcr)
                if err == nil && vcr.Success{
                    mysql,err := MySQL.NewAPSMySQL(nil)
                    if err == nil {
                        m, ok := mysql.Connect().(*MySQL.APSMySQL)
                        if ok {
                            querySQL := "update `merchant_account` set psw = ? where cellphone = ? limit 1"

                            _ , err := m.Query(querySQL,Encryption.PswEncryption(newPsw),cellphone).RowsAffected()
                            if err == nil  {
                                res,returnErr = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantChangePswResponse{Success:err == nil})
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
            returnErr = err
        } else {
            returnErr = AphroError.New(AphroError.BizError,"session 过期 请重新登录")
        }
    } else {
        returnErr = err
    }
    return res,returnErr
}

func (s *MerchantServiceIMP) MerchantLogin(ctx context.Context, in *merchantServicePB.MerchantLoginRequest) (*Aphro_CommonBiz.Response, error) {
    tokenRequest := in.TokenRequest
    inSessionToken := in.SessionToken
    var returnErr error
    var (
        uid uint32
        name string
        role int
        merchantID uint32
    )
    var res *Aphro_CommonBiz.Response = nil
    if inSessionToken != ""{
        ok,err := Session.IsSessionTokenVailate(inSessionToken)

        if ok {
           _,returnErr = Session.RenewSessionToken(inSessionToken)
        }
        res,returnErr = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantLoginResponse{SessionToken:inSessionToken,Success:ok && returnErr == nil})
    } else {
        cellphone, psw, err:= Encryption.ParseUsernameAndPsw(in.Key)

        //merchantID := in.MerchantID
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "SELECT id as uid, name ,role,merchant_id FROM merchant_account WHERE cellphone = ? AND psw = ? LIMIT 1"
                err := m.Query(querySQL,cellphone, Encryption.PswEncryption(psw)).FetchRow(&uid,&name,&role,&merchantID)
                if err == nil && uid > 0  {
                    token, _, err := Session.CreateSessionToken(tokenRequest,uid,merchantID)
                    if err == nil {
                        res,returnErr = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantLoginResponse{SessionToken:token,Success:true})
                    }
                    returnErr = err
                } else {
                    returnErr = AphroError.New(AphroError.BizError,"账号密码错误")
                }
                defer m.Close()
            } else {
                returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        } else {
            returnErr = err
        }
    }
    return res,returnErr
}

func (s *MerchantServiceIMP) MerchantInfo(ctx context.Context, in *merchantServicePB.MerchantInfoRequest) (*Aphro_CommonBiz.Response, error) {

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
                    res,returnErr = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantInfoResponse{MerchantName:merchantName,MerchantAccount:&merchantServicePB.InnerMerchantAccount{Role:uint32(role),Name:name}})
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

func (s *MerchantServiceIMP) MerchantRoomInfo(ctx context.Context, in *merchantServicePB.MerchantRoomInfoRequest) (*Aphro_CommonBiz.Response, error) {
    return nil,nil
}
//  新增商户服务信息
func (s *MerchantServiceIMP) MerchantWaiterCreate(ctx context.Context, in *merchantServicePB.MerchantWaiterCreateRequest) (*Aphro_CommonBiz.Response, error) {
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
                    res,returnErr = Response.NewCommonBizResponseWithError(0,err,&merchantServicePB.MerchantWaiterCreateResponse{Success:true})
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
        isVaildate := value == smsCode
        var errMsg string
        if err != nil {
            errMsg = err.Error()
        }
        if errMsg == "" {
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
        smsCodeTmp := "123456"
        ttl := uint64(verifyCodeDuration * time.Second)//60秒过期
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
            returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
        }
    } else {
        returnErr = err
    }
    return res,returnErr
}