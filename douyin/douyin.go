package douyin

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/config"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/models/order"
	json "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"net/url"

	"strconv"
	"time"
)

const userVerifyUrl = "https://usdk.dailygn.com/gsdk/usdk/account/verify_user"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".douyin."
	appId := config.Get(conPre + "app_id")
	secretKey := config.Get(conPre + "secret_key")
	req := url.Values{}
	req.Add("app_id", appId)
	req.Add("access_token", s.Session)
	req.Add("ts", strconv.Itoa(int(time.Now().Unix())))
	req.Add("sign", GetSign(req, secretKey))
	res, err := myhttp.FormRequest().SetBody(req.Encode()).Post(userVerifyUrl)
	if err != nil {
		return
	}
	var m DyResponse
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.Code != 0 {
		err = fmt.Errorf("session验证失败，错误码:%s", m.Message)
		return
	}
	s.Uid = m.Data.SdkOpenId
	return
}

type DyResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		SdkOpenId string `json:"sdk_open_id"`
		AgeType   int32  `json:"age_type"`
	}
	LogId string `json:"log_id"`
}

func GetSign(params url.Values, secretKey string) string {
	//使用密钥进行Hmac-sha1加密
	mac := hmac.New(sha1.New, []byte(secretKey))
	c := wtplatform.Sort(params)
	mac.Write(c.Bytes())
	//base64编码获得最终的sign
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

const orderURL = "https://usdk.dailygn.com/gsdk/usdk/payment/live_pre_order"

type OrderReqeust struct {
	Aid           int    `json:"aid"`
	CpOrderId     string `json:"cp_order_id"`
	ProductId     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	ProductDesc   string `json:"product_desc"`
	ProductAmount int    `json:"product_amount"`
	SdkOpenId     string `json:"sdk_open_id"`
	//user_agent
	ClientIp        string `json:"client_ip"`
	CallbackUrl     string `json:"callback_url"`
	ActualAmount    int    `json:"actual_amount"`
	RiskControlInfo string `json:"risk_control_info"`
	TradeType       int    `json:"trade_type"`
	//extra_info
	Sign string `json:"sign"`
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	appId := strconv.Itoa(orderModel.AppId)
	param := url.Values{}
	param.Add("aid", config.Get(appId+".douyin.app_id"))
	param.Add("cp_order_id", orderModel.BsTradeNo)
	param.Add("product_id", orderModel.GoodsId)
	param.Add("product_name", orderModel.GoodsName)
	param.Add("product_desc", orderModel.Body)
	param.Add("product_amount", strconv.Itoa(orderModel.TotalFee))
	param.Add("sdk_open_id", orderModel.PlatUid)
	param.Add("client_ip", orderModel.UserIp)
	param.Add("callback_url", config.Get("app.url")+"/order/douyin/notify/"+appId)
	param.Add("actual_amount", strconv.Itoa(orderModel.TotalFee))
	param.Add("risk_control_info", orderModel.CpExtraInfo)
	param.Add("trade_type", "2")

	param.Add("sign", GetSign(param, config.Get(appId+".douyin.secret_key")))
	res, err := myhttp.FormRequest().SetBody(param.Encode()).Post(orderURL)
	if err != nil {
		return nil, err
	}
	var ret map[string]interface{}
	err = json.Unmarshal(res.Body(), &ret)
	if err != nil {
		return nil, err
	}
	code := cast.ToInt(ret["code"])
	if code != 0 {
		return nil, errors.New("CreateErr:" + cast.ToString(ret["message"]))
	}
	return map[string]interface{}{"sdk_param": ret["sdk_param"]}, nil
}
