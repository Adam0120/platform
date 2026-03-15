package wk

import (
	"bytes"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	json "github.com/json-iterator/go"
	"github.com/spf13/cast"
	"net/url"
	"sort"
)

const WK_URL = "http://unapi.jhsdk.wankatj.com/api/cp/user/check"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".wk."
	appId := config.Get(conPre + "app_id")
	key := config.Get(conPre + "app_key")
	vals := url.Values{}
	vals.Set("app_id", appId)
	vals.Set("mem_id", s.Uid)
	vals.Set("user_token", s.Session)
	vals.Set("sign", wtplatform.Sign(vals, "&app_key="+key))

	res, err := myhttp.FormRequest().SetBody(vals.Encode()).Post(WK_URL)
	if err != nil {
		return
	}
	var m map[string]interface{}
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if cast.ToInt(m["status"]) != 1 {
		err = fmt.Errorf("session验证失败，错误码:%s", cast.ToString(m["msg"]))
		return
	}
	return
}

func OrderSign(appKey string, params map[string]string) string {
	var keys []string
	newparams := make(map[string]string, len(params))
	for k, v := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
		newparams[k] = url.QueryEscape(v)
	}
	sort.Strings(keys) // 对键进行排序
	buf := bytes.Buffer{}
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s=%s&", k, cast.ToString(newparams[k])))

	}
	buf.WriteString("app_key=")
	buf.WriteString(appKey)
	return helpers.Md5String(buf.String())
}
