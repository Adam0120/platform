package huawei

import (
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	myhttp "github.com/Adam0120/sdk-utils/http"
	json "github.com/json-iterator/go"
	"net/url"
)

type SessionVerifyData struct {
	RtnCode  int    `json:"rtnCode"`
	UnionId  string `json:"unionId,omitempty"`
	OpenId   string `json:"openId,omitempty"`
	Expire   int    `json:"expire,omitempty"`
	ClientId string `json:"clientId,omitempty"`
	PlayerId string `json:"playerId,omitempty"`
}
type UserInfoData struct {
	OpenID         string `json:"openID"`
	Email          string `json:"email"`
	DisplayName    string `json:"displayName"`
	HeadPictureURL string `json:"headPictureURL"`
}

var tokenUrl = "https://jos-open-api.cloud.huawei.com/gameservice/api/gbClientApi"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	postData := url.Values{}
	postData.Add("method", "external.hms.gs.getTokenInfo")
	postData.Add("accessToken", s.Session)
	res, err := myhttp.FormRequest().SetBody(postData.Encode()).Post(tokenUrl)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if m.RtnCode != 0 {
		err = fmt.Errorf("session验证失败,错误信息：%s", res)
		return
	}
	if s.Uid != m.OpenId {
		err = fmt.Errorf("pid 验证失败,错误信息：%s", m.OpenId)
	}
	return
}
