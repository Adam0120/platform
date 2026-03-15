package ry

import (
	"errors"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/crypt"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/wtredis"
	"github.com/google/uuid"
	json "github.com/json-iterator/go"
	"net/url"
	"time"
)

type RyOrderData struct {
	Env       string `json:"env"`
	EventType string `json:"eventType"`
	Version   string `json:"version"`
	EventTime string `json:"eventTime"`
	//NotificationMessage NotificationMessage `json:"notificationMessage"`
	Data string `json:"data"`
}

type NotificationMessage struct {
	PurchaseToken string `json:"purchaseToken"`
	ProductId     string `json:"productId"`
}

var redisKey = "RY_PAY_ACCESS_TOKEN"
var tokenUrl1 = "https://hnoauth-login-drcn.cloud.hihonor.com/oauth2/v3/token"
var versionTokenUrl = "https://iap-api-drcn.cloud.hihonor.com/iap/server/verifyToken"
var consumeProductUrl = "https://iap-api-drcn.cloud.hihonor.com/iap/server/consumeProduct"

// GetAccessToken 应用级accessToken获取方法
func GetAccessToken(appId, clientSecret string) (string, error) {
	accessToken := wtredis.Redis.Get(redisKey)
	if accessToken == "" {
		newToken, expiresIn, err := getAccessToken(tokenUrl1, appId, clientSecret)
		if err != nil {
			return "", err
		}
		wtredis.Redis.Set(redisKey, newToken, time.Duration(expiresIn)*time.Second)
		return newToken, nil
	}
	return accessToken, nil
}

func getAccessToken(url2, appId, clientSecret string) (string, int, error) {
	vals := url.Values{}
	vals.Add("grant_type", "client_credentials")
	vals.Add("client_secret", clientSecret)
	vals.Add("client_id", appId)
	res, err := myhttp.FormRequest().SetBody(vals.Encode()).Post(url2)
	if err != nil {
		return "", 0, err
	}
	resJson := map[string]interface{}{}
	err = json.Unmarshal(res.Body(), &resJson)
	if err != nil {
		return "", 0, err
	}

	if _, ok := resJson["error"]; ok {
		return "", 0, errors.New(resJson["error_description"].(string))
	}
	accessToken := resJson["access_token"]
	expiresIn := resJson["expires_in"]
	return accessToken.(string), int(expiresIn.(float64)), nil
}

type PurchaseTokenVersionRes struct {
	Code    int                      `json:"code"`
	Message string                   `json:"message"`
	Data    PurchaseTokenVersionData `json:"data"`
}

type PurchaseTokenVersionData struct {
	PurchaseProductInfo string `json:"purchaseProductInfo"`
	DataSig             string `json:"dataSig"`
	SigAlgorithm        string `json:"sigAlgorithm"`
}

type PurchaseProductInfo struct {
	AppId            string  `json:"appId"`
	OrderId          string  `json:"orderId"`
	BizOrderNo       string  `json:"bizOrderNo"`
	ProductType      int     `json:"productType"`
	ProductId        string  `json:"productId"`
	ProductName      string  `json:"productName"`
	PurchaseTime     int64   `json:"purchaseTime"`
	PurchaseState    int     `json:"purchaseState"`
	ConsumptionState int     `json:"consumptionState"`
	PurchaseToken    string  `json:"purchaseToken"`
	Currency         string  `json:"currency"`
	Price            float64 `json:"price"`
	DeveloperPayload string  `json:"developerPayload"`
	DisplayPrice     string  `json:"displayPrice"`
	OriOrder         string  `json:"oriOrder"`
}

// PurchaseTokenVersion purchaseToken验证
func PurchaseTokenVersion(accessToken, purchaseToken, gameId string) (*PurchaseTokenVersionRes, error) {
	appId := config.Get(gameId + ".ry.app_id")
	publicKey := config.Get(gameId + ".ry.public_key")
	return purchaseTokenVersion(versionTokenUrl, appId, publicKey, accessToken, purchaseToken)
}

func purchaseTokenVersion(url2, appId, publicKey, accessToken, purchaseToken string) (*PurchaseTokenVersionRes, error) {
	res, err := myhttp.Request().SetHeaders(map[string]string{
		"Content-Type":  "application/json",
		"access-token":  accessToken,
		"x-iap-appid":   appId,
		"purchaseToken": purchaseToken,
	}).Post(url2)
	if err != nil {
		return nil, err
	}
	data := &PurchaseTokenVersionRes{}
	err = json.Unmarshal(res.Body(), data)
	if err != nil {
		return nil, err
	}
	if data.Code != 0 {
		return nil, errors.New(data.Message)
	}
	err = crypt.RSAVerify([]byte(data.Data.PurchaseProductInfo), data.Data.DataSig, publicKey)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// ConsumeProduct consumeProduct 商品消耗
func ConsumeProduct(purchaseToken, accessToken, gameId string) error {
	publicKey := config.Get(gameId + ".ry.public_key")
	appId := config.Get(gameId + ".ry.app_id")
	return consumeProduct(consumeProductUrl, appId, publicKey, purchaseToken, accessToken)
}

func consumeProduct(url2, appId, publicKey, purchaseToken, accessToken string) error {
	u := uuid.New()
	m := map[string]string{
		"purchaseToken":      purchaseToken,
		"developerChallenge": u.String(),
	}
	bb, err := json.Marshal(m)
	if err != nil {
		return err
	}
	res, err := myhttp.Request().SetHeaders(map[string]string{
		"Content-Type":  "application/json",
		"access-token":  accessToken,
		"x-iap-appid":   appId,
		"purchaseToken": purchaseToken,
	}).SetBody(bb).Post(url2)
	if err != nil {
		return err
	}
	data := &PurchaseTokenVersionRes{}
	err = json.Unmarshal(res.Body(), data)
	if err != nil {
		return err
	}
	if data.Code != 0 {
		return errors.New(data.Message)
	}
	err = crypt.RSAVerify([]byte(data.Data.PurchaseProductInfo), data.Data.DataSig, publicKey)
	if err != nil {
		return err
	}
	return nil
}
