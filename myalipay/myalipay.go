package myalipay

import (
	"context"
	"errors"
	"github.com/Adam0120/sdk-utils/config"
	"github.com/Adam0120/sdk-utils/helpers"
	"github.com/Adam0120/sdk-utils/logger"
	"github.com/Adam0120/sdk-utils/models/order"
	"github.com/gin-gonic/gin"
	"github.com/go-pay/gopay"
	"github.com/go-pay/gopay/alipay"
	"strconv"
	"sync"
)

var AlLoginPass = new(AiPassport)

// AliPayClient 服务
type AliPayClient struct {
	Client  *alipay.Client
	Context context.Context
}

// once 确保全局的 Wechat 对象只实例一次
var once sync.Once

var aliyunCli map[string]*AliPayClient

// ConAliPay 连接 支付宝
func ConAliPay(gameId string) *AliPayClient {
	once.Do(func() {
		aliyunCli = make(map[string]*AliPayClient)
		appIDs := config.GetStringSlice("app_ids")
		if len(appIDs) == 0 {
			logger.ErrorString("alipay", "ConAliPay", "app_ids is empty")
			panic("app_ids is empty")
			return
		}
		for _, appID := range appIDs {
			aliyunCli[appID] = newClient(appID)
		}
	})
	if aliyunCli[gameId] == nil {
		return nil
	}
	return aliyunCli[gameId]
}

// newClient 创建一个新的 AliPay 连接
func newClient(gameId string) *AliPayClient {
	// 初始化自定的 AliyunClient 实例
	alc := &AliPayClient{}
	// 使用默认的 context
	alc.Context = context.Background()
	var err error
	alc.Client, err = alipay.NewClient(config.GetString(gameId+".aliyun.pay.appid"), config.GetToPath(gameId+".aliyun.pay.app_pri_key"), true)
	logger.LogIf(err)

	alc.Client.SetLocation(config.GetString("app.timezone", "Asia/Shanghai")). // 设置时区，不设置或出错均为默认服务器时间
											SetNotifyUrl(config.GetString(gameId + ".aliyun.pay.notify_url")) // 设置异步通知URL
	alc.Client.SetReturnUrl(config.GetString(gameId + ".aliyun.pay.return_url"))
	logger.LogIf(err)

	return alc
}

func (w *AiPassport) CreateOrder(c *gin.Context, orderModel *order.Order) (orderInfo string, err error) {
	gameId := strconv.Itoa(orderModel.AppId)
	// 初始化参数Map
	bm := make(gopay.BodyMap)
	bm.Set("subject", orderModel.GoodsName).
		Set("out_trade_no", orderModel.BsTradeNo).
		Set("time_expire", orderModel.ExpirationTime.Format("2006-01-02 15:04:05")).
		Set("total_amount", helpers.Divide(orderModel.TotalFee))
	orderInfo, err = ConAliPay(gameId).Client.TradeAppPay(c, bm)
	return
}

func (w *AiPassport) CreateH5Order(c *gin.Context, orderModel *order.Order, url string) (orderInfo string, err error) {
	gameId := strconv.Itoa(orderModel.AppId)
	// 初始化参数Map
	bm := make(gopay.BodyMap)
	bm.Set("subject", orderModel.GoodsName).
		Set("out_trade_no", orderModel.BsTradeNo).
		Set("time_expire", orderModel.ExpirationTime.Format("2006-01-02 15:04:05")).
		Set("total_amount", helpers.Divide(orderModel.TotalFee)).
		Set("quit_url", url)
	orderInfo, err = ConAliPay(gameId).Client.TradeWapPay(c, bm)
	return
}

func (w *AiPassport) CreateQrOrder(c *gin.Context, orderModel *order.Order) (orderInfo string, err error) {
	gameId := strconv.Itoa(orderModel.AppId)
	// 初始化参数Map
	bm := make(gopay.BodyMap)
	bm.Set("subject", orderModel.GoodsName).
		Set("out_trade_no", orderModel.BsTradeNo).
		Set("time_expire", orderModel.ExpirationTime.Format("2006-01-02 15:04:05")).
		Set("total_amount", helpers.Divide(orderModel.TotalFee))
	orderInfo2, err2 := ConAliPay(gameId).Client.TradePrecreate(c, bm)
	return orderInfo2.Response.QrCode, err2
}

func (w *AiPassport) SearchOrder(c *gin.Context, orderModel *order.Order) (map[string]interface{}, error) {
	gameId := strconv.Itoa(orderModel.AppId)
	orderMap := make(map[string]interface{})
	bm := make(gopay.BodyMap)
	bm.Set("out_trade_no", orderModel.BsTradeNo)

	// 查询订单
	aliRsp, err := ConAliPay(gameId).Client.TradeQuery(c, bm)
	if err != nil {
		return orderMap, err
	}

	// 同步返回验签
	ok, err := alipay.VerifySyncSignWithCert(config.Get(gameId+".aliyun.pay.pub_key"), aliRsp.SignData, aliRsp.Sign)
	if err != nil {
		return orderMap, err
	}
	if !ok {
		return orderMap, errors.New("验签失败")
	}
	orderMap = map[string]interface{}{
		"trade_state":      CasePayStatus(aliRsp.Response.TradeStatus),
		"trade_state_desc": aliRsp.Response.TradeSettleInfo,
		"transaction_id":   aliRsp.Response.TradeNo,
		"pay_time":         aliRsp.Response.SendPayDate,
		"trade_type":       aliRsp.Response.BuyerUserType,
		"pay_fee":          helpers.ReDivide(aliRsp.Response.TotalAmount),
	}
	return orderMap, err
}

func CasePayStatus(status string) int {
	switch status {
	case "TRADE_SUCCESS": //交易支付成功
		return 1
	case "WAIT_BUYER_PAY": //交易创建，等待买家付款
		return 2
	case "TRADE_CLOSED": //未付款交易超时关闭，或支付完成后全额退款未付款交易超时关闭，或支付完成后全额退款
		return 9
	case "TRADE_FINISHED": //交易结束，不可退款
		return 4
	default: //未知
		return 0
	}
}
