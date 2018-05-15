package CommdityServiceIMP

import (
	"golang.org/x/net/context"
    "strings"
    _ "github.com/go-sql-driver/mysql"
    "github.com/lampard1014/aphro/Biz/commodity/pb"

    "github.com/lampard1014/aphro/CommonBiz/Response/PB"
    "github.com/lampard1014/aphro/CommonBiz/Session"
    "github.com/lampard1014/aphro/PersistentStore/MySQL"
    "github.com/lampard1014/aphro/CommonBiz/Response"
)

const (
    RPCPort = ":10085"
)
type CommodityServiceImp struct{}

func (s *CommodityServiceImp) CommodityCreate(ctx context.Context, in *Aphro_Commodity_pb.CommodityCreateRequest) (*Aphro_CommonBiz.Response, error) {
    commodityInfo := in.Good;
    var returnError error
    var res *Aphro_CommonBiz.Response

    _, merchantID ,err := Session.FetchSessionTokenValue(in.SessionToken)
    if err == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "INSERT INTO `merchant_commodity` (`name`,`price`,`merchant_id`) VALUES( ?, ?, ?)"
                _,err := m.Query(querySQL,commodityInfo.Name,commodityInfo.Price,merchantID).LastInsertId()
                if err == nil {
                    //制作 令牌
                    res,returnError = Response.NewCommonBizResponseWithError(0,err,&Aphro_Commodity_pb.CommodityCreateResponse{Successed:true})
                } else {
                    returnError = err
                }
                defer m.Close()
            } else {
                res,returnError = Response.NewCommonBizResponse(Response.BizError,"mysql类型断言错误",nil)
            }
        } else {
            returnError = err
        }
    } else {
        returnError = err
    }
   return res, returnError
}

    
func (s *CommodityServiceImp) CommodityDelete(ctx context.Context, in *Aphro_Commodity_pb.CommodityDeleteRequest) (*Aphro_CommonBiz.Response, error) {

    inMerchantID := in.MerchantID
    inCommodityID := in.Id

    var returnError error
    var res *Aphro_CommonBiz.Response

    _, _ ,err := Session.FetchSessionTokenValue(in.SessionToken)

    if err == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                var whereCondition []string = []string{"1"}
                var binds []interface{} = []interface{}{}
                if inMerchantID != 0 {
                    whereCondition = append(whereCondition, "`merchant_id` =  ?")
                    binds = append(binds,inMerchantID)
                }
                if inCommodityID != 0 {
                    whereCondition = append(whereCondition, "`ID` =  ?")
                    binds = append(binds,inCommodityID)
                }

                querySQL := "DELETE FROM  `merchant_commodity` WHERE " + strings.Join(whereCondition," AND ")
                _,err := m.Query(querySQL,binds...).RowsAffected()
                if err == nil {
                    //制作 令牌
                    res,returnError = Response.NewCommonBizResponseWithError(0,err,&Aphro_Commodity_pb.CommodityDeleteResponse{Successed:true})
                } else {
                    returnError = err
                }
                defer m.Close()
            } else {
                res,returnError = Response.NewCommonBizResponse(Response.BizError,"mysql类型断言错误",nil)
            }
        } else {
            returnError = err
        }
    } else {
        returnError = err
    }
   return res, returnError
}



func (s *CommodityServiceImp) CommodityUpdate(ctx context.Context, in *Aphro_Commodity_pb.CommodityUpdateRequest) (*Aphro_CommonBiz.Response, error) {

    sessionToken := in.SessionToken
    commodityID := in.Id
    merchantID := in.MerchantID
    name := in.Name
    price := in.Price
    var res *Aphro_CommonBiz.Response

    var returnError error
    //验证token的合法性
    _, sessionTokenError := Session.IsSessionTokenVailate(sessionToken)

    if sessionTokenError == nil {

        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {

                querySQL := "UPDATE `merchant_commodity` SET `name`= ? AND `price` = ? AND `merchant_id` = ? WHERE ID = ?"
                _,err := m.Query(querySQL,name,price,merchantID,commodityID).RowsAffected()
                if err == nil {
                    res,returnError = Response.NewCommonBizResponseWithError(0,err,&Aphro_Commodity_pb.CommodityUpdateResponse{Success:true})
                } else {
                    returnError = err
                }
                defer m.Close()
            } else {
                res,returnError = Response.NewCommonBizResponse(Response.BizError,"mysql类型断言错误",nil)
            }
        } else {
            returnError = err
        }
    } else {
        returnError = sessionTokenError
    }
   return res, returnError
}

func (s *CommodityServiceImp) CommodityQuery(ctx context.Context, in *Aphro_Commodity_pb.CommodityQueryRequest) (*Aphro_CommonBiz.Response, error) {
    sessionToken := in.SessionToken
    merchantID := in.MerchantID
    commdityID := in.Id
    // commodityInfo := in.Goods;
    var goodRes *Aphro_Commodity_pb.CommodityQueryResponse = &Aphro_Commodity_pb.CommodityQueryResponse{}

    var res *Aphro_CommonBiz.Response
    var returnError error

    _, sessionTokenError := Session.IsSessionTokenVailate(sessionToken)
    if sessionTokenError == nil {
        mysql,err := MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                var whereCondition []string = []string{"1"}
                var binds []interface{} = []interface{}{}
                if merchantID != 0 {
                    whereCondition = append(whereCondition, "`merchant_id` =  ?")
                    binds = append(binds,merchantID)
                }
                if commdityID != 0 {
                    whereCondition = append(whereCondition, "`ID` =  ?")
                    binds = append(binds,commdityID)
                }

                var (
                    _name string
                    _ID uint32
                    _price float32
                    _merchant_id uint32
                )

                querySQL := "SELECT `name`,`ID`,`price`,`merchant_id` FROM  `merchant_commodity` WHERE " + strings.Join(whereCondition," AND ")
                err := m.Query(querySQL,binds...).FetchAll(func(dest...interface{}){
                    if err != nil {
                        return
                    }
                    rsResult := &Aphro_Commodity_pb.InnerComodityInfo{
                        Name:_name,
                        Price:_price,
                        MerchantID:_merchant_id,
                        Id:_ID,
                    }
                    goodRes.Goods = append(goodRes.Goods, rsResult)
                },&_name,&_ID,&_price,&_merchant_id)

                if err == nil {
                    goodRes.Success = true
                    res,returnError = Response.NewCommonBizResponseWithError(0,err,goodRes)
                } else {
                    returnError = err
                }
                defer m.Close()
            } else {
                res,returnError = Response.NewCommonBizResponse(Response.BizError,"mysql类型断言错误",nil)
            }
        } else {
            returnError = err
        }
    } else {
        returnError = sessionTokenError
    }
    return res , returnError
}