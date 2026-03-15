package lx

import (
	"encoding/xml"
	"fmt"
	wtplatform "github.com/Adam0120/platform"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/crypt"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/Adam0120/sdk-utils/models/channel_shop"
	"github.com/Adam0120/sdk-utils/models/order"
)

type SessionVerifyData struct {
	IdentityInfo xml.Name `xml:"IdentityInfo"`
	AccountID    string   `xml:"AccountID"`
	Username     string   `xml:"Username"`
	DeviceID     string   `xml:"DeviceID"`
}

var tokenUrl = "http://passport.lenovo.com/interserver/authen/1.2/getaccountid"

type Verify struct{}

// SessionVerify Session验证
func (sv *Verify) SessionVerify(s *wtplatform.SessionVerifyRequest) (err error) {
	conPre := s.AppId + ".lx."
	query := "?lpsust=" + s.Session + "&realm=" + config.Get(conPre+"app_id")
	res, err := myhttp.Request().Get(tokenUrl + query)
	if err != nil {
		return
	}
	m := SessionVerifyData{}
	err = xml.Unmarshal(res.Body(), &m)
	if err != nil {
		return
	}

	if m.AccountID == "" {
		err = fmt.Errorf("session验证失败，错误信息:%s", res)
		return
	}
	return
}

type OrderData struct {
	Result    int    `json:"result"`
	TransType int    `json:"transtype"`
	Count     int    `json:"count"`
	WaresId   int    `json:"waresid"`
	PayType   int    `json:"paytype"`
	Money     int    `json:"money"`
	FeeType   int    `json:"feetype"`
	ExOrdErno string `json:"exorderno"`
	TransId   string `json:"transid"`
	CppRivAte string `json:"cpprivate"`
	Appid     string `json:"appid"`
	TransTime string `json:"transtime"`
}

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	params := make(map[string]interface{})
	data := channel_shop.GetByPlatformAndPrice(13, orderModel.PayFee)
	if data.ID == 0 {
		return nil, fmt.Errorf("找不到对应的商品")
	}
	params["shopId"] = data.ShopId
	return params, nil
}

func GetOrderSign(transData string) string {
	return crypt.SignCert(transData, config.Get("lx.private_key"))
}
