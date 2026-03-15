package huawei

import (
	"fmt"
	"log"
)

type SubscriptionClient struct {
}

var SubscriptionDemo = &SubscriptionClient{}

func getSubUrl(accountFlag int) string {
	if accountFlag == 1 {
		// site for telecom carrier
		return "https://subscr-drcn.iap.cloud.huawei.com.cn"
	} else {
		// TODO: replace the (ip:port) to the real one
		return "http://exampleserver/_mockserver_"
	}
}

func (subscriptionDemo *SubscriptionClient) GetSubscription(subscriptionId, purchaseToken, appId string, accountFlag int) ([]byte, error) {
	bodyMap := map[string]string{
		"subscriptionId": subscriptionId,
		"purchaseToken":  purchaseToken,
		"appId":          appId,
	}
	url := getSubUrl(accountFlag) + "/sub/applications/v2/purchases/get"
	bodyBytes, err := SendRequest(url, bodyMap)
	if err != nil {
		log.Printf("err is %s", err)
		return nil, err
	}
	log.Printf("%s", bodyBytes)
	return bodyBytes, nil
}

func (subscriptionDemo *SubscriptionClient) StopSubscription(subscriptionId, purchaseToken string, accountFlag int) {
	bodyMap := map[string]string{
		"subscriptionId": subscriptionId,
		"purchaseToken":  purchaseToken,
	}
	url := getSubUrl(accountFlag) + "/sub/applications/v2/purchases/stop"
	bodyBytes, err := SendRequest(url, bodyMap)
	if err != nil {
		log.Printf("err is %s", err)
	}
	// TODO: display the response as string in console, you can replace it with your business logic.
	log.Printf("%s", bodyBytes)
}

func (subscriptionDemo *SubscriptionClient) DelaySubscription(subscriptionId, purchaseToken string, currentExpirationTime, desiredExpirationTime int64, accountFlag int) {
	bodyMap := map[string]string{
		"subscriptionId":        subscriptionId,
		"purchaseToken":         purchaseToken,
		"currentExpirationTime": fmt.Sprintf("%v", currentExpirationTime),
		"desiredExpirationTime": fmt.Sprintf("%v", desiredExpirationTime),
	}
	url := getSubUrl(accountFlag) + "/sub/applications/v2/purchases/delay"
	bodyBytes, err := SendRequest(url, bodyMap)
	if err != nil {
		log.Printf("err is %s", err)
	}
	// TODO: display the response as string in console, you can replace it with your business logic.
	log.Printf("%s", bodyBytes)
}

func (subscriptionDemo *SubscriptionClient) ReturnFeeSubscription(subscriptionId, purchaseToken string, accountFlag int) {
	bodyMap := map[string]string{
		"subscriptionId": subscriptionId,
		"purchaseToken":  purchaseToken,
	}

	url := getSubUrl(accountFlag) + "/sub/applications/v2/purchases/returnFee"
	bodyBytes, err := SendRequest(url, bodyMap)
	if err != nil {
		log.Printf("err is %s", err)
	}
	// TODO: display the response as string in console, you can replace it with your business logic.
	log.Printf("%s", bodyBytes)
}

func (subscriptionDemo *SubscriptionClient) WithdrawalSubscription(subscriptionId, purchaseToken string, accountFlag int) {
	bodyMap := map[string]string{
		"subscriptionId": subscriptionId,
		"purchaseToken":  purchaseToken,
	}
	url := getSubUrl(accountFlag) + "/sub/applications/v2/purchases/withdrawal"
	bodyBytes, err := SendRequest(url, bodyMap)
	if err != nil {
		log.Printf("err is %s", err)
	}
	// TODO: display the response as string in console, you can replace it with your business logic.
	log.Printf("%s", bodyBytes)
}
