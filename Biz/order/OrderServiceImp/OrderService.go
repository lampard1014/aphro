package OrderServiceIMP

import (
    "golang.org/x/net/context"
    _ "github.com/go-sql-driver/mysql"
    "github.com/lampard1014/aphro/CommonBiz/Response/PB"
    "github.com/lampard1014/aphro/Biz/order/PB"
    "github.com/lampard1014/aphro/CommonBiz/Session"
    "github.com/lampard1014/aphro/PersistentStore/MySQL"
    "github.com/lampard1014/aphro/CommonBiz/Response"
    "github.com/lampard1014/aphro/Biz/commodity/pb"
    "github.com/lampard1014/aphro/CommonBiz/Error"
)

const (
    RPCPort = ":10088"
)
type OrderServiceImp struct{}


// 创建订单
func (s *OrderServiceImp) CreateOrder(ctx context.Context, in *Aphro_Order_PB.OSCreateOrderRequest) (res *Aphro_CommonBiz.Response,err error) {

    transactionID := in.TransactionID
    roomID := in.RoomID
    amount := in.Amount
    outterTransactionID := in.OutterTransactionID
    userAgent := in.UserAgent
    comment := in.Comment
    propersalAccountType := in.PropersalAccountType
    commodities := in.Commodities

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
// 更新订单
func (s *OrderServiceImp) UpdateOrder(ctx context.Context, in *Aphro_Order_PB.OSUpdateOrderRequest) (*Aphro_CommonBiz.Response, error) {

}
// 删除订单
func (s *OrderServiceImp) DeleteOrder(ctx context.Context, in *Aphro_Order_PB.OSDeleteOrderRequest) (*Aphro_CommonBiz.Response, error) {

}
// 查询订单
func (s *OrderServiceImp) QueryOrder(ctx context.Context, in *Aphro_Order_PB.OSQueryOrderRequest) (*Aphro_CommonBiz.Response, error) {

}