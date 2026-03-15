package oppo

import (
	"crypto"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/Adam0120/sdk-utils/app"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/crypt"
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/go-pay/crypto/xpem"
	"strconv"
	"time"
)

var oppoNitroUrl = "https://iopen.game.oppomobile.com/sdkopen/v2/cp/deliveryNotify"

func GetOrderSign(pram map[string]string, gameId string) error {
	str := "notifyId=" + pram["notifyId"] + "&" +
		"partnerOrder=" + pram["partnerOrder"] + "&" +
		"productName=" + pram["productName"] + "&" +
		"productDesc=" + pram["productDesc"] + "&" +
		"price=" + pram["price"] + "&" +
		"count=" + pram["count"] + "&" +
		"attach=" + pram["attach"]
	// 密钥为oppo公共公钥
	return verifySignCert(str, pram["sign"], config.Get(gameId+".oppo.public_key"))
}

func verifySignCert(signData, sign, publicKeyStr string) (err error) {
	publicKey, err := xpem.DecodePublicKey([]byte(publicKeyStr))
	if err != nil {
		return err
	}
	signBytes, _ := base64.StdEncoding.DecodeString(sign)
	hashs := crypto.SHA1
	h := hashs.New()
	h.Write([]byte(signData))
	if err = rsa.VerifyPKCS1v15(publicKey, hashs, h.Sum(nil), signBytes); err != nil {
		return fmt.Errorf("[%s]: %v", "签名效验失败", err)
	}
	return nil
}
func Deliver(pram map[string]string, appId string) error {
	tm := app.TimenowInTimezone()
	t := strconv.FormatInt(tm.UnixMilli(), 10)
	str := `{"orderId":"` + pram["notifyId"] + `","cpOrderId":"` + pram["partnerOrder"] + `","msg":"ok","sendPropsTime":"` + tm.Format(time.DateTime) + `","sendPropsRole":"` + pram["userId"] + `"}`
	data := crypt.AesEncrypt(str, config.Get(appId + ".oppo.app_secret")[:16])
	str = "client=" + `{"pkg":"` + config.Get(appId+".oppo.pkg_name") + `"}&data=` + data + "&t=" + t + "&"
	sign := crypt.SignCert(str, config.Get(appId+".oppo.private_key"))
	json := `{
"t": ` + t + `, 
"data": "` + data + `", 
"sign": "` + sign + `",
"client": { "pkg": "` + config.Get(appId+".oppo.pkg_name") + `" }
}`
	_, err := myhttp.JsonRequest().SetBody(json).Post(oppoNitroUrl)
	if err != nil {
		return err
	}
	return nil
}
