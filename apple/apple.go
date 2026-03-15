package apple

import (
	myhttp "github.com/Adam0120/sdk-utils/http"
	"github.com/go-pay/gopay/apple"
	json "github.com/json-iterator/go"
)

const (
	UrlSandbox = "https://sandbox.itunes.apple.com/verifyReceipt"
	UrlProd    = "https://buy.itunes.apple.com/verifyReceipt"
)

type RefundParam struct {
	SignedPayload string `json:"signedPayload"`
}

type AppleReqData struct {
	BsTradeNo string `json:"bs_trade_no"`
	Uid       string `json:"uid"`
	ProductId string `json:"product_id"`
	Receipt   string `json:"receipt"`
	PayEnv    string `json:"pay_env"`
}

// VerifyReceipt 请求APP Store 校验支付请求,实际测试时发现这个文档介绍的返回信息只有那个status==0表示成功可以用，其他的返回信息跟文档对不上
// url：取 UrlProd 或 UrlSandbox
// pwd：苹果APP秘钥，https://help.apple.com/app-store-connect/#/devf341c0f01
// 文档：https://developer.apple.com/documentation/appstorereceipts/verifyreceipt
func VerifyReceipt(url, pwd, receipt string) (rsp *apple.VerifyResponse, err error) {
	req := &apple.VerifyRequest{Receipt: receipt, Password: pwd}
	rsp = new(apple.VerifyResponse)
	post, err := myhttp.FormRequest().SetBody(req).Post(url)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(post.Body(), &rsp)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}
