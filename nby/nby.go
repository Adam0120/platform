package nby

import (
	"bytes"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/models/order"
	json "github.com/json-iterator/go"
	"net/url"
	"sort"
	"strconv"
)

type SessionVerifyData struct {
	Code int `json:"code"`
	Data struct {
	} `json:"data"`
	Message string `json:"message"`
}

var tokenUrl = "https://niugamecenter.nubia.com/VerifyAccount/CheckLogined"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".nby."
	appId := config.Get(conPre + "app_id")
	appSecret := config.Get(conPre + "app_secret")
	ts := app.GetStringUnix()
	signStr := sign(appId, appSecret, s, ts)
	val := url.Values{}
	val.Add("uid", s.Uid)
	val.Add("data_timestamp", ts)
	val.Add("game_id", s.Uid)
	val.Add("session_id", s.Session)
	val.Add("sign", signStr)
	res, err := myhttp.FormRequest().SetBody(val.Encode()).Post(tokenUrl)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.Code != 0 {
		err = fmt.Errorf("session验证失败，错误码:%s", m.Message)
		return
	}
	return
}

type NbyOrderData struct {
	OrderNo       string `json:"order_no" form:"order_no"`
	DataTimestamp string `json:"data_timestamp" form:"data_timestamp"`
	PaySuccess    int    `json:"pay_success" form:"pay_success"`
	Sign          string `json:"sign" form:"sign"`
	AppId         string `json:"app_id" form:"app_id"`
	Uid           string `json:"uid" form:"uid"`
	Amount        string `json:"amount" form:"amount"`
	ProductName   string `json:"product_name" form:"product_name"`
	ProductDes    string `json:"product_des" form:"product_des"`
	Number        int    `json:"number" form:"number"`
	OrderSerial   int    `json:"order_serial" form:"order_serial"`
	OrderSign     string `json:"order_sign" form:"order_sign"`
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	// 初始化参数Map
	aId := strconv.Itoa(orderModel.AppId)
	appId := config.Get(aId + ".nby.app_id")
	secretKey := config.Get(aId + ".nby.secret_key")
	prams := make(map[string]interface{})
	prams["app_id"] = appId
	prams["uid"] = orderModel.PlatUid
	prams["cp_order_id"] = orderModel.BsTradeNo
	prams["amount"] = helpers.DivideToString(orderModel.TotalFee)
	prams["product_name"] = orderModel.GoodsName
	prams["product_des"] = orderModel.Body
	prams["number"] = 1
	prams["data_timestamp"] = app.GetStringUnix()
	prams["game_id"] = orderModel.PlatUid
	prams["cp_order_sign"] = OrderSign(appId, secretKey, prams)
	return prams, nil
}

func OrderSign(appId, secretKey string, prams map[string]any) string {
	var keys []string
	for k := range prams {
		if k == "game_id" || k == "sign" || k == "order_serial" || k == "cp_order_sign" || k == "order_sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys) // 对键进行排序
	buf := bytes.Buffer{}
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s=%v&", k, prams[k]))
	}
	buf.Truncate(buf.Len() - 1)
	buf.WriteString(":")
	buf.WriteString(appId)
	buf.WriteString(":")
	buf.WriteString(secretKey)
	var str = buf.String()
	return helpers.Md5String(str)
}

func sign(appId, appSecret string, request *wtplatform.SessionVerifyRequest, ts string) string {
	s := bytes.Buffer{}
	s.WriteString("data_timestamp=")
	s.WriteString(ts)
	s.WriteString("&game_id=")
	s.WriteString(request.Uid)
	s.WriteString("&session_id=")
	s.WriteString(request.Session)
	s.WriteString("&uid=")
	s.WriteString(request.Uid)
	s.WriteString(":")
	s.WriteString(appId)
	s.WriteString(":")
	s.WriteString(appSecret)
	return helpers.Md5String(s.String())
}
