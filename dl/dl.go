package dl

import (
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/models/order"
	json "github.com/json-iterator/go"
	"go.uber.org/zap/buffer"
	"strconv"
)

type SessionVerifyData struct {
	Birthday    int    `json:"birthday"`
	IdMd5       string `json:"id_md5"`
	Gender      int    `json:"gender"`
	IsCertified int    `json:"is_certified"`
	Roll        bool   `json:"roll"`
	IsAdult     int    `json:"is_adult"`
	Oversea     int    `json:"oversea"`
	Valid       string `json:"valid"`
	Times       int    `json:"times"`
	MsgCode     int    `json:"msg_code"`
	IdStatus    int    `json:"id_status"`
	MsgDesc     string `json:"msg_desc"`
	Interval    int    `json:"interval"`
	IdType      int    `json:"id_type"`
	Pi          string `json:"pi "`
	Age         int    `json:"age"`
}

var tokenUrl = "https://ctmaster.d.cn/api/cp/checkToken"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".dl."
	query := "?token=" + s.Session + "&appid=" + config.Get(conPre+"app_id") + "&umid=" + s.Uid + "&sig=" + loginSign(s.Session, s.Uid, conPre)
	res, err := myhttp.Request().Get(tokenUrl + query)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.MsgCode != 2000 {
		err = fmt.Errorf("session验证失败，错误码:%d,错误信息：%s", m.MsgCode, m.MsgDesc)
		return
	}
	return
}

func loginSign(token, umid, conPre string) string {
	str := config.Get(conPre+"app_id") + "|" + config.Get(conPre+"login_key") + "|" + token + "|" + umid
	return helpers.Md5String(str)
}

type OrderData struct {
	Result    string `json:"result" form:"result"`
	Money     string `json:"money" form:"money"`
	Order     string `json:"order" form:"order"`
	Mid       string `json:"mid" form:"mid"`
	Time      string `json:"time" form:"time"`
	Ext       string `json:"ext" form:"ext"`
	CpOrder   string `json:"cpOrder" form:"cpOrder"`
	Signature string `json:"signature" form:"signature"`
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	// 初始化参数Map
	prams := make(map[string]interface{})
	prams["cpOrder"] = orderModel.BsTradeNo
	prams["ext"] = ""
	prams["money"] = helpers.DivideToString(orderModel.TotalFee)
	prams["roleId"] = orderModel.RoleId
	prams["umid"] = orderModel.PlatUid

	buff := buffer.Buffer{}
	buff.WriteString(orderModel.BsTradeNo)
	buff.WriteString("|")
	buff.WriteString("")
	buff.WriteString("|")
	buff.WriteString(helpers.DivideToString(orderModel.TotalFee))
	buff.WriteString("|")
	buff.WriteString(orderModel.RoleId)
	buff.WriteString("|")
	buff.WriteString(orderModel.PlatUid)
	buff.WriteString("|")
	buff.WriteString(config.Get(strconv.Itoa(orderModel.AppId) + ".dl.payment_key"))
	sign := helpers.Md5String(buff.String())
	prams["cpSign"] = sign
	return prams, nil
}

func OrderSign(data OrderData, gameId string) string {
	buff := buffer.Buffer{}
	buff.WriteString("order=")
	buff.WriteString(data.Order)
	buff.WriteString("&money=")
	buff.WriteString(data.Money)
	buff.WriteString("&mid=")
	buff.WriteString(data.Mid)
	buff.WriteString("&time=")
	buff.WriteString(data.Time)
	buff.WriteString("&result=")
	buff.WriteString(data.Result)
	buff.WriteString("&cpOrder=")
	buff.WriteString(data.CpOrder)
	buff.WriteString("&ext=")
	buff.WriteString(data.Ext)
	buff.WriteString("&key=")
	buff.WriteString(config.Get(gameId + ".dl.payment_key"))
	return helpers.Md5String(buff.String())
}
