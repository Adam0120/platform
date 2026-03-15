package huawei

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/Adam0120/sdk-utils/config"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/wtredis"
	json "github.com/json-iterator/go"
	"net/url"
	"time"
)

type AtResponse struct {
	AccessToken string `json:"access_token"`
}

type NotificationResponse struct {
	ErrorCode string `json:"responseCode"`
	ErrorMsg  string `json:"responseMessage"`
}

type AtClient struct {
}

var AtDemo = &AtClient{}

func (atDemo *AtClient) GetAppAt(appId string) (string, error) {
	AccessToken := wtredis.GetToString("huawei:appAt:" + appId)
	if AccessToken != "" {
		return AccessToken, nil
	}
	urlValue := url.Values{"grant_type": {"client_credentials"}, "client_secret": {config.Get(appId + ".hw.app_secret")}, "client_id": {config.Get(appId + ".hw.app_id")}}
	resp, err := myhttp.FormRequest().SetBody(urlValue.Encode()).Post("https://oauth-login.cloud.huawei.com/oauth2/v3/token")
	if err != nil {
		return "", err
	}
	var atResponse AtResponse
	json.Unmarshal(resp.Body(), &atResponse)
	if atResponse.AccessToken != "" {
		wtredis.SetString("huawei:appAt:"+appId, atResponse.AccessToken, time.Second*3000)
		return atResponse.AccessToken, nil
	} else {
		return "", errors.New("Get token fail, " + string(resp))
	}
}

func BuildAuthorization(appId string) (string, error) {
	appAt, err := AtDemo.GetAppAt(appId)
	if err != nil {
		return "", err
	}
	oriString := fmt.Sprintf("APPAT:%s", appAt)
	var authString = base64.StdEncoding.EncodeToString([]byte(oriString))
	var authHeaderString = fmt.Sprintf("Basic %s", authString)
	return authHeaderString, nil
}

func ConfirmPurchase(bodyMap map[string]string) error {
	bodyBytes, err := SendRequest("https://orders-drcn.iap.cloud.huawei.com.cn/applications/v2/purchases/confirm", bodyMap)
	if err != nil {
		return errors.New("ConfirmPurchase: " + err.Error())
	}
	var atResponse NotificationResponse
	err = json.Unmarshal(bodyBytes, &atResponse)
	if err != nil {
		return err
	}
	if atResponse.ErrorCode != "0" {
		return errors.New(atResponse.ErrorMsg)
	}
	return nil
}

func SendRequest(url string, bodyMap map[string]string) ([]byte, error) {
	authHeaderString, err := BuildAuthorization(bodyMap["appId"])
	if err != nil {
		return nil, err
	}
	delete(bodyMap, "appId")
	res, err := myhttp.Request().SetHeaders(map[string]string{
		"Authorization": authHeaderString,
		"Content-Type":  "application/json; charset=UTF-8",
	}).SetBody(bodyMap).Post(url)
	return res.Body(), err
}

func VerifyRsaSign(content, sign, appId string) error {
	publicKeyByte, err := base64.StdEncoding.DecodeString(config.Get(appId + ".hw.public_key"))
	if err != nil {
		return err
	}
	pub, err := x509.ParsePKIXPublicKey(publicKeyByte)
	if err != nil {
		return err
	}
	hashed := sha256.Sum256([]byte(content))
	signature, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(pub.(*rsa.PublicKey), crypto.SHA256, hashed[:], signature)
}
