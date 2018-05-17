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
    "github.com/lampard1014/aphro/CommonBiz/Error"
	"strconv"
)

const (
    RPCPort = ":10085"
)
type CommodityServiceImp struct{}

func (s *CommodityServiceImp) CommodityCreate(ctx context.Context, in *Aphro_Commodity_pb.CommodityCreateRequest) (res *Aphro_CommonBiz.Response,err error) {
    commodityInfo := in.Good;
    var  merchantID string
    _, merchantID ,err = Session.FetchSessionTokenValue(in.SessionToken)
    if err == nil {
        var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                querySQL := "INSERT INTO `merchant_commodity` (`name`,`price`,`merchant_id`) VALUES( ?, ?, ?)"
                var lastInsertID int64
                lastInsertID,err = m.Query(querySQL,commodityInfo.Name,commodityInfo.Price,merchantID).LastInsertId()
                if err == nil {
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Commodity_pb.CommodityCreateResponse{true,uint32(lastInsertID)})
                }
                defer m.Close()
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

    
func (s *CommodityServiceImp) CommodityDelete(ctx context.Context, in *Aphro_Commodity_pb.CommodityDeleteRequest) (res *Aphro_CommonBiz.Response, err error) {

    inMerchantID := in.MerchantID
    inCommodityID := in.Id

    _, _ ,err = Session.FetchSessionTokenValue(in.SessionToken)

    if err == nil {
        var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
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
                _,err = m.Query(querySQL,binds...).RowsAffected()
                if err == nil {
                    //制作 令牌
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Commodity_pb.CommodityDeleteResponse{true})
                }
                defer m.Close()
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



func (s *CommodityServiceImp) CommodityUpdate(ctx context.Context, in *Aphro_Commodity_pb.CommodityUpdateRequest) (res *Aphro_CommonBiz.Response, err error) {

    sessionToken := in.SessionToken
    commodityID := in.Id
    inMerchantID := in.MerchantID
    name := in.Name
    price := in.Price

    //验证token的合法性
    var merchantID string
    _, merchantID,err = Session.FetchSessionTokenValue(sessionToken)

	if inMerchantID != 0{
		merchantID = strconv.Itoa(int(inMerchantID))
	}

    if err == nil {
        var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {

                querySQL := "UPDATE `merchant_commodity` SET `name`= ? , `price` = ? WHERE ID = ? AND `merchant_id` = ? "
                _,err = m.Query(querySQL,name,price,commodityID,merchantID).RowsAffected()
                if err == nil {
                    res,err  = Response.NewCommonBizResponseWithCodeWithError(0,err,&Aphro_Commodity_pb.CommodityUpdateResponse{true})
                }
                defer m.Close()
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

func (s *CommodityServiceImp) CommodityQuery(ctx context.Context, in *Aphro_Commodity_pb.CommodityQueryRequest) (res *Aphro_CommonBiz.Response, err error) {
    sessionToken := in.SessionToken
    inMerchantID := in.MerchantID
    commdityID := in.Id
    // commodityInfo := in.Goods;
    var goodRes *Aphro_Commodity_pb.CommodityQueryResponse = &Aphro_Commodity_pb.CommodityQueryResponse{}

	var merchantID string
	_, merchantID,err = Session.FetchSessionTokenValue(sessionToken)
	if inMerchantID != 0{
		merchantID = strconv.Itoa(int(inMerchantID))
	}
    if err == nil {
        var mysql *MySQL.APSMySQL
        mysql,err = MySQL.NewAPSMySQL(nil)
        if err == nil {
            m, ok := mysql.Connect().(*MySQL.APSMySQL)
            if ok {
                var whereCondition []string = []string{"1"}
                var binds []interface{} = []interface{}{}
                if merchantID != "" {
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
                err = m.QueryAll(querySQL,binds...).FetchAll(func(dest...interface{}){
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
                    res,err = Response.NewCommonBizResponseWithCodeWithError(0,err,goodRes)
                }
                defer m.Close()
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