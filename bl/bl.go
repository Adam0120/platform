package bl

import (
	"bytes"
	"fmt"
	"net/url"
	"sort"
	"strconv"

	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/logger"
	"github.com/Adam0120/sdk-utils/models/order"
	json "github.com/json-iterator/go"
	"go.uber.org/zap"
)

type SessionVerifyData struct {
	Code      int    `json:"code"`
	OpenId    int    `json:"open_id"`
	Uname     string `json:"uname"`
	RequestId string `json:"requestId"`
}

type AccountCloseData struct {
	Data struct {
		VoucherNo string `json:"voucher_no"`
		Uid       int    `json:"uid"`
		GameId    int    `json:"game_id"`
		Sign      string `json:"sign"`
	} `form:"data"`
}

var baseUrl = "http://line3-game-api-adapter-sh.biligame.net/api/server/"

func (a *AccountCloseData) AccountClose() string {
	logger.Info("blAccountClose", zap.Any("data", a.Data))
	return "success"
}

type Verify struct{}

func (sv *Verify) SignOrderString(orderModel order.Order) string {
	str := "1" + strconv.Itoa(orderModel.TotalFee) + orderModel.BsTradeNo
	return helpers.Md5String(str + config.Get(strconv.Itoa(orderModel.AppId)+".bl.app_secret"))
}

func SearchOrder(params url.Values, gameId string) (map[string]string, error) {
	conPre := gameId + ".bl."
	res, err := biliPost(conPre, "query.pay.order", params)
	if err != nil {
		return nil, err
	}
	return helpers.JsonToMapString(res), nil
}
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".bl."
	postData := url.Values{}
	postData.Add("uid", s.Uid)
	postData.Add("access_key", s.Session)
	res, err := biliPost(conPre, "session.verify", postData)
	if err != nil {
		return
	}
	var m SessionVerifyData
	err = json.Unmarshal([]byte(res), &m)
	if err != nil {
		return
	}
	if m.Code != 0 {
		err = fmt.Errorf("session验证失败，错误码:%d", m.Code)
		return
	}
	if s.Uid != strconv.Itoa(m.OpenId) {
		err = fmt.Errorf("uid效验失败")
		return
	}
	return
}

func Sign(params url.Values, secretKey string) string {
	var keys []string
	for k := range params {
		if k == "item_name" || k == "item_desc" || k == "token" || k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys) // 对键进行排序
	buf := bytes.Buffer{}
	for _, k := range keys {
		buf.WriteString(params.Get(k))
	}
	buf.WriteString(secretKey)
	return helpers.Md5String(buf.String())
}

func biliPost(conPre, uri string, postData url.Values) (resp string, err error) {
	postData.Add("game_id", config.Get(conPre+"app_id"))
	postData.Add("timestamp", app.GetStringUnix())
	postData.Add("version", config.Get(conPre+"version"))
	postData.Add("server_id", config.Get(conPre+"server_id"))
	postData.Add("merchant_id", config.Get(conPre+"merchant_id"))
	postData.Add("sign", Sign(postData, config.Get(conPre+"app_secret")))

	body, err := myhttp.FormRequest().SetHeader("User-Agent", "Mozilla/5.0 GameServer").
		SetBody(postData.Encode()).Post(baseUrl + uri)
	resp = body.String()
	return
}
