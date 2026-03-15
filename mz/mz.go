package mz

import (
	"errors"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/models/order"
	json "github.com/json-iterator/go"
	"net/url"
	"time"

	"strconv"
)

type SessionVerifyData struct {
	RetCode int    `json:"code"`
	Message string `json:"message,omitempty"`
	Value   string `json:"value,omitempty"`
}

var tokenUrl = "https://sdk-store.mlinkapp.com/game/security/checksession"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".mz."
	ts := strconv.Itoa(int(app.TimenowInTimezone().UnixMilli()))
	// 初始化参数Map
	postData := url.Values{}
	postData.Add("app_id", config.Get(conPre+"app_id"))
	postData.Add("session_id", s.Session)
	postData.Add("uid", s.Uid)
	postData.Add("ts", ts)
	postData.Add("sign", wtplatform.Sign(postData, ":"+config.Get(conPre+"app_secret")))
	postData.Add("sign_type", "md5")

	res, err := myhttp.FormRequest().SetBody(postData.Encode()).Post(tokenUrl)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.RetCode != 200 {
		err = fmt.Errorf("session验证失败，错误码:%d,错误信息：%s", m.RetCode, m.Message)
		return
	}
	return
}

type OrderData struct {
	RetCode int    `json:"code"`
	Message string `json:"message,omitempty"`
	Value   struct {
		AppId       int     `json:"appId"`
		OrderId     int     `json:"orderId"`
		Uid         int     `json:"uid"`
		BuyAmount   int     `json:"buyAmount,omitempty"`
		CpOrderId   string  `json:"cpOrderId,omitempty"`
		TotalPrice  float64 `json:"totalPrice,omitempty"`
		TradeStatus int     `json:"tradeStatus,omitempty"`
		PayTime     int64   `json:"payTime,omitempty"`
		UserInfo    string  `json:"userInfo,omitempty"`
	}
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	appId := strconv.Itoa(orderModel.AppId)
	// 初始化参数Map
	prams := make(map[string]any)
	prams["app_id"] = config.Get(appId + ".mz.app_id")
	prams["cp_order_id"] = orderModel.BsTradeNo
	prams["uid"] = orderModel.PlatUid
	prams["product_id"] = "0"
	prams["product_subject"] = orderModel.GoodsName
	prams["product_body"] = orderModel.Body
	prams["product_unit"] = ""
	prams["buy_amount"] = "1"
	prams["product_per_price"] = helpers.DivideToString(orderModel.TotalFee)
	prams["total_price"] = helpers.DivideToString(orderModel.TotalFee)
	prams["create_time"] = strconv.Itoa(int(app.TimenowInTimezone().Unix()))
	prams["pay_type"] = "0"
	prams["user_info"] = ""
	prams["sign_type"] = "md5"
	prams["sign"] = OrderSign(prams, appId)
	return prams, nil
}

var orderUrl = "https://sdk-store.mlinkapp.com/game/order/query"

func SearchOrder(orderModel order.Order) (map[string]any, error) {
	appId := strconv.Itoa(orderModel.AppId)
	params := make(map[string]any)
	params["app_id"] = config.Get(appId + ".mz.app_id")
	params["cp_order_id"] = orderModel.BsTradeNo
	params["ts"] = strconv.Itoa(int(app.TimenowInTimezone().Unix()))
	params["sign_type"] = "md5"
	params["sign"] = OrderSign(params, appId)
	postData := url.Values{}
	for k, v := range params {
		postData.Add(k, v.(string))
	}
	res, err := myhttp.FormRequest().SetBody(postData.Encode()).Post(orderUrl)
	if err != nil {
		return nil, err
	}
	// request := map[string]string{}
	request := OrderData{}
	err = json.Unmarshal(res.Body(), &request)
	if err != nil {
		return nil, err
	}
	if request.RetCode != 200 {
		return nil, errors.New(request.Message)
	}
	if orderModel.BsTradeNo != request.Value.CpOrderId {
		return nil, errors.New("订单号不存在")
	}
	pic := int(request.Value.TotalPrice * 100)
	if pic != orderModel.TotalFee {
		return nil, errors.New("订单价格异常")
	}
	if strconv.Itoa(request.Value.Uid) != orderModel.PlatUid {
		return nil, errors.New("用户id错误")
	}
	tradeState := CasePayStatus(strconv.Itoa(request.Value.TradeStatus))
	orderMap := map[string]interface{}{
		"trade_state":      tradeState,
		"trade_state_desc": "",
		"transaction_id":   request.Value.OrderId,
		"pay_time":         time.UnixMilli(request.Value.PayTime).Format("2006-01-02 15:04:05"),
		"trade_type":       "MZ_APP",
		"pay_fee":          pic,
	}
	return orderMap, nil
}

func OrderSign(params map[string]any, gameId string) string {
	return wtplatform.MapSign(params, ":"+config.Get(gameId+".mz.app_secret"))
}

func CasePayStatus(status string) int {
	switch status {
	case "3": // 交易支付成功
		return 1
	case "1", "2": // 未支付
		return 2
	case "4": // 已关闭
		return 9
	default: // 未知
		return 0
	}
}
