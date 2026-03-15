package kp

import (
	"bytes"
	"errors"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/logger"
	"github.com/Adam0120/sdk-utils/models/channel_shop"
	"github.com/Adam0120/sdk-utils/models/order"
	"github.com/Adam0120/sdk-utils/models/user_access_token"
	json "github.com/json-iterator/go"
	"go.uber.org/zap"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type SessionVerifyData struct {
	Success    bool   `json:"success"`
	ErrCode    string `json:"errCode"`
	ErrMessage string `json:"errMessage"`
	Data       struct {
		OpenId       string `json:"openId"`
		AccessToken  string `json:"accessToken"`
		RefreshToken string `json:"refreshToken"`
		ExpiredIn    int64  `json:"expiredIn"`
		Scope        string `json:"scope"`
	} `json:"data"`
}

var tokenUrl = "https://auth.api.coolpad.com/oauth/authorize"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".kp."
	reqVersionData := url.Values{}
	appId := config.Get(conPre + "app_id")
	appKey := config.Get(conPre + "app_key")
	ts := app.GetStringUnixMilli()
	reqVersionData.Add("appId", appId)
	reqVersionData.Add("timestamp", ts)
	reqVersionData.Add("authCode", s.Session)
	reqVersionData.Add("grantType", "authorization_code")
	originString := bytes.Buffer{}
	originString.WriteString("appId=")
	originString.WriteString(appId)
	originString.WriteString("&authCode=")
	originString.WriteString(s.Session)
	originString.WriteString("&grantType=authorization_code&timestamp=")
	originString.WriteString(ts)
	originString.WriteString("&key=")
	originString.WriteString(appKey)
	reqVersionData.Add("sign", helpers.Md5String(originString.String()))
	res, err := myhttp.Request().Get(tokenUrl + "?" + reqVersionData.Encode())
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}
	if !m.Success {
		err = fmt.Errorf("session验证失败，错误码:%s", m.ErrCode)
		return
	}
	tk := &user_access_token.UserAccessToken{
		OpenId:       m.Data.OpenId,
		AccessToken:  m.Data.AccessToken,
		RefreshToken: m.Data.RefreshToken,
		ExpiredIn:    m.Data.ExpiredIn,
	}
	err = tk.CreateOrUpdate()
	if err != nil {
		return err
	}
	return
}

type Body struct {
	Amount    string `json:"amount"`
	CpPrivate string `json:"cpPrivate,omitempty"`
	PayAmount string `json:"payAmount"`
	OrderId   string `json:"orderId"`
	PayTime   string `json:"payTime"`
	CpTradeId string `json:"cpTradeId"`
	AppId     int64  `json:"appId"`
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	// 初始化参数Map
	params := make(map[string]interface{})
	accessToken, err := GetAccessToken(orderModel.PlatUid, strconv.Itoa(orderModel.AppId))
	if err != nil {
		return nil, err
	}
	params["accessToken"] = accessToken
	params["openId"] = orderModel.PlatUid
	data := channel_shop.GetByPlatformAndPrice(19, orderModel.TotalFee)
	if data.ID == 0 {
		return nil, fmt.Errorf("找不到对应的商品")
	}
	params["shopId"] = data.ShopId
	return params, nil
}

func OrderSign(appId, appKey string, body *Body) string {
	sign := bytes.Buffer{}
	sign.WriteString("appId=")
	sign.WriteString(appId)
	sign.WriteString("&")
	bb, err := json.Marshal(body)
	if err != nil {
		return ""
	}
	sign.Write(bb)
	sign.WriteString("&")
	sign.WriteString("key=")
	sign.WriteString(appKey)
	return strings.ToUpper(helpers.Md5String(sign.String()))
}

func GetAccessToken(openId, aId string) (string, error) {
	uToken := user_access_token.UserAccessToken{}
	Token := uToken.GetByOpenId(openId)
	if Token.OpenId == "" {
		return "", errors.New("AccessTokenNil")
	}

	if time.Now().Unix() >= Token.ExpiredIn {
		err := freshToken(Token, aId)
		if err != nil {
			logger.Error("GetAccessToken", zap.Error(err))
			return "", err
		}
		err = Token.CreateOrUpdate()
		if err != nil {
			logger.Error("GetAccessToken", zap.Error(err))
			return "", err
		}
		return Token.AccessToken, nil
	}

	return Token.AccessToken, nil
}

func freshToken(token *user_access_token.UserAccessToken, aId string) error {
	//刷新token
	val := url.Values{}
	appId := config.Get(aId + ".kp.app_id")
	appKey := config.Get(aId + ".kp.app_key")
	baseUrl := config.Get(aId + ".kp.refresh_url")
	val.Add("appId", appId)
	ts := app.GetStringUnixMilli()
	val.Add("timestamp", ts)
	val.Add("refreshToken", token.RefreshToken)
	val.Add("grantType", "refresh_token")
	val.Add("sign", accessTokenSign(appId, appKey, ts, token.RefreshToken))
	res, err := myhttp.Request().Get(baseUrl + "?" + val.Encode())
	if err != nil {
		return err
	}
	t := map[string]string{}
	err = json.Unmarshal(res.Body(), &t)
	if err != nil {
		return err
	}
	expiredIn, err := strconv.ParseInt(t["expiredIn"], 10, 64)
	if err != nil {
		return err
	}
	token.OpenId = t["openId"]
	token.AccessToken = t["accessToken"]
	token.RefreshToken = t["refreshToken"]
	token.ExpiredIn = expiredIn
	return nil
}

func accessTokenSign(appId, appKey, ts, refreshToken string) string {
	originString := bytes.Buffer{}
	originString.WriteString("appId=")
	originString.WriteString(appId)
	originString.WriteString("&grantType=")
	originString.WriteString("refresh_token")
	originString.WriteString("&refreshToken=")
	originString.WriteString(url.QueryEscape(refreshToken))
	originString.WriteString("&timestamp=")
	originString.WriteString(ts)
	originString.WriteString("&key=")
	originString.WriteString(appKey)
	return helpers.Md5String(originString.String())
}

func getChannelShop(price int) string {
	data := channel_shop.GetByPlatformAndPrice(19, price)
	if data.ID == 0 {
		return ""
	}
	return data.ShopId
}
