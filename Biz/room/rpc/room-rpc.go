package  main

import (
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    "fmt"
    _ "github.com/go-sql-driver/mysql"
    "github.com/lampard1014/aphro/Biz/room/pb/room"
    "github.com/lampard1014/aphro/Biz/room/pb/roomChargeRule"
    "github.com/lampard1014/aphro/CommonBiz/Response/PB"
    "github.com/lampard1014/aphro/CommonBiz/Session"
    "github.com/lampard1014/aphro/PersistentStore/MySQL"
    "github.com/lampard1014/aphro/CommonBiz/Response"
    "github.com/lampard1014/aphro/Gateway/error"
    "strconv"
)

const (
	port  = ":10090"
)

const (
    //房间状态 可用
    RoomStatusEnable int = 0
    //房间状态 使用中
    RoomStatusInUse int = 1
    //房间状态 不可用
    RoomStatusDisable int = 2
)

type roomService struct{}

func (s *roomService) TerminalBind(ctx context.Context, in *Aphro_Room_pb.RSTerminalBindRequest) (*Aphro_CommonBiz.Response, error) {
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

func (s *roomService) TerminalUnbind(ctx context.Context, in *Aphro_Room_pb.RSTerminalUnbindRequest) (*Aphro_CommonBiz.Response, error) {
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

func (s *roomService) Create(ctx context.Context, in *Aphro_Room_pb.RSCreateRequest) (*Aphro_CommonBiz.Response, error) {
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

func (s *roomService) Update(ctx context.Context, in *Aphro_Room_pb.RSUpdateRequest) (*Aphro_CommonBiz.Response, error) {


    sessionToken := in.SessionToken
    terminalCode := in.TerminalCode
    location := in.Location
    latitude := location.Latitude
    longitude := location.Longitude
    roomName := in.RoomName
    roomId := in.RoomID
    chargeRules := in.ChargeRules
    status := in.Status
    //float	fee	=	2;
    //string	start	=3;
    //string	end	=	4;
    //uint32 interval = 5;
    //uint32 intervalUnit = 6;
    //uint32 merchantID = 7;
    //uint32 roomID = 8;

    var returnErr error = nil
    var res *Aphro_CommonBiz.Response = nil
    _, merchantID, sessionTokenError := Session.FetchSessionTokenValue(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            defer m.Close()
            if ok {
                
                var insertData [][]interface{}
                for _,cr := range chargeRules {
cr.
                    cr.(Aphro_RoomChargeRule.RCRCreateRequest)
                    if d, ok := cr.(*Aphro_RoomChargeRule.RCRCreateRequest); ok {

                    }
                }
                
                _,err := m.Update("merchant_room",map[string]interface{}{
                    "merchant_id":"?",
                    "longitude":"?",
                    "latitude":"?",
                    "room_name":"?",
                    "status":"?",
                    "terminal_code":"?",
                    "charge_rules":"?",
                }).Where(&MySQL.APSMySQLCondition{MySQL.APSMySQLOperator_Equal,"ID","?"}).Execute(merchantID,longitude,latitude,roomName,status,terminalCode,"",roomId).RowsAffected()
                if err == nil {
                    res,returnErr = Response.NewCommonBizResponse(0,err.Error(),&Aphro_Room_pb.RSUpdateResponse{Success:true})
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

func (s *roomService) Delete(ctx context.Context, in *Aphro_Room_pb.RSDeleteRequest) (*Aphro_CommonBiz.Response, error) {
}

func (s *roomService) Query(ctx context.Context, in *Aphro_Room_pb.RSQueryRequest) (*Aphro_CommonBiz.Response, error) {
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
