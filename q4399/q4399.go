package q4399

import (
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	json "github.com/json-iterator/go"
	"strings"
)

type SessionVerifyData struct {
	Code   string `json:"code"`
	Msg    string `json:"message"`
	Result struct {
		Uid string `json:"uid"`
	} `json:"result"`
}

var tokenUrl = "http://m.4399api.com/openapi/oauth-check.html"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".q4399."
	query := "?state=" + s.Session + "&uid=" + s.Uid + "&key=" + config.Get(conPre+"app_id")
	res, err := myhttp.Request().Get(tokenUrl + query)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.Code != "100" {
		err = fmt.Errorf("session验证失败，错误码:%s,错误信息：%s", m.Code, m.Msg)
		return
	}
	return
}

func OrderSign(d map[string]string, gameId string) string {
	var builder strings.Builder

	builder.WriteString(d["orderid"])
	builder.WriteString(d["uid"])
	builder.WriteString(d["money"])
	builder.WriteString(d["gamemoney"])
	builder.WriteString(d["serverid"])
	builder.WriteString(config.Get(gameId + ".q4399.app_secret"))
	builder.WriteString(d["mark"])
	builder.WriteString(d["roleid"])
	builder.WriteString(d["time"])
	builder.WriteString(d["coupon_mark"])
	builder.WriteString(d["coupon_money"])

	return helpers.Md5String(builder.String())
}
