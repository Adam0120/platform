package jy

import (
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/models/order"
	json "github.com/json-iterator/go"
	"sort"
	"strconv"
)

type SessionVerifyData struct {
	ID    int `json:"id"`
	State struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	} `json:"state,omitempty"`
	Data struct {
		AccountId string `json:"accountId"`
		Creator   string `json:"creator"`
		NickName  string `json:"nickName"`
	} `json:"data,omitempty"`
}

var tokenUrl = "http://sdk.9game.cn/cp/account.verifySession"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".jy."
	data := map[string]interface{}{
		"sid": s.Session,
	}
	key := config.Get(conPre + "app_key")
	// 初始化参数Map
	prams := map[string]interface{}{
		"id":   app.TimenowInTimezone().Unix(),
		"data": data,
		"game": map[string]int{
			"gameId": config.GetInt(conPre + "app_id"),
		},
		"sign": Sign(data, key),
	}

	res, err := myhttp.JsonRequest().SetBody(prams).Post(tokenUrl)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.State.Code != 1 {
		err = fmt.Errorf("session验证失败，错误码:%d,错误信息：%s", m.State.Code, m.State.Msg)
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
	} `json:"value,omitempty"`
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	// 初始化参数Map
	prams := make(map[string]any)
	prams["cpOrderId"] = orderModel.BsTradeNo
	prams["accountId"] = orderModel.PlatUid
	prams["amount"] = helpers.DivideToString(orderModel.TotalFee)
	prams["callbackInfo"] = ""
	prams["notifyUrl"] = config.Get("app.url") + "/order/jy/notify/" + strconv.Itoa(orderModel.AppId)
	prams["signType"] = "MD5"
	prams["sign"] = Sign(prams, strconv.Itoa(orderModel.AppId))
	return prams, nil
}

func Sign(pram map[string]any, key string) string {
	var str string
	var keys []string
	for k := range pram {
		if k == "sign" || k == "signType" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys) // 对键进行排序
	for _, k := range keys {
		str += fmt.Sprintf("%s=%s", k, pram[k].(string))
	}
	return helpers.Md5String(str + key)
}
