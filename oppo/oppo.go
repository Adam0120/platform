package oppo

import (
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	json "github.com/json-iterator/go"
	"net/url"

	"strconv"
)

type SessionVerifyData struct {
	ResultCode   string `json:"resultCode"`
	ResultMsg    string `json:"resultMsg"`
	LoginToken   string `json:"loginToken,omitempty"`
	Ssoid        string `json:"ssoid,omitempty"`
	AppKey       string `json:"appKey,omitempty"`
	UserName     string `json:"userName,omitempty"`
	Email        string `json:"email,omitempty"`
	MobileNumber string `json:"mobileNumber,omitempty"`
	CreateTime   string `json:"createTime,omitempty"`
	UserStatus   string `json:"userStatus,omitempty"`
}

var tokenUrl = "https://iopen.game.oppomobile.com/sdkopen/user/fileIdInfo"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".oppo."
	token := url.QueryEscape(s.Session)
	header := getTokenSign(token, conPre)
	query := "?fileId=" + s.Uid + "&token=" + token
	res, err := myhttp.Request().SetHeaders(header).Get(tokenUrl + query)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.ResultCode != "200" {
		err = fmt.Errorf("session验证失败，错误码：%s，错误信息:%s", m.ResultCode, m.ResultMsg)
		return
	}
	return
}

func getTokenSign(token, conPre string) map[string]string {
	requestString := "oauthConsumerKey=" + config.Get(conPre+"app_key") +
		"&oauthToken=" + token +
		"&oauthSignatureMethod=HMAC-SHA1&oauthTimestamp=" +
		strconv.FormatInt(app.TimenowInTimezone().UnixMilli(), 10) +
		"&oauthNonce=7" + strconv.FormatInt(app.TimenowInTimezone().Unix(), 10) + "&oauthVersion=1.0&"
	key := config.Get(conPre+"app_secret") + "&"
	return map[string]string{
		"param":          requestString,
		"oauthsignature": url.QueryEscape(helpers.HMACSHA1(requestString, key)),
	}
}
