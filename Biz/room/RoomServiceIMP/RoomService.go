package RoomServiceIMP

import (
    "golang.org/x/net/context"
    "github.com/lampard1014/aphro/Biz/room/pb"
    "github.com/lampard1014/aphro/CommonBiz/Response/PB"
    "github.com/lampard1014/aphro/CommonBiz/Session"
    "github.com/lampard1014/aphro/PersistentStore/MySQL"
    "github.com/lampard1014/aphro/CommonBiz/Response"
    "github.com/lampard1014/aphro/CommonBiz/Error"
    "strconv"
    "strings"
    "time"
    "github.com/lampard1014/aphro/Biz"
)

const (
    Port  = ":10084"
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

func (s *RoomServiceImp) TerminalBind(ctx context.Context, in *Aphro_Room_pb.RSTerminalBindRequest) (res *Aphro_CommonBiz.Response,err error) {
    sessionToken := in.SessionToken
    terminalCode := in.TerminalCode
    location := in.Location
    latitude := location.Latitude
    longitude := location.Longitude
    roomID := in.RoomID
	//var isVaild bool
    //isVaild, err = Session.IsSessionTokenVailate(sessionToken)
    _,_,err = Session.FetchSessionTokenValue(sessionToken)
    if err == nil {
        var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "UPDATE `merchant_room` SET terminal_code =? , longitude = ? , latitude = ? , status = " + strconv.Itoa(RoomStatusEnable) + " where roomID = ?"

                _,err = m.Query(querySQL,terminalCode,latitude,longitude,roomID).RowsAffected()
                if err == nil {
                    res,err = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTerminalBindResponse{Success:true})
                }
            } else {
                err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

func (s *RoomServiceImp) TerminalUnbind(ctx context.Context, in *Aphro_Room_pb.RSTerminalUnbindRequest) (res *Aphro_CommonBiz.Response,err error) {
    sessionToken := in.SessionToken
    roomID := in.RoomID

    //var isVaild bool
    //isVaild, err = Session.IsSessionTokenVailate(sessionToken)

	_,_,err = Session.FetchSessionTokenValue(sessionToken)

	if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "UPDATE `merchant_room` SET terminal_code =\"\" , status = " + strconv.Itoa(RoomStatusDisable) + " where roomID = ?"
                _,err = m.Query(querySQL,roomID).RowsAffected()
                if err == nil {
                    //制作 令牌
                    res,err = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSTerminalUnbindResponse{Success:true})
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

func (s *RoomServiceImp) CreateRoom(ctx context.Context, in *Aphro_Room_pb.RSCreateRequest) (res *Aphro_CommonBiz.Response,err error) {
    sessionToken := in.SessionToken
    terminalCode := in.TerminalCode
    location := in.Location
    latitude := location.Latitude
    longitude := location.Longitude
    roomName := in.RoomName

    var merchantID string
     _, merchantID, err = Session.FetchSessionTokenValue(sessionToken)
    if err == nil {
		var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "INSERT INTO `merchant_room` (`merchant_id`,`longitude`,`latitude`,`room_name`,`status`,`terminal_code`) VALUES (?,?,?,?,?,?)"
                _,err = m.Query(querySQL,merchantID,longitude,latitude,roomName,RoomStatusDisable,terminalCode).RowsAffected()
                if err == nil {
                    //制作 令牌
                    querySQL := "SELECT `ID` FROM `merchant_room` WHERE `merchant_id` = ? AND `longitude` = ? AND `latitude` =? AND `room_name` = ? AND `status` = ? AND `terminal_code`= ? LIMIT 1"
                    var roomID uint32
                    err = m.Query(querySQL,merchantID,longitude,latitude,roomName,RoomStatusDisable,terminalCode).FetchRow(&roomID)
                    if err == nil {
                        res,err = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSCreateResponse{Success:true,RoomID:roomID})
                    }
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

func (s *RoomServiceImp) UpdateRoom(ctx context.Context, in *Aphro_Room_pb.RSUpdateRequest) (res *Aphro_CommonBiz.Response, err error) {

    var (
        st string
        tc string
        isSetTerminalCode bool
        isSetLocation bool
        rid uint32
        isSetRoomID bool

        status uint32
        isSetStatus bool

        rn string
        isSetRoomName bool

        //isSetChargeRules bool
        )
    if in.SessionToken != nil {
        st = in.SessionToken.Value
    }

    if in.TerminalCode != nil {
        tc = in.TerminalCode.Value
        isSetTerminalCode = true
    }

    location := in.Location
    var latitude string
    var longitude string
    if location != nil {
        latitude = location.Latitude
        longitude = location.Longitude
        isSetLocation = true
    }

    if in.RoomName != nil {
        rid = in.RoomID.Value
        isSetRoomID = true
    }

    if in.RoomName != nil {
        rn = in.RoomName.Value
        isSetRoomName = true
    }

    if in.Status != nil {
        status = in.Status.Value
        isSetStatus = true
    }
	var merchantID string
    _, merchantID, err = Session.FetchSessionTokenValue(st)
    if err == nil {
		var mysql *MySQL.APSMySQL
		mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                var bindValues []interface{} = []interface{}{}
                var updateMaps map[string]interface{} = map[string]interface{}{}

                updateMaps["merchant_id"] = "?"
                bindValues = append(bindValues, merchantID)


                if isSetLocation {
                    updateMaps["longitude"] = "?"
                    bindValues = append(bindValues, longitude)
                    updateMaps["latitude"] = "?"
                    bindValues = append(bindValues, latitude)
                }
                if isSetRoomName {
                    updateMaps["room_name"] = "?"
                    bindValues = append(bindValues, rn)
                }
                if isSetStatus {
                    updateMaps["status"] = "?"
                    bindValues = append(bindValues, status)
                }

                if isSetTerminalCode {
                    updateMaps["staterminal_codetus"] = "?"
                    bindValues = append(bindValues, tc)
                }

                if isSetTerminalCode {
                    updateMaps["staterminal_codetus"] = "?"
                    bindValues = append(bindValues, tc)
                }

                //if isSetChargeRules {
                //    updateMaps["charge_rules"] = "?"
                //    bindValues = append(bindValues, strings.Join(insertData,RCRIDDelimiter))
                //}

                if isSetRoomID {
                    bindValues = append(bindValues, rid)
                } else {
                    bindValues = append(bindValues, 0)
                }

                _,err = m.Update("merchant_room",updateMaps).Where(&MySQL.APSMySQLCondition{MySQL.APSMySQLOperator_Equal,"ID","?"}).Execute(bindValues...).RowsAffected()
                if err == nil {
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RSUpdateResponse{Success:true})
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return

}

func (s *RoomServiceImp) DeleteRoom(ctx context.Context, in *Aphro_Room_pb.RSDeleteRequest) (res *Aphro_CommonBiz.Response,err error) {
    sessionToken := in.SessionToken
    roomID := in.RoomID

    //_, err = Session.IsSessionTokenVailate(sessionToken)
	_,_,err = Session.FetchSessionTokenValue(sessionToken)

	if err == nil {
		var mysql *MySQL.APSMySQL
		mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                querySQL := "DELETE FROM `merchant_room` WHERE `ID`= ? LIMIT 1"
                _,err = m.Query(querySQL,roomID).RowsAffected()
                if err == nil {
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RSDeleteResponse{true})
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

func (s *RoomServiceImp) QueryRoom(ctx context.Context, in *Aphro_Room_pb.RSQueryRequest) (res *Aphro_CommonBiz.Response,err error) {
    sessionToken := in.SessionToken
    roomId := in.RoomID

	var merchantID string
    _, merchantID, err = Session.FetchSessionTokenValue(sessionToken)
    if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
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
                querySQL := "SELECT `ID`,`longitude`,`latitude`,`room_name`,`status`,`charge_rules`,`terminal_code` ,`flag` FROM `merchant_room` WHERE " + whereCondition
                var (
                    roomID uint32
                    longitude string
                    latitude string
                    room_name string
                    status int
                    charge_rules string
                    terminal_code string
                    flag uint32
                    )

                qr := &Aphro_Room_pb.RSQueryResponse{}
                //var rooms []*Aphro_Room_pb.RSResult
                err = m.QueryAll(querySQL).FetchAll(func(dest...interface{}){
                	var mid int
                    mid,err = strconv.Atoi(merchantID)
                    if err != nil {
                        return
                    }
                    var ruleList []*Aphro_Room_pb.RCRResult;

                    charge_rule := strings.Split(charge_rules,RCRIDDelimiter)
                    for _,rcrid := range charge_rule {
                        //获取rcr。。
                        var i int
                        i,err = strconv.Atoi(rcrid)
                        if err != nil {
                            return
                        }
                        rcrRequest := &Aphro_Room_pb.RCRQueryRequest{RCRID:uint32(i),SessionToken:sessionToken,MerchantID:uint32(mid),RoomID:roomID}
                        rcrResponse ,err1 := s.QueryRoomChargeRule(ctx,rcrRequest)
                        err = err1
                        if err != nil {
                            return
                        } else {
                            var r *Aphro_Room_pb.RCRQueryResponse = &Aphro_Room_pb.RCRQueryResponse{}
                            err1 := Response.UnmarshalAny(rcrResponse.Result,r)
                            err = err1
                            if err != nil || !r.Success{
                                return
                            } else {
                                ruleList = r.Results
                            }
                        }
                    }
                    rsResult := &Aphro_Room_pb.RSResult{roomID,uint32(mid),terminal_code,&Aphro_Room_pb.RSLocation{longitude,latitude},uint32(status),room_name,flag,ruleList}
                    qr.Rooms = append(qr.Rooms, rsResult)
                },&roomID,&longitude,&latitude,&room_name,&status,&charge_rules,&terminal_code,&flag)
                if err == nil {
                    qr.Success = true
                    res,err = Response.NewCommonBizResponse(0,"",qr)
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

func (s *RoomServiceImp) BatchCreateRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRBatchCreateRequest) (res *Aphro_CommonBiz.Response,err error) {

    inMerchantID := in.MerchantID
    roomID := in.RoomID
    rules := in.Rules
    sessionToken := in.SessionToken

    _, merchantID, err := Session.FetchSessionTokenValue(sessionToken)

    if inMerchantID != 0{
        merchantID = strconv.Itoa(int(inMerchantID))
    }

    if err == nil {
        var ruleVals []string
        var binds []interface{}
        querySQL := "INSERT INTO `merchant_charge_rule` (`fee_per`,`start`,`end`,`interval`,`interval_unit`,`merchant_id`,`room_id`,`flag`,`name`) VALUES "

        for _,rule := range rules {
            fee := rule.Fee
            start := rule.Start
            end := rule.End
            interval := rule.Interval
            intervalUnit := rule.IntervalUnit
            flag := rule.Flag
            name := rule.Name

            var tmpRule []string = []string{"?","?","?","?","?","?","?","?"}
            ruleVals = append(ruleVals,MySQL.LeftBrackets + strings.Join(tmpRule,MySQL.DelimiterComma) + MySQL.RightBrackets)
            binds = append (binds,fee,start,end,interval,intervalUnit,merchantID,roomID,flag,name)
        }
        var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                _,err = m.Query(querySQL + strings.Join(ruleVals,MySQL.DelimiterComma) ,binds...).RowsAffected()
                if err == nil {
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RCRBatchCreateResponse{true,})
                }
            } else {
                err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
    if err != nil {
        res,err = Response.NewCommonBizResponseWithError(err,nil)
    }
    return

}

func (s *RoomServiceImp) CreateRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRCreateRequest) (res *Aphro_CommonBiz.Response,err error) {
	fee := in.Fee
    start := in.Start
    end := in.End
    interval := in.Interval
    intervalUnit := in.IntervalUnit
    roomId := in.RoomID
    sessionToken := in.SessionToken
    flag := in.Flag
    name := in.Name
    _, merchantID, err := Session.FetchSessionTokenValue(sessionToken)
    if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {

                //var (
                //    startTime time.Time
                //    endTime time.Time
                // )
				//
                //startTime, err = time.Parse("2006-01-02 15:04:05",start)
                //endTime, err = time.Parse("2006-01-02 15:04:05",end)
                //st := startTime.Format("2006-01-02 15:04:05")
                //et := endTime.Format("2006-01-02 15:04:05")

                querySQL := "INSERT INTO `merchant_charge_rule` (`fee_per`,`start`,`end`,`interval`,`interval_unit`,`merchant_id`,`room_id`,`flag`,`name`) VALUES (?,?,?,?,?,?,?,?,?)"
                var lastInsertId int64
                lastInsertId,err = m.Query(querySQL,fee,start,end,interval,intervalUnit,merchantID,roomId,flag,name).LastInsertId()
                if err == nil {
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RCRCreateResponse{true,uint32(lastInsertId)})
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

func (s *RoomServiceImp) UpdateRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRUpdateRequest) (res*Aphro_CommonBiz.Response,err error) {
    fee := in.Fee
    start := in.Start
    end := in.End
    interval := in.Interval
    intervalUnit := in.IntervalUnit
    roomId := in.RoomID
    sessionToken := in.SessionToken
    rcrid := in.RCRID
    roomName := in.Name

    _, merchantID, err := Session.FetchSessionTokenValue(sessionToken)
    if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {

                querySQL := "UPDATE `merchant_charge_rule` SET `fee_per` = ? , `start` = ? , `end` = ? , `interval` = ? , `interval_unit` = ? , `merchant_id` = ? , `room_id` = ? ,`name` = ? WHERE `ID` = ? "
                _,err = m.Query(querySQL,fee,start,end,interval,intervalUnit,merchantID,roomId,roomName,rcrid).RowsAffected()
                if err == nil {
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RCRUpdateResponse{true})
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

func (s *RoomServiceImp) QueryRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRQueryRequest) (res *Aphro_CommonBiz.Response,err error) {

    sessionToken := in.SessionToken
    roomId := in.RoomID
    inMerchantID := in.MerchantID
    rcrID := in.RCRID
    _, merchantID, err := Session.FetchSessionTokenValue(sessionToken)
    if err == nil {
		var mysql *MySQL.APSMySQL
		mysql,err = MySQL.NewAPSMySQL(nil)
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
                querySQL := "SELECT `ID`,`fee_per`,`start`,`end`,`interval`,`interval_unit`,`merchant_id`,`room_id`,`flag`,`name` FROM `merchant_charge_rule` WHERE " + strings.Join(whereCondition," AND ")
                var (
                    r_ID            uint32
                    r_fee           float32
                    r_start         string
                    r_end           string
                    r_interval      int
                    r_interval_unit int
                    r_merchant_id   uint32
                    r_room_id       uint32
                    r_flag          int
                    r_name          string
                )

                qr := &Aphro_Room_pb.RCRQueryResponse{}
                //var rooms []*Aphro_Room_pb.RSResult
                err = m.QueryAll(querySQL,binds...).FetchAll(func(dest...interface{}){
                    if err != nil {
                        return
                    }
                    rsResult := &Aphro_Room_pb.RCRResult{
                        MerchantID:   r_merchant_id,
                        RCRID:        r_ID,
                        Fee:          r_fee,
                        Start:        r_start,
                        End:          r_end,
                        Interval:     uint32(r_interval),
                        IntervalUnit: uint32(r_interval_unit),
                        RoomID:       r_room_id,
                        Flag:         uint32(r_flag),
                        Name:         r_name,
                    }
                    qr.Results = append(qr.Results, rsResult)
                },&r_ID,&r_fee,&r_start,&r_end,&r_interval,&r_interval_unit,&r_merchant_id,&r_room_id,&r_flag,&r_name)

                if err == nil {
                    qr.Success = true
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,qr)
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

func (s *RoomServiceImp) DeleteRoomChargeRule(ctx context.Context, in *Aphro_Room_pb.RCRDeleteRequest) (res *Aphro_CommonBiz.Response,err error) {

    sessionToken := in.SessionToken

    roomId := in.RoomID
    inMerchantID := in.MerchantID
    rcrID := in.RCRID

    _, merchantID, err := Session.FetchSessionTokenValue(sessionToken)
    if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
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
                _,err = m.Query(querySQL,binds...).RowsAffected()
                if err == nil {
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RCRDeleteResponse{true})
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return

}

func (s *RoomServiceImp) RoomTransactionBegin(ctx context.Context, in *Aphro_Room_pb.RSTransactionBeginRequest) (res *Aphro_CommonBiz.Response,err error) {
    roomChargeRules:= in.RoomChargeRules
    roomId := in.RoomID
    sessionToken := in.SessionToken
    var mid string
	_,mid,err = Session.FetchSessionTokenValue(sessionToken)

	if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
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
                err = m.Query(querySQL,roomId).FetchRow(&room_name,&status,&terminal_code)
                if err == nil && status == RoomStatusEnable {
                    //新增一个事务,房间快照
                    querySQL := "INSERT  INTO `transaction_room` (`room_id`,`room_name`,`start_time`,`update_time`,`status`,`terminal_code`) VALUES (?,?,?,?,?,?)"
                    startTime := time.Now()
                    var transactionId int64
                    transactionId,err = m.Query(querySQL,roomId,room_name,startTime,startTime,RoomStatusInUse,terminal_code).LastInsertId()
                    if err == nil {
                        //房间规则快照
                        rulesCount := len(roomChargeRules)
                        insertValuesPlaceholder := [][]string{}
                        bv := []interface{}{}
                        for i := rulesCount -1 ; i >= 0 ; i--  {
                            insertValuesPlaceholder = append(insertValuesPlaceholder,[]string{"?","?","?","?","?","?","?","?","?","?"})
                            rule := roomChargeRules[i]
                            bv = append(bv,rule.Fee)
                            bv = append(bv,rule.Start)
                            bv = append(bv,rule.End)
                            bv = append(bv,rule.Interval)
                            bv = append(bv,rule.IntervalUnit)
                            bv = append(bv,mid)
                            bv = append(bv,roomId)
                            bv = append(bv,transactionId)
                            bv = append(bv,rule.Flag)
                            bv = append(bv,rule.RcrID)
                        }

                        insertColumns := []string{
                                "fee_per_interval",
                                "start",
                                "end",
                                "interval",
                                "interval_unit",
                                "merchant_id",
                                "room_id",
                                "transaction_id",
                                "flag",
                                "rcrID",
                        }

                        _,err = m.Insert("transaction_room_charge_rules",insertColumns , insertValuesPlaceholder).Execute(bv...).RowsAffected()
                        if err == nil {
                            //todo 计费开始

                            t := Biz.TransactionCalculator{}
                            rs := t.BatchReformerRuleByRCRCreatePB(roomChargeRules, mid)
                            var fee float64
                            fee,err = t.ScheduleRulesByRules(rs,startTime,startTime)




                            //当前的房间 不支持多个事务
                            _,err := m.
                                Update("merchant_room",map[string]interface{}{"status":"?"}).
                                Where(&MySQL.APSMySQLCondition{MySQL.APSMySQLOperator_Equal,"ID","?"}).
                                Execute(RoomStatusInUse,roomId).
                                RowsAffected()
                            if err == nil {
                                res ,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RSTransactionBeginResponse{true,uint32(transactionId),uint64(time.Now().Unix() * 1000),float32(fee)})
                            }
                        }
                    }
                } else if status != RoomStatusEnable {
					err = Error.NewCustomError(Error.BizError,"房间不可用")
				}
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}
// 挂起一个房间的事务
func (s *RoomServiceImp) RoomTransactionSuspend(ctx context.Context, in *Aphro_Room_pb.RSTransactionSuspendRequest) (res *Aphro_CommonBiz.Response,err error){
    transactionID := in.TransactionID
    roomId := in.RoomID
    sessionToken := in.SessionToken
    //merchantID := in.MerchantID
    //_,  err = Session.IsSessionTokenVailate(sessionToken)
	_,_,err = Session.FetchSessionTokenValue(sessionToken)

	if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
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
                err = m.Query(querySQL,roomId).FetchRow(&room_name,&status,&terminal_code)
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

                    querySQL := "UPDATE `transaction_room` SET `update_time` = ? , `status` = ? WHERE " + strings.Join(whereCondition," AND ")
                    _,err = m.Query(querySQL,binds...).RowsAffected()
                    if err == nil {
                        res ,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RSTransactionSuspendResponse{true})
                    }
                } else if err == nil {
					err = Error.NewCustomError(Error.BizError,"房间不可用")
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}

// 结束一个房间的事务
func (s *RoomServiceImp) RoomTransactionEnd(ctx context.Context, in *Aphro_Room_pb.RSTransactionEndRequest) (res *Aphro_CommonBiz.Response,err error) {
    transactionID := in.TransactionID
    roomId := in.RoomID
    sessionToken := in.SessionToken
    //merchantID := in.MerchantID
    //_,  err = Session.IsSessionTokenVailate(sessionToken)
	_,_,err = Session.FetchSessionTokenValue(sessionToken)

	if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
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
                err = m.Query(querySQL,roomId).FetchRow(&room_name,&status,&terminal_code)
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

                    querySQL := "UPDATE `transaction_room` SET `update_time` = ? , `status` = ? WHERE " + strings.Join(whereCondition," AND ")
                    _,err = m.Query(querySQL,binds...).RowsAffected()
                    if err == nil {
                        querySQL := "UPDATE `merchant_room` SET `status` = ?  WHERE `ID` = ? LIMIT 1"
                        _,err = m.Query(querySQL,RoomStatusEnable,roomId).RowsAffected()
                        if err == nil {
                            res ,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RSTransactionSuspendResponse{true})
                        }
                    }
                } else if err == nil {
					err = Error.NewCustomError(Error.BizError,"房间不可用")
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")            }
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}
// 创建一个房间的房费
func (s *RoomServiceImp) RoomTransactionCreateRoomFee(ctx context.Context, in *Aphro_Room_pb.RSTransactionCreateRoomFeeRequest) (res *Aphro_CommonBiz.Response,err error) {

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
    //_,  err = Session.IsSessionTokenVailate(sessionToken)
	_,_,err = Session.FetchSessionTokenValue(sessionToken)

	if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                //startTime := time.Now()
                querySQL := "INSERT  INTO `transaction_roomfee` (`fee`,`fee_per_interval`,`start`,`end`,`interval`,`interval_unit`,`merchant_id`,`room_id`,`transaction_id`,`flag`) VALUES (?,?,?,?,?,?,?,?,?,?)"
                _,err = m.Query(querySQL,
                    0,
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
                    res ,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RSTransactionCreateRoomFeeResponse{true})
                }
            }
        } else {
			err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}
// 查询一个房间的房费
func (s *RoomServiceImp) RoomTransactionQueryRoomFee(ctx context.Context, in *Aphro_Room_pb.RSTransactionQueryRoomFeeRequest) (res *Aphro_CommonBiz.Response,err error) {
    transactionID := in.TransactionID
    roomId := in.RoomID
    sessionToken := in.SessionToken
    transactionRoomFeeID := in.TransactionRoomFeeID
    //merchantID := in.MerchantID
    //_,  err = Session.IsSessionTokenVailate(sessionToken)
	_,_,err = Session.FetchSessionTokenValue(sessionToken)

	if err == nil {
    	var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
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
                querySQL := "SELECT `ID`,`fee`,`create_time`,`update_time`,`fee_per_interval`,`start`,`end`, `interval`,`interval_unit`,`merchant_id`,`room_id`,`transaction_id`,`flag` FROM `transaction_roomfee` WHERE " + strings.Join(whereCondition," AND ")
                var (
                    ID               uint32
                    fee              float32
                    create_time      string
                    update_time      string
                    fee_per_interval float32
                    start            string
                    end              string
                    interval         uint32
                    interval_unit    uint32
                    merchant_id      uint32
                    room_id          uint32
                    transaction_id   uint32
                    flag             uint32
                )

                var roomFeeResultList []*Aphro_Room_pb.RSTransactionRoomFeeResult;

                err = m.QueryAll(querySQL,binds...).FetchAll(func(dest...interface{}){
                    t := &Aphro_Room_pb.RSTransactionRoomFeeResult{
						ID, fee, create_time, update_time, fee_per_interval,
						start, end, interval, interval_unit, merchant_id,
						room_id, transaction_id, flag,}
                    roomFeeResultList = append(roomFeeResultList,t)
                },&ID,&fee,&create_time,&update_time,&fee_per_interval,&start,&end,&interval,&interval_unit,&merchant_id,&room_id,&transaction_id,&flag)
                if err == nil {
                    res ,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Room_pb.RSTransactionQueryRoomFeeResponse{true,roomFeeResultList})
                }
            } else {
				err = Error.NewCustomError(Error.BizError,"mysql类型断言错误")
			}
        }
    }
	if err != nil {
		res,err = Response.NewCommonBizResponseWithError(err,nil)
	}
    return
}