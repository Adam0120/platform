// Package mywxpay 微信支付
package mywxpay

import (
	"context"
	"errors"
	"fmt"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/conv"
	"github.com/Adam0120/sdk-utils/logger"
	"github.com/Adam0120/sdk-utils/models/order"
	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/wechat/v3"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"sync"
	"time"
)

var WxLoginPass = new(MobileWxPassport)

// WechatClient Wechat 服务
type WechatClient struct {
	Client  *wechat.ClientV3
	Context context.Context
}

// once 确保全局的 Wechat 对象只实例一次
var once sync.Once

// Wechat 全局 WechatClient
var wechatCli *WechatClient

// ConWechat 连接 Wechat
func ConWechat() *WechatClient {
	once.Do(func() {
		wechatCli = newClient()
	})
	return wechatCli
}

// newClient 创建一个新的 Wechat 连接
func newClient() *WechatClient {
	// 初始化自定的 WechatClient 实例
	wec := &WechatClient{}
	// 使用默认的 context
	wec.Context = context.Background()
	var err error
	//MchId, SerialNo, APIv3Key, PrivateKeyContent
	wec.Client, err = wechat.NewClientV3(
		config.GetString("mywxpay.mch_id"),
		config.GetString("mywxpay.mch_certificate_serial_number"),
		config.GetString("mywxpay.mch_api_v3_key"),
		config.GetToPath("mywxpay.pem_path"),
	)
	logger.LogIf(err)
	err = wec.Client.AutoVerifySign()
	logger.LogIf(err)

	return wec
}

func (w *MobileWxPassport) V3NotifyReqDesc(req *http.Request) (*wechat.V3DecryptPayResult, error) {
	notify, err := wechat.V3ParseNotify(req)
	if err != nil {
		return nil, err
	}
	err = notify.VerifySignByPK(ConWechat().Client.WxPublicKey())
	if err != nil {
		return nil, err
	}
	res, err := notify.DecryptPayCipherText(config.GetString("mywxpay.mch_api_v3_key"))
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (w *MobileWxPassport) SearchOrder(c *gin.Context, orderNo string) (map[string]interface{}, error) {
	orderMap := make(map[string]interface{})
	ret, err := ConWechat().Client.V3TransactionQueryOrder(c, 2, orderNo)
	if err != nil {
		return orderMap, err
	}
	if ret.Code != 0 {
		return orderMap, fmt.Errorf("%s", ret.Error)
	}
	orderMap = map[string]interface{}{
		"trade_state":      CasePayStatus(ret.Response.TradeState),
		"trade_state_desc": ret.Response.TradeStateDesc,
		"transaction_id":   ret.Response.TransactionId,
		"pay_time":         ret.Response.SuccessTime,
		"trade_type":       ret.Response.TradeType,
		"pay_fee":          ret.Response.Amount.Total,
	}
	return orderMap, err
}

func (w *MobileWxPassport) CreateOrder(c *gin.Context, orderModel *order.Order) (orderInfo interface{}, err error) {
	appId := strconv.Itoa(orderModel.AppId)
	// 初始化参数Map
	bm := make(gopay.BodyMap)
	aId := config.GetString(appId+".mywxpay.app_id_"+orderModel.ChannelId, config.GetString(appId+".mywxpay.app_id"))
	bm.Set("appid", aId).
		Set("description", orderModel.GoodsName).
		Set("out_trade_no", orderModel.BsTradeNo).
		Set("time_expire", orderModel.ExpirationTime.Format(time.RFC3339)).
		Set("notify_url", config.GetString("mywxpay.pay_notify_url")+appId).
		SetBodyMap("amount", func(b gopay.BodyMap) {
			b.Set("total", orderModel.TotalFee).
				Set("currency", "CNY")
		}).SetBodyMap("scene_info", func(b gopay.BodyMap) {
		b.Set("payer_client_ip", c.ClientIP())
	})
	wxRsp, err := ConWechat().Client.V3TransactionApp(c, bm)
	if err != nil {
		logger.Error("wx_pay_error",
			zap.Time("time", time.Now()),         // 记录时间
			zap.Any("error", err),                // 记录错误信息
			zap.String("request", bm.JsonBody()), // 请求信息
		)
		return "", err
	}
	prepayId := wxRsp.Response.PrepayId
	orderInfo, err = ConWechat().Client.PaySignOfApp(appId, prepayId)
	if err != nil {
		return "", err
	}
	return orderInfo, nil
}

// CreateH5Order 创建H5订单
func (w *MobileWxPassport) CreateH5Order(c *gin.Context, orderModel *order.Order) (orderInfo interface{}, err error) {
	// 初始化参数Map
	bm := make(gopay.BodyMap)
	appId := config.GetString(strconv.Itoa(orderModel.AppId)+".mywxpay.app_id_"+orderModel.ChannelId, config.GetString(strconv.Itoa(orderModel.AppId)+".mywxpay.app_id"))
	bm.Set("appid", appId).
		Set("mchid", config.GetString("mywxpay.mch_id")).
		Set("description", orderModel.GoodsName).
		Set("out_trade_no", orderModel.BsTradeNo).
		Set("time_expire", orderModel.ExpirationTime.Format(time.RFC3339)).
		Set("notify_url", config.GetString("mywxpay.pay_notify_url")+strconv.Itoa(orderModel.AppId)).
		SetBodyMap("amount", func(b gopay.BodyMap) {
			b.Set("total", orderModel.TotalFee).
				Set("currency", "CNY")
		})

	bm.SetBodyMap("scene_info", func(b gopay.BodyMap) {
		b.Set("payer_client_ip", c.ClientIP()).
			SetBodyMap("h5_info", func(bs gopay.BodyMap) {
				bs.Set("type", "Wap")
			})
	})

	wxRsp, err := ConWechat().Client.V3TransactionH5(c, bm)
	if err != nil {
		logger.Error("wx_pay_error",
			zap.Time("time", time.Now()),         // 记录时间
			zap.Any("error", err),                // 记录错误信息
			zap.String("request", bm.JsonBody()), // 请求信息
		)
		return "", err
	}
	logger.Debug("CreateH5Order", zap.String("wxRsp.Response", conv.String(wxRsp)))
	orderInfo = wxRsp.Response.H5Url
	return orderInfo, nil
}

// CreateQrOrder 创建二维码支付订单
func (w *MobileWxPassport) CreateQrOrder(c *gin.Context, orderModel *order.Order) (orderInfo interface{}, err error) {
	// 初始化参数Map
	bm := make(gopay.BodyMap)
	appId := config.GetString(strconv.Itoa(orderModel.AppId)+".mywxpay.app_id_"+orderModel.ChannelId, config.GetString(strconv.Itoa(orderModel.AppId)+".mywxpay.app_id"))
	bm.Set("appid", appId).
		Set("description", orderModel.GoodsName).
		Set("out_trade_no", orderModel.BsTradeNo).
		Set("time_expire", orderModel.ExpirationTime.Format(time.RFC3339)).
		Set("notify_url", config.GetString("mywxpay.pay_notify_url")+strconv.Itoa(orderModel.AppId)).
		SetBodyMap("amount", func(b gopay.BodyMap) {
			b.Set("total", orderModel.TotalFee).
				Set("currency", "CNY")
		})
	native, err := ConWechat().Client.V3TransactionNative(c, bm)
	if err != nil {
		return "", err
	}
	if native.Code == 400 {
		logger.Error("pay error",
			zap.Time("time", time.Now()),               // 记录时间
			zap.Any("error", errors.New(native.Error)), // 记录错误信息
			zap.String("request", bm.JsonBody()),       // 请求信息
		)
		return "", errors.New(native.Error)
	}
	orderInfo = native.Response.CodeUrl
	return orderInfo, nil
}

func CasePayStatus(status string) int {
	switch status {
	case "SUCCESS": //交易支付成功
		return 1
	case "NOTPAY": //未支付
		return 2
	case "CLOSED": //已关闭
		return 9
	case "REFUND": //转入退款
		return 7
	default: //未知
		return 0
	}
}
