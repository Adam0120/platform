package mg

import (
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	myhttp "github.com/Adam0120/sdk-utils/http"
	json "github.com/json-iterator/go"
	"net/url"
)

type SessionVerifyData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Ukey      string `json:"ukey"`
		ChannelId int    `json:"channelId"`
	} `json:"data"`
}

var tokenUrl = "https://api.game.mgtv.com/fusion/cp/user/checkToken"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".mg."
	value := url.Values{}
	value.Add("ukey", s.Uid)
	value.Add("appId", config.Get(conPre+"app_id"))
	value.Add("token", s.Session)
	value.Add("time", app.GetStringUnix())
	sign := wtplatform.Sign(value, "&key="+config.Get(conPre+"secret_key"))
	value.Add("sign", sign)
	res, err := myhttp.FormRequest().SetBody(value.Encode()).Post(tokenUrl)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.Code != 0 {
		err = fmt.Errorf("session验证失败，错误码:%s", m.Msg)
		return
	}

	if m.Data.Ukey != s.Uid {
		err = fmt.Errorf("session验证失败，错误码:%s", m.Msg)
		return
	}

	return
}

func OrderSign(secretKey string, params map[string]any) string {
	return wtplatform.MapSign(params, "&key="+secretKey)
}
