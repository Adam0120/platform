package uu

import (
	"bytes"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/crypt"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/models/order"
	json "github.com/json-iterator/go"
	"strconv"
)

type VerifyResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

var tokenUrl = "https://yofun.163.com/api/v1/verify"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".ry."
	appId := config.Get(conPre + "app_id")
	d := map[string]string{
		"app_id":        appId,
		"user_id":       s.Uid,
		"channel_token": s.Session,
	}
	res, err := myhttp.JsonRequest().SetBody(d).Post(tokenUrl)
	if err != nil {
		return
	}
	var m VerifyResponse
	err = json.Unmarshal(res.Body(), &m)
	if m.Code != 0 {
		err = fmt.Errorf("code:%d", m.Code)
		return
	}
	return
}

type UUOrderCallback struct {
	OrderId     int    `json:"order_id"`
	GameOrderId string `json:"game_order_id"`
	AppId       string `json:"app_id"`
	UserId      string `json:"user_id"`
	Status      int    `json:"status"`
	OrderPrice  int    `json:"order_price"`
}

func OrderSign(appId, sign string, body []byte) (u UUOrderCallback, ok bool) {
	buf := bytes.Buffer{}
	buf.WriteString(`/order/uu/notify/`)
	buf.WriteString(appId)
	buf.WriteString("?")
	buf.Write(body)
	pk := config.Get(appId + ".uu.public_key")
	err := crypt.RSAVerify([]byte(buf.String()), sign, pk)
	if err != nil {
		return
	}
	ok = true
	err = json.Unmarshal(body, &u)
	return
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	params := make(map[string]any)
	params["callback_url"] = config.Get("app.url") + "/order/uu/notify/" + strconv.Itoa(orderModel.AppId)
	return params, nil
}
