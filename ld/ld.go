package ld

import (
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	json "github.com/json-iterator/go"
	"strings"
)

type SessionVerifyData struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	PopupsType string `json:"popupsType"`
}

var tokenUrl = "https://ldapi.ldmnq.com/ext/loginverify"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".ld."
	data := make(map[string]string)
	data["gameid"] = config.Get(conPre + "app_id")
	data["useruid"] = s.Uid
	data["usertoken"] = s.Session
	data["timestamp"] = app.TimenowInTimezone().Format("20060102150405")
	data["appkey"] = config.Get(conPre + "app_key")
	data["sign"] = computeSign(helpers.ToJson(data))
	res, err := myhttp.JsonRequest().SetBody(data).Post(tokenUrl)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.Code != 0 {
		err = fmt.Errorf("session验证失败，错误码:%d,错误信息：%s", m.Code, m.Message)
		return
	}
	return
}

type OrderData struct {
	Amount       string `xml:"amount"`
	OrderId      string `xml:"orderId"`
	UserId       string `xml:"userId"`
	RoleId       string `xml:"roleId"`
	ReturnCode   string `xml:"return_code"`
	OutOrderId   string `xml:"out_order_id"`
	GameServerId string `xml:"game_server_id"`
	Sign         string `xml:"sign"`
}

func OrderSign(data OrderData, gameId string) string {
	var builder strings.Builder
	builder.WriteString("amount=")
	builder.WriteString(data.Amount)
	builder.WriteString("&game_server_id=")
	builder.WriteString(data.GameServerId)
	builder.WriteString("&orderId=")
	builder.WriteString(data.OrderId)
	builder.WriteString("&out_order_id=")
	builder.WriteString(data.OutOrderId)
	builder.WriteString("&returnCode=")
	builder.WriteString(data.ReturnCode)
	builder.WriteString("&roleId=")
	builder.WriteString(data.RoleId)
	builder.WriteString("&userId=")
	builder.WriteString(data.UserId)
	builder.WriteString("&key=")
	builder.WriteString(config.Get(gameId + ".ld.app_key"))
	return computeSign(builder.String())
}

func computeSign(input string) string {
	return helpers.ToUpper(helpers.Md5String(input))
}
