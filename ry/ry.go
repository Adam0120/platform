package ry

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/google/uuid"
	json "github.com/json-iterator/go"
)

type SessionVerifyData struct {
	ErrorCode    int      `json:"errorCode"`
	ErrorMessage string   `json:"errorMessage"`
	Data         UserInfo `json:"data"`
}

type UserInfo struct {
	OpenId         string `json:"openId"`
	UnionId        string `json:"unionId"`
	DisplayName    string `json:"displayName"`
	HeadPictureURL string `json:"headPictureURL"`
	HasRealName    bool   `json:"hasRealName"`
	IsAdult        bool   `json:"isAdult"`
	Age            int    `json:"age"`
	AccessToken    string `json:"accessToken"`
}

var tokenUrl = "https://gamecenter-api.cloud.honor.com/game/service/cp/v1/user/auth"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".ry."
	m := map[string]string{
		"token": s.Session,
	}
	jsonBody, err := json.Marshal(m)
	if err != nil {
		return
	}
	appId := config.Get(conPre + "app_id")
	u, _ := uuid.NewUUID()
	traceId := u.String()
	ts := app.GetStringUnixMilli()
	unionToken := s.Session
	secretKey := config.Get(conPre + "secret_key")
	res, err := myhttp.JsonRequest().SetHeaders(map[string]string{
		"x-app-id":       appId,
		"x-sign-type":    "HmacSha256",
		"x-ra-traceid":   traceId,
		"x-ra-timestamp": ts,
		"x-union-token":  unionToken,
		"x-sign-value":   Sign(secretKey, ts, traceId, string(jsonBody)),
	}).SetBody(jsonBody).Post(tokenUrl)
	if err != nil {
		return
	}
	resJson := &SessionVerifyData{}
	err = json.Unmarshal(res.Body(), resJson)
	if err != nil {
		return
	}

	if resJson.ErrorCode != 0 {
		err = fmt.Errorf("session验证失败，错误码:%s", resJson.ErrorMessage)
		return
	}

	if s.Uid != resJson.Data.OpenId {
		err = fmt.Errorf("账号不匹配")
		return
	}

	return
}

func Sign(secretKey, ts, traceId, jsonBody string) string {
	signStr := fmt.Sprintf("{\"x-ra-timestamp\":\"%s\",\"x-ra-traceid\":\"%s\"}%s", ts, traceId, jsonBody)
	hash := hmac.New(sha256.New, []byte(secretKey))
	hash.Write([]byte(signStr))
	sign := hex.EncodeToString(hash.Sum(nil))
	return sign
}
