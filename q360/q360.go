package q360

import (
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	myhttp "github.com/Adam0120/sdk-utils/http"
	json "github.com/json-iterator/go"
	"net/url"
	"sort"
	"strings"

	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
)

type SessionVerifyData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Avatar    string `json:"avatar"`
	ErrorCode string `json:"error_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

type Date struct {
	Success bool   `json:"success"`
	Openid  string `json:"openid"`
}

var tokenUrl = "https://openapi.360.cn/user/me.json?"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	pData := url.Values{}
	pData.Add("access_token", s.Session)
	res, err := myhttp.Request().Get(tokenUrl + pData.Encode())
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.ErrorCode != "" {
		err = fmt.Errorf("session验证失败，错误码:%s, %s", m.ErrorCode, m.Error)
		return
	}
	return
}

func OrderSign(pram map[string]string, gameId string) string {
	var strBuilder strings.Builder
	keys := make([]string, 0)
	for k, v := range pram {
		if (k == "sign" || k == "sign_return") || v == "" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys) // 对键进行排序
	for _, k := range keys {
		strBuilder.WriteString(pram[k])
		strBuilder.WriteString("#")
	}

	strBuilder.WriteString(config.Get(gameId + ".q360.app_secret"))
	return helpers.Md5String(strBuilder.String())
}
