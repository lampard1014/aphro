package RoomService

import (
    "golang.org/x/net/context"
    "github.com/lampard1014/aphro/Biz/room/pb"
    "github.com/lampard1014/aphro/CommonBiz/Response/PB"
    "github.com/lampard1014/aphro/CommonBiz/Session"
    "github.com/lampard1014/aphro/PersistentStore/MySQL"
    "github.com/lampard1014/aphro/CommonBiz/Response"
    "github.com/lampard1014/aphro/Gateway/error"
    "strconv"
    "strings"
    "time"
)

//const (
//	port  = ":10090"
//)


const (
    //房间状态 可用
    RoomStatusEnable = 0
    //房间状态 使用中
    RoomStatusInUse = 1
    //房间状态 不可用
    RoomStatusDisable = 2

    RCRIDDelimiter  = "&"

    RoomTransactionBegin = 0
    RoomTransactionSuspend = 1
    RoomTransactionEnd = 2

)

type RoomServiceImp struct{}

func (s *RoomServiceImp) TerminalBind(ctx context.Context, in *Aphro_Room_pb.RSTerminalBindRequest) (*Aphro_CommonBiz.Response, error) {
    sessionToken := in.SessionToken
    terminalCode := in.TerminalCode
    location := in.Location
    latitude := location.Latitude
    longitude := location.Longitude
    roomID := in.RoomID

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    isVaild, checkSessionError := Session.IsSessionTokenVailate(sessionToken)
    if isVaild {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "UPDATE `merchant_room` SET terminal_code =? AND longitude = ? AND latitude = ? AND status = " + strconv.Itoa(RoomStatusEnable) + " where roomID = ?"

                _,err := m.Query(querySQL,terminalCode,latitude,longitude,roomID).RowsAffected()
                if err == nil {
                    //制作 令牌
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTerminalBindResponse{Success:true})
                } else {
                    returnErr = err
                }
            } else {
                returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        } else {
            returnErr = err
        }

    } else {
        returnErr = checkSessionError
    }
    return res,returnErr
}

func (s *RoomServiceImp) TerminalUnbind(ctx context.Context, in *Aphro_Room_pb.RSTerminalUnbindRequest) (*Aphro_CommonBiz.Response, error) {
    sessionToken := in.SessionToken
    roomID := in.RoomID

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    isVaild, checkSessionError := Session.IsSessionTokenVailate(sessionToken)
    if isVaild {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "UPDATE `merchant_room` SET terminal_code =\"\" AND status = " + strconv.Itoa(RoomStatusDisable) + " where roomID = ?"
                _,err := m.Query(querySQL,roomID).RowsAffected()
                if err == nil {
                    //制作 令牌
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTerminalUnbindResponse{Success:true})
                } else {
                    returnErr = err
                }
            } else {
                returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
            }
        } else {
            returnErr = err
        }

    } else {
        returnErr = checkSessionError
    }

    return res,returnErr
}

func (s *RoomServiceImp) CreateRoom(ctx context.Context, in *Aphro_Room_pb.RSCreateRequest) (*Aphro_CommonBiz.Response, error) {
    sessionToken := in.SessionToken
    terminalCode := in.TerminalCode
    location := in.Location
    latitude := location.Latitude
    longitude := location.Longitude
    roomName := in.RoomName

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
     _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "INSERT INTO `merchant_room` (`merchant_id`,`longitude`,`latitude`,`room_name`,`status`,`terminal_code`) VALUES (?,?,?,?,?,?)"
                _,err := m.Query(querySQL,merchantID,longitude,latitude,roomName,RoomStatusDisable,terminalCode).RowsAffected()
                if err == nil {
                    //制作 令牌

                    querySQL := "SELECT `ID` FROM `merchant_room` WHERE `merchant_id` = ? AND `longitude` = ? AND `latitude` =? AND `room_name` = ? AND `status` = ? AND `terminal_code`= ? LIMIT 1"
                    var roomID uint32
                    err := m.Query(querySQL,merchantID,longitude,latitude,roomName,RoomStatusDisable,terminalCode).FetchRow(&roomID)
                    if err == nil {
                        res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSCreateResponse{Success:true,RoomID:roomID})
                    } else {
                        returnErr = err
                    }
                } else {
                    returnErr = err
                }
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

func (s *RoomServiceImp) UpdateRoom(ctx context.Context, in *Aphro_Room_pb.RSUpdateRequest) (*Aphro_CommonBiz.Response, error) {

    sessionToken := in.SessionToken
    terminalCode := in.TerminalCode
    location := in.Location
    latitude := location.Latitude
    longitude := location.Longitude
    roomName := in.RoomName
    roomId := in.RoomID
    chargeRules := in.ChargeRules
    status := in.Status

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                
                var insertData []string
                for _,cr := range chargeRules {
                    r,err := s.CreateRoomChargeRule(ctx,cr)
                    if err != nil  {
                        return res,nil
                    } else {
                        var crcCreateResponse *Aphro_Room_pb.RCRCreateResponse
                        err := Response.UnmarshalAny(r.Result,crcCreateResponse)
                        if err == nil {
                            insertData = append(insertData,strconv.Itoa(int(crcCreateResponse.RecodeID)))
                        } else {
                            return res,err
                        }
                    }
                }
                if len(insertData) > 0 {
                    _,err := m.Update("merchant_room",map[string]interface{}{
                        "merchant_id":"?",
                        "longitude":"?",
                        "latitude":"?",
                        "room_name":"?",
                        "status":"?",
                        "terminal_code":"?",
                        "charge_rules":"?",
                    }).Where(&MySQL.APSMySQLCondition{MySQL.APSMySQLOperator_Equal,"ID","?"}).Execute(merchantID,longitude,latitude,roomName,status,terminalCode,strings.Join(insertData,RCRIDDelimiter),roomId).RowsAffected()
                    if err == nil {
                        res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSUpdateResponse{Success:true})
                    } else {
                        returnErr = err
                    }
                } else {
                    returnErr = AphroError.New(AphroError.BizError,"计费模式不能为空")
                }
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

func (s *RoomServiceImp) DeleteRoom(ctx context.Context, in *Aphro_Room_pb.RSDeleteRequest) (*Aphro_CommonBiz.Response, error) {
    sessionToken := in.SessionToken
    roomID := in.RoomID

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _, sessionTokenError := Session.IsSessionTokenVailate(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "DELETE FROM `merchant_room` WHERE `ID`= ? LIMIT 1"
                _,err := m.Query(querySQL,roomID).RowsAffected()
                if err == nil {
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSDeleteResponse{Success:true})
                } else {
                    returnErr = err
                }
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

func (s *RoomServiceImp) QueryRoom(ctx context.Context, in *Aphro_Room_pb.RSQueryRequest) (*Aphro_CommonBiz.Response, error) {
    sessionToken := in.SessionToken
    roomId := in.RoomID

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                var whereCondition string = "1 "
                if roomId == 0 {
                    whereCondition = " `merchant_id` =  " + merchantID
                } else {
                    whereCondition = " `ID` = ? AND `merchantID` =  " + merchantID
                }
                querySQL := "SELECT `ID`,`merchant_id`,`longitude`,`latitude`,`room_name`,`status`,`charge_rules`,`terminal_code` FROM `merchant_room` WHERE " + whereCondition
                var (
                    roomID uint32
                    longitude string
                    latitude string
                    room_name string
                    status int
                    charge_rules string
                    terminal_code string
                    )

                qr := &Aphro_Room_pb.RSQueryResponse{}
                //var rooms []*Aphro_Room_pb.RSResult
                err := m.Query(querySQL,roomId).FetchAll(func(dest...interface{}){
                    mid,err := strconv.Atoi(merchantID)
                    if err != nil {
                        return
                    }
                    var ruleList []*Aphro_Room_pb.RCRResult;
                    rsResult := &Aphro_Room_pb.RSResult{RoomID:roomID,MerchantID:uint32(mid),TerminalCode:terminal_code,Location:&Aphro_Room_pb.RSLocation{Longitude:longitude,Latitude:latitude},Status:uint32(status),RoomName:room_name,ChargeRules:ruleList}
                    qr.Rooms = append(qr.Rooms, rsResult)
                    charge_rule := strings.Split(charge_rules,RCRIDDelimiter)
                    for _,rcrid := range charge_rule {
                        //获取rcr。。
                        i,err := strconv.Atoi(rcrid)
                        if err != nil {
                            return
                        }
                        rcrRequest := &Aphro_Room_pb.RCRQueryRequest{RCRID:uint32(i),SessionToken:sessionToken,MerchantID:uint32(mid),RoomID:roomID}
                        rcrResponse ,err := s.QueryRoomChargeRule(ctx,rcrRequest)
                        if err != nil {
                            return
                        } else {
                            var r *Aphro_Room_pb.RCRQueryResponse
                            err := Response.UnmarshalAny(rcrResponse.Result,r)
                            if err != nil || !r.Success{
                                return
                            } else {
                                ruleList = r.Results
                            }
                        }
                    }
                },&roomID,&longitude,&latitude,&room_name,&status,&charge_rules,&terminal_code)
                if err == nil {
                    qr.Success = true
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),qr)
                } else {
                    returnErr = err
                }
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

func (s *RoomServiceImp) CreateRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRCreateRequest) (*Aphro_CommonBiz.Response, error) {

    fee := in.Fee
    start := in.Start
    end := in.End
    interval := in.Interval
    intervalUnit := in.IntervalUnit
    roomId := in.RoomID
    sessionToken := in.SessionToken

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "INSERT INTO `merchant_charge_rule` (`fee_per`,`start`,`end`,`interval`,`interval_unit`,`merchant_id`,`room_id`,`flag`) VALUES (?,?,?,?,?,?,?,?)"
                _,err := m.Query(querySQL,fee,start,end,interval,intervalUnit,merchantID,roomId,"").RowsAffected()
                if err == nil {
                    querySQL := "SELECT `ID` FROM `merchant_charge_rule` WHERE `fee_per` = ? AND `start` = ? AND `end` =? AND `interval` = ? AND `interval_unit` = ? AND `merchant_id`= ? AND `room_id` = ? AND `flag`=? LIMIT 1"
                    var ID uint32
                    err := m.Query(querySQL,fee,start,end,interval,intervalUnit,merchantID,roomId,"").FetchRow(&ID)
                    if err == nil {
                        res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RCRCreateResponse{Success:true,RecodeID:ID})
                    } else {
                        returnErr = err
                    }
                } else {
                    returnErr = err
                }
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

func (s *RoomServiceImp) UpdateRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRUpdateRequest) (*Aphro_CommonBiz.Response, error) {
    fee := in.Fee
    start := in.Start
    end := in.End
    interval := in.Interval
    intervalUnit := in.IntervalUnit
    roomId := in.RoomID
    sessionToken := in.SessionToken
    rcrid := in.RCRID

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "UPDATE `merchant_charge_rule` SET `fee_per` = ? AND `start` = ? AND `end` = ? AND `interval` = ? AND `interval_unit` = ? AND `merchant_id` = ? AND `room_id` = ? WHERE `ID` = ?"
                _,err := m.Query(querySQL,fee,start,end,interval,intervalUnit,merchantID,roomId,rcrid).RowsAffected()
                if err == nil {
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RCRUpdateResponse{Success:true})
                } else {
                    returnErr = err
                }
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

func (s *RoomServiceImp) QueryRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRQueryRequest) (*Aphro_CommonBiz.Response, error) {

    sessionToken := in.SessionToken
    roomId := in.RoomID
    inMerchantID := in.MerchantID
    rcrID := in.RCRID
    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                var whereCondition []string = []string{}
                var binds []interface{} = []interface{}{}
                if roomId != 0 {
                    whereCondition = append(whereCondition, "`room_id` =  ?")
                    binds = append(binds,roomId)
                }
                if inMerchantID != 0 {
                    whereCondition = append(whereCondition, "`merchant_id` =  ?")
                    binds = append(binds,inMerchantID)

                } else {
                    whereCondition = append(whereCondition,"`merchant_id` =  ?")
                    binds = append(binds,merchantID)
                }
                if rcrID != 0 {
                    whereCondition = append(whereCondition, "`ID` =  ?")
                    binds = append(binds,rcrID)
                }
                querySQL := "SELECT `ID`,`fee_per`,`start`,`end`,`interval`,`interval_unit`,`merchant_id`,`room_id`,`flag` FROM `merchant_charge_rule` WHERE " + strings.Join(whereCondition," AND ")
                var (
                    r_ID uint32
                    r_fee float32
                    r_start string
                    r_end string
                    r_interval int
                    r_interval_unit int
                    r_merchant_id uint32
                    r_room_id uint32
                    r_flag int
                )

                qr := &Aphro_Room_pb.RCRQueryResponse{}
                //var rooms []*Aphro_Room_pb.RSResult
                err := m.Query(querySQL,binds...).FetchAll(func(dest...interface{}){
                    if err != nil {
                        return
                    }
                    rsResult := &Aphro_Room_pb.RCRResult{
                        MerchantID:r_merchant_id,
                        RCRID:r_ID,
                        Fee:r_fee,
                        Start:r_start,
                        End:r_end,
                        Interval:uint32(r_interval),
                        IntervalUnit:uint32(r_interval_unit),
                        RoomID:r_room_id,
                        Flag:uint32(r_flag),
                    }
                    qr.Results = append(qr.Results, rsResult)
                },&r_ID,&r_fee,&r_start,&r_end,&r_interval,&r_interval_unit,&r_merchant_id,&r_room_id,&r_flag)

                if err == nil {
                    qr.Success = true
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),qr)
                } else {
                    returnErr = err
                }
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

func (s *RoomServiceImp) DeleteRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRDeleteRequest) (*Aphro_CommonBiz.Response, error) {

    sessionToken := in.SessionToken

    roomId := in.RoomID
    inMerchantID := in.MerchantID
    rcrID := in.RCRID

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {


                var whereCondition []string = []string{}
                var binds []interface{} = []interface{}{}
                if roomId != 0 {
                    whereCondition = append(whereCondition, "`room_id` =  ?")
                    binds = append(binds,roomId)
                }
                if inMerchantID != 0 {
                    whereCondition = append(whereCondition, "`merchant_id` =  ?")
                    binds = append(binds,inMerchantID)

                } else {
                    whereCondition = append(whereCondition,"`merchant_id` =  ?")
                    binds = append(binds,merchantID)
                }
                if rcrID != 0 {
                    whereCondition = append(whereCondition, "`ID` =  ?")
                    binds = append(binds,rcrID)
                }

                querySQL := "DELETE FROM `merchant_charge_rule` WHERE  " + strings.Join(whereCondition, " AND ")
                _,err := m.Query(querySQL,binds...).RowsAffected()
                if err == nil {
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RCRDeleteResponse{Success:true})
                } else {
                    returnErr = err
                }
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

func (s *RoomServiceImp) RoomTransactionBegin(ctx context.Context, in *Aphro_Room_pb.RSTransactionBeginRequest) (*Aphro_CommonBiz.Response, error) {
    roomChargeRuleID := in.RoomChargeRuleID
    roomId := in.RoomID
    sessionToken := in.SessionToken

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _,  sessionTokenError := Session.IsSessionTokenVailate(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                //先检查房间是否可用
                querySQL := "SELECT `room_name`,`status`,`terminal_code` FROM `merchant_room` WHERE `ID` = ? LIMIT 1"
                var (
                    room_name string
                    status int
                    terminal_code string
                )
                err := m.Query(querySQL,roomId).FetchRow(&room_name,&status,&terminal_code)
                if err == nil && status == RoomStatusEnable {
                    //新增一个事务
                    querySQL := "INSERT  INTO `transaction_room` (`room_id`,`room_name`,`start_time`,`update_time`,`status`,`terminal_code`) VALUES (?,?,?,?,?,?)"
                    startTime := time.Now()
                    transactionId,err := m.Query(querySQL,roomId,room_name,startTime,startTime,RoomTransactionBegin,terminal_code).LastInsertId()
                    if err == nil {

                        rcrRequest := &Aphro_Room_pb.RCRQueryRequest{}
                        rcrRequest.RCRID = roomChargeRuleID
                        rcrRequest.SessionToken = sessionToken
                        rcr,err := s.QueryRoomChargeRule(ctx,rcrRequest)
                        if err == nil {
                            var rcrRes *Aphro_Room_pb.RCRQueryResponse
                            err := Response.UnmarshalAny(rcr.Result,rcrRes)
                            if err == nil && rcrRes.Success{
                                rcr := rcrRes.Results[0]
                                _,err := s.RoomTransactionCreateRoomFee(ctx,&Aphro_Room_pb.RSTransactionCreateRoomFeeRequest{
                                    SessionToken:sessionToken,
                                    Fee:rcr.Fee,
                                    Start:rcr.Start,
                                    End:rcr.End,
                                    IntervalUnit:rcr.IntervalUnit,
                                    Interval:rcr.Interval,
                                    MerchantID:rcr.MerchantID,
                                    RoomID:roomId,
                                    TransactionID:uint32(transactionId),
                                    Flag:0,
                                })
                                if err == nil {
                                    querySQL := "UPDATE `merchant_room` SET `status` = ?  WHERE `ID` = ? LIMIT 1"
                                    _,err := m.Query(querySQL,RoomStatusInUse,roomId).RowsAffected()
                                    if err == nil {
                                        res ,err := Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTransactionBeginResponse{Success:true,TransactionID:uint32(transactionId)})
                                        return res,err
                                    } else {
                                        return res,err
                                    }
                                } else {
                                    return res,err
                                }

                            } else {
                                return res, err
                            }
                        } else {
                            return res,err
                        }

                    } else {
                        return res,err
                    }
                } else if status != RoomStatusEnable {
                    return res, AphroError.New(AphroError.BizError,"房间不可用")
                } else {
                    return res,err
                }

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
// 挂起一个房间的事务
func (s *RoomServiceImp) RoomTransactionSuspend(ctx context.Context, in *Aphro_Room_pb.RSTransactionSuspendRequest) (*Aphro_CommonBiz.Response, error){
    transactionID := in.TransactionID
    roomId := in.RoomID
    sessionToken := in.SessionToken
    //merchantID := in.MerchantID

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _,  sessionTokenError := Session.IsSessionTokenVailate(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                //先检查房间是否可用
                querySQL := "SELECT `room_name`,`status`,`terminal_code` FROM `merchant_room` WHERE `ID` = ? LIMIT 1"
                var (
                    room_name string
                    status int
                    terminal_code string
                )
                err := m.Query(querySQL,roomId).FetchRow(&room_name,&status,&terminal_code)
                if err == nil && status == RoomStatusInUse {
                    //更新一个事务
                    var whereCondition []string = []string{}
                    var binds []interface{} = []interface{}{}
                    startTime := time.Now()
                    binds = append(binds,startTime,RoomTransactionSuspend)
                    if roomId != 0 {
                        whereCondition = append(whereCondition, "`room_id` =  ?")
                        binds = append(binds,roomId)
                    }

                    if transactionID != 0 {
                        whereCondition = append(whereCondition, "`transaction_id` =  ?")
                        binds = append(binds,transactionID)
                    }

                    querySQL := "UPDATE `transaction_room` SET `update_time` = ? AND `status` = ? WHERE " + strings.Join(whereCondition," AND ")
                    _,err := m.Query(querySQL,binds...).RowsAffected()
                    if err == nil {
                        res ,err := Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTransactionSuspendResponse{Success:true})
                        return res ,err
                    } else {
                        return res,nil
                    }
                } else if err == nil {
                    return res, AphroError.New(AphroError.BizError,"房间不可用")
                } else {
                    return res,err
                }

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

// 结束一个房间的事务
func (s *RoomServiceImp) RoomTransactionEnd(ctx context.Context, in *Aphro_Room_pb.RSTransactionEndRequest) (*Aphro_CommonBiz.Response, error) {
    transactionID := in.TransactionID
    roomId := in.RoomID
    sessionToken := in.SessionToken
    //merchantID := in.MerchantID

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _,  sessionTokenError := Session.IsSessionTokenVailate(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                //先检查房间是否可用
                querySQL := "SELECT `room_name`,`status`,`terminal_code` FROM `merchant_room` WHERE `ID` = ? LIMIT 1"
                var (
                    room_name string
                    status int
                    terminal_code string
                )
                err := m.Query(querySQL,roomId).FetchRow(&room_name,&status,&terminal_code)
                if err == nil && status == RoomStatusInUse {
                    //更新一个事务
                    var whereCondition []string = []string{}
                    var binds []interface{} = []interface{}{}
                    startTime := time.Now()
                    binds = append(binds,startTime,RoomTransactionEnd)
                    if roomId != 0 {
                        whereCondition = append(whereCondition, "`room_id` =  ?")
                        binds = append(binds,roomId)
                    }

                    if transactionID != 0 {
                        whereCondition = append(whereCondition, "`transaction_id` =  ?")
                        binds = append(binds,transactionID)
                    }

                    querySQL := "UPDATE `transaction_room` SET `update_time` = ? AND `status` = ? WHERE " + strings.Join(whereCondition," AND ")
                    _,err := m.Query(querySQL,binds...).RowsAffected()
                    if err == nil {
                        querySQL := "UPDATE `merchant_room` SET `status` = ?  WHERE `ID` = ? LIMIT 1"
                        _,err := m.Query(querySQL,RoomStatusEnable,roomId).RowsAffected()
                        if err == nil {
                            res ,err := Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTransactionSuspendResponse{Success:true})
                            return res ,err
                        }  else {
                            return res, err
                        }
                    } else {
                        return res,nil
                    }
                } else if err == nil {
                    return res, AphroError.New(AphroError.BizError,"房间不可用")
                } else {
                    return res,err
                }
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
// 创建一个房间的房费
func (s *RoomServiceImp) RoomTransactionCreateRoomFee(ctx context.Context, in *Aphro_Room_pb.RSTransactionCreateRoomFeeRequest) (*Aphro_CommonBiz.Response, error) {

    sessionToken := in.SessionToken
    fee := in.Fee
    start := in.Start
    end := in.End
    interval := in.Interval
    intervalUnit := in.IntervalUnit
    merchant_id := in.MerchantID
    roomID := in.RoomID
    transactionID := in.TransactionID
    flag := in.Flag
    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _,  sessionTokenError := Session.IsSessionTokenVailate(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                startTime := time.Now()
                querySQL := "INSERT  INTO `transaction_roomfee` (`fee`,`create_time`,`fee_per_interval`,`start`,`end`,`interval`,`interval_unit`,`merchant_id`,`room_id`,`transaction_id`,`flag`) VALUES (?,?,?,?,?,?,?,?,?,?,?,?)"
                _,err := m.Query(querySQL,
                    0,
                    startTime,
                    fee,
                    start,
                    end,
                    interval,
                    intervalUnit,
                    merchant_id,
                    roomID,
                    transactionID,
                    flag).LastInsertId()
                if err == nil {
                    res ,err := Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTransactionCreateRoomFeeResponse{Success:true})
                    return res,err
                } else {
                    return res,err
                }
            }
        } else {
            returnErr = AphroError.New(AphroError.BizError,"mysql类型断言错误")
        }
    } else {
        returnErr = sessionTokenError
    }
    return res,returnErr
}
// 查询一个房间的房费
func (s *RoomServiceImp) RoomTransactionQueryRoomFee(ctx context.Context, in *Aphro_Room_pb.RSTransactionQueryRoomFeeRequest) (*Aphro_CommonBiz.Response, error) {
    transactionID := in.TransactionID
    roomId := in.RoomID
    sessionToken := in.SessionToken
    transactionRoomFeeID := in.TransactionRoomFeeID
    //merchantID := in.MerchantID
    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _,  sessionTokenError := Session.IsSessionTokenVailate(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {

                //更新一个事务
                var whereCondition []string = []string{}
                var binds []interface{} = []interface{}{}
                if transactionRoomFeeID != 0 {
                    whereCondition = append(whereCondition, "`ID` =  ?")
                    binds = append(binds,transactionRoomFeeID)
                }
                if roomId != 0 {
                    whereCondition = append(whereCondition, "`room_id` =  ?")
                    binds = append(binds,roomId)
                }
                if transactionID != 0 {
                    whereCondition = append(whereCondition, "`transaction_id` =  ?")
                    binds = append(binds,transactionID)
                }
                querySQL := "SELECT `ID`,`fee`,`create_time`,`end_time`,`fee_per_interval`,`start`,`end`, `interval`,`interval_unit`,`merchant_id`,`room_id`,`transaction_id`,`flag` FROM `transaction_roomfee` WHERE " + strings.Join(whereCondition," AND ")
                var (
                    ID uint32
                    fee float32
                    create_time uint32
                    end_time uint32
                    fee_per_interval float32
                    start string
                    end string
                    interval uint32
                    interval_unit uint32
                    merchant_id uint32
                    room_id uint32
                    transaction_id uint32
                    flag uint32
                )

                var roomFeeResultList []*Aphro_Room_pb.RSTransactionRoomFeeResult;

                err := m.Query(querySQL,binds...).FetchAll(func(dest...interface{}){
                    t := &Aphro_Room_pb.RSTransactionRoomFeeResult{
                        ID,fee,create_time,end_time,fee_per_interval,
                        start,end,interval,interval_unit,merchant_id,
                        room_id,transaction_id,flag,}
                    roomFeeResultList = append(roomFeeResultList,t)
                },&ID,&fee,&create_time,&end_time,&fee_per_interval,&start,&end,&interval,&interval_unit,&merchant_id,&room_id,&transaction_id,&flag)
                if err == nil {
                    res ,err := Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTransactionQueryRoomFeeResponse{Success:true,RoomFees:roomFeeResultList})
                    return res ,err
                }  else {
                    return res, err
                }
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