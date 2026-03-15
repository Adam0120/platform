package huawei

import (
	"github.com/Adam0120/sdk-utils/models/order"
	json "github.com/json-iterator/go"
	"strings"
)

func (sv *Verify) CreateOrder(orderModel order.Order) (map[string]any, error) {
	// 初始化参数Map
	params := make(map[string]interface{})
	//是否是订阅商品
	if strings.Contains(orderModel.GoodsName, "订阅") {
		params["isSub"] = 1
	} else {
		params["isSub"] = 0
	}
	return params, nil
}

type OrderNotific struct {
	AppId             string `json:"appId"`
	Version           string `json:"version"`
	NotifyTime        int64  `json:"notifyTime"`
	EventType         string `json:"eventType"`
	ApplicationId     string `json:"applicationId"`
	OrderNotification struct {
		Version          string `json:"version"`
		NotificationType int    `json:"notificationType"`
		PurchaseToken    string `json:"purchaseToken"`
		ProductId        string `json:"productId"`
	} `json:"orderNotification,omitempty"`
	SubNotification struct {
		Version                  string `json:"version"`
		StatusUpdateNotification string `json:"statusUpdateNotification"`
		NotificationSignature    string `json:"notificationSignature"`
		SignatureAlgorithm       string `json:"signatureAlgorithm"`
	} `json:"subNotification,omitempty"`
}

func OrderVerify(request OrderNotific) (*StatusUpdateNotification, error) {
	return VerifyToken(request.OrderNotification.PurchaseToken, request.OrderNotification.ProductId, request.AppId)
}
func VerifyToken(purchaseToken, productId, appId string) (*StatusUpdateNotification, error) {
	bodyMap := map[string]string{"purchaseToken": purchaseToken, "productId": productId, "appId": appId}
	bodyBytes, err := SendRequest("https://orders-drcn.iap.cloud.huawei.com.cn/applications/purchases/tokens/verify", bodyMap)
	if err != nil {
		return nil, err
	}
	notification, err := NotificationDemo.DealNotification(bodyBytes, appId)
	if err != nil {
		return nil, err
	}
	bodyMap["appId"] = appId
	err = ConfirmPurchase(bodyMap)
	if err != nil && err.Error() != "already consumed" {
		return nil, err
	}
	return notification, nil
}

func VerifySubscription(inappPurchaseData, dataSignature, appId string) (*SubscriptionNotification, error) {
	err := VerifyRsaSign(inappPurchaseData, dataSignature, appId)
	if err != nil {
		return nil, err
	}
	var info SubscriptionNotification
	err = json.Unmarshal([]byte(inappPurchaseData), &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func GetSubscription(subscriptionId, purchaseToken, appId string) (*StatusUpdateNotification, error) {
	bodyBytes, err := SubscriptionDemo.GetSubscription(subscriptionId, purchaseToken, appId, 1)
	if err != nil {
		return nil, err
	}
	var info SubscriptionRequest
	err = json.Unmarshal(bodyBytes, &info)
	if err != nil {
		return nil, err
	}

	var request StatusUpdateNotification
	err = json.Unmarshal([]byte(info.InappPurchaseData), &request)
	if err != nil {
		return nil, err
	}
	//TODO 不确定是否需要确认订单
	//bodyMap := map[string]string{"purchaseToken": request.PurchaseToken, "productId": request.ProductId}
	//err = ConfirmPurchase(bodyMap)
	//if err != nil && err.Error() != "already consumed" {
	//	return nil, err
	//}
	return &request, nil
}
