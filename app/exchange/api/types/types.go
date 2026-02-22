package types

type UserOrdersResp struct {
	UserID int64       `json:"user_id"`
	Orders []OrderInfo `json:"user_orders"`
}

type UserOrdersReq struct {
	UserID int64 `json:"user_id"`
}

type OrderInfo struct {
	Id             string `json:"id"`               //id
	OrderId        string `json:"order_id"`         //订单id
	UserId         int64  `json:"user_id"`          //用户id
	SymbolName     string `json:"symbol_name"`      //交易对名
	Price          string `json:"price"`            //价格
	Qty            string `json:"qty"`              //数量
	Amount         string `json:"amount"`           //金额
	Side           int32  `json:"side"`             //方向
	Status         int32  `json:"status"`           // 状态
	OrderType      int32  `json:"order_type"`       //订单类型
	FilledQty      string `json:"filled_qty"`       //成交数量
	FilledAmount   string `json:"filled_amount"`    //成交金额
	FilledAvgPrice string `json:"filled_avg_price"` //成交均价
	CreatedAt      int64  `json:"created_at"`       //创建时间
}
