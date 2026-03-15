package vivo

import (
	"errors"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
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
	RetCode int   `json:"retcode"`
	Data    *Date `json:"data,omitempty"`
}

type Date struct {
	Success bool   `json:"success"`
	Openid  string `json:"openid"`
}

var tokenUrl1 = "https://joint-account.vivo.com.cn/cp/user/auth"
var tokenUrl2 = "https://joint-account-cp.vivo.com.cn/cp/user/auth"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	postData := url.Values{}
	postData.Add("opentoken", s.Session)
	turl := tokenUrl1
	var num int
DoPost:
	res, err := myhttp.FormRequest().SetBody(postData.Encode()).Post(turl)
	if err != nil {
		if num == 0 {
			turl = tokenUrl2
			num++
			goto DoPost
		}
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		if num == 0 {
			turl = tokenUrl2
			num++
			goto DoPost
		}
		return
	}
	if m.RetCode != 0 {
		err = fmt.Errorf("session验证失败，错误码:%d", m.RetCode)
		return
	}
	if m.Data.Openid != s.Uid {
		err = fmt.Errorf("pid 验证失败")
		return
	}

	return
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	conPre := strconv.Itoa(orderModel.AppId) + ".vivo."
	// 初始化参数Map
	prams := make(map[string]any)
	prams["appId"] = config.Get(conPre + "app_id")
	prams["cpOrderNumber"] = orderModel.BsTradeNo
	prams["productName"] = orderModel.GoodsName
	prams["productDesc"] = orderModel.Body
	prams["orderAmount"] = strconv.Itoa(orderModel.TotalFee)
	prams["notifyUrl"] = config.Get("app.url") + "/order/vivo/notify/" + strconv.Itoa(orderModel.AppId)
	prams["vivoSignature"] = Sign(prams, config.Get(conPre+"app_key_pay"))
	prams["platUid"] = orderModel.PlatUid
	return prams, nil
}

func Sign(pram map[string]any, key string) string {
	str := ""
	keys := make([]string, 0)
	for k, v := range pram {
		if k == "signature" || k == "signMethod" || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys) // 对键进行排序
	for _, k := range keys {
		str += fmt.Sprintf("%s=%v&", k, pram[k])
	}
	return helpers.Md5String(str + helpers.Md5String(key))
}

var orderUrl = "https://pay.vivo.com.cn/vcoin/queryv2"

func SearchOrder(orderModel order.Order) (map[string]interface{}, error) {
	conPre := strconv.Itoa(orderModel.AppId) + ".vivo."
	postData := url.Values{}
	postData.Add("appId", config.Get(conPre+"app_id"))
	postData.Add("cpId", config.Get(conPre+"cp_id"))
	postData.Add("cpOrderNumber", orderModel.BsTradeNo)
	postData.Add("orderAmount", strconv.Itoa(orderModel.TotalFee))
	postData.Add("version", "1.0.0")
	postData.Add("signature", wtplatform.Sign(postData, helpers.Md5String(config.Get(conPre+"app_key_pay"))))
	postData.Add("signMethod", "MD5")
	res, err := myhttp.FormRequest().SetBody(postData.Encode()).Post(orderUrl)
	if err != nil {
		return nil, err
	}
	var request map[string]any
	err = json.Unmarshal(res.Body(), &request)
	if err != nil {
		return nil, err
	}
	if request["signature"] != Sign(request, config.Get(conPre+"app_key_pay")) {
		return nil, errors.New("验签失败")
	}
	if orderModel.BsTradeNo != request["cpOrderNumber"] {
		return nil, errors.New("订单号不存在")
	}
	if helpers.StringToInt(request["orderAmount"].(string)) != orderModel.TotalFee {
		return nil, errors.New("订单价格异常")
	}
	if request["uid"] != orderModel.PlatUid {
		return nil, errors.New("用户id错误")
	}
	tradeState := 0
	if request["respCode"] == "200" && request["tradeStatus"] == "0000" {
		tradeState = 1
	}
	orderMap := map[string]interface{}{
		"trade_state":      tradeState,
		"trade_state_desc": request["respMsg"],
		"transaction_id":   request["orderNumber"],
		"pay_time":         request["payTime"],
		"trade_type":       "VIVO_APP",
		"pay_fee":          request["orderAmount"],
	}
	return orderMap, nil
}
