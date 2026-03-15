package qq

import (
	"encoding/hex"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	json "github.com/json-iterator/go"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

type SessionVerifyData struct {
	Ret int    `json:"ret"`
	Msg string `json:"msg,omitempty"`
}

type Date struct {
	Success bool   `json:"success"`
	Openid  string `json:"openid"`
}

var tokenUrl = "https://ysdk.qq.com"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".qq."
	postData := url.Values{}
	ts := app.GetStringUnix()
	postData.Add("timestamp", ts)
	postData.Add("appid", config.Get(conPre+"app_id"))
	postData.Add("openid", s.Uid)
	postData.Add("openkey", s.Session)
	appKey := ""
	qqUrl := tokenUrl
	if s.LoginType == "qq" {
		appKey = config.Get(conPre + "app_key")
		qqUrl += "/auth/qq_check_token?"
	} else if s.LoginType == "wx" {
		appKey = config.Get(conPre + "wx_app_secret")
		qqUrl += "/auth/wx_check_token?"
	} else {
		err = fmt.Errorf("登录参数错误")
		return
	}
	postData.Add("sig", helpers.Md5String(appKey+ts))

	res, err := myhttp.Request().Get(qqUrl + postData.Encode())
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.Ret != 0 {
		err = fmt.Errorf("session验证失败，错误码:%s", m.Msg)
		return
	}
	return
}

func OrderSign(pram map[string]string, gameId string) string {
	var builder strings.Builder

	builder.WriteString("GET&")
	builder.WriteString(url.QueryEscape("/order/qq/notify"))
	builder.WriteString("&")

	keys := make([]string, 0)
	for k := range pram {
		if k == "sig" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys) // 对键进行排序

	for i, k := range keys {
		if i > 0 {
			builder.WriteString("&")
		}
		builder.WriteString(k)
		builder.WriteString("=")
		builder.WriteString(RegexpS(pram[k]))
	}

	str := builder.String()
	str = url.QueryEscape(str)
	key := config.Get(gameId+".qq.app_secret") + "&"

	return helpers.HMACSHA1(str, key)
}

func RegexpS(str2 string) string {
	pat := `[^a-zA-Z0-9!\(\)*]{1,1}`
	f := func(s string) string {
		return "%" + strings.ToUpper(hex.EncodeToString([]byte(s)))
	}
	re, _ := regexp.Compile(pat)
	return re.ReplaceAllStringFunc(str2, f)
}
