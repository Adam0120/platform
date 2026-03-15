package xiaomi

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/config"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/logger"
	json "github.com/json-iterator/go"
	"go.uber.org/zap"
	"net/url"
	"sort"
	"strings"
)

type SessionVerifyData struct {
	ErrCode int    `json:"errcode"`
	Adult   int    `json:"adult"`
	Age     int    `json:"age"`
	UnionId string `form:"unionId"`
}

var tokenUrl = "https://mis.migc.xiaomi.com/api/biz/service/loginvalidate"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".mi."
	postData := url.Values{}
	postData.Add("appId", config.Get(conPre+"app_id"))
	postData.Add("uid", s.Uid)
	postData.Add("session", s.Session)
	postData.Add("signature", Sign(postData, config.Get(conPre+"app_secret")))
	res, err := myhttp.FormRequest().SetBody(postData.Encode()).Post(tokenUrl)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.ErrCode != 200 {
		err = fmt.Errorf("session验证失败，错误码：%d", m.ErrCode)
		return
	}
	return
}

func Sign(params url.Values, key string) string {
	o := hmac.New(sha1.New, []byte(key))
	s := wtplatform.Sort(params)
	o.Write(s.Bytes())
	return hex.EncodeToString(o.Sum(nil))
}

var orderUrl = "https://mis.migc.xiaomi.com/api/biz/service/queryOrder.do"

func SearchOrder(outNo, platUid, appId string) (map[string]interface{}, error) {
	Query := url.Values{}
	Query.Add("appId", config.Get(appId+".mi.app_id"))
	Query.Add("uid", platUid)
	Query.Add("cpOrderId", outNo)
	sign := Sign(Query, config.Get(appId+".mi.app_secret"))
	Query.Add("signature", sign)
	res, err := myhttp.Request().Get(orderUrl + "?" + Query.Encode())
	if err != nil {
		return nil, err
	}
	request := map[string]interface{}{}
	err = json.Unmarshal(res.Body(), &request)
	if err != nil {
		return nil, err
	}

	// 验签
	signStr := make(map[string]string)
	for k, v := range request {
		signStr[k], _ = url.QueryUnescape(fmt.Sprintf("%v", v))
	}
	signature := signStr["signature"]
	delete(signStr, "signature")
	if sign = OrderSign(signStr, appId); sign != signature {
		logger.Error("小米验签失败", zap.String("sign", sign), zap.String("signature", signature))
	}
	orderMap := map[string]interface{}{
		"trade_state":      CasePayStatus(fmt.Sprintf("%v", request["orderStatus"])),
		"trade_state_desc": request["orderStatus"],
		"transaction_id":   request["orderId"],
		"pay_time":         request["payTime"],
		"trade_type":       "APP",
		"pay_fee":          request["payFee"],
	}
	return orderMap, nil
}

func CasePayStatus(status string) int {
	switch status {
	case "TRADE_SUCCESS": // 交易支付成功
		return 1
	case "WAIT_BUYER_PAY": // 交易创建，等待买家付款
		return 2
	case "TRADE_CLOSED": // 未付款交易超时关闭，或支付完成后全额退款未付款交易超时关闭，或支付完成后全额退款
		return 9
	case "TRADE_FINISHED": // 交易结束，不可退款
		return 4
	default: // 未知
		return 0
	}
}

func OrderSign(params map[string]string, gameId string) string {
	key := config.Get(gameId + ".mi.app_secret")
	o := hmac.New(sha1.New, []byte(key))
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for i, k := range keys {
		keys[i] = fmt.Sprintf("%s=%s", k, params[k])
	}
	o.Write([]byte(strings.Join(keys, "&")))
	return hex.EncodeToString(o.Sum(nil))
}
