/*
 * Copyright 2020. Huawei Technologies Co., Ltd. All rights reserved.
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 *
 */
package huawei

import (
	"encoding/json"
)

const (
	INITIAL_BUY            = 0
	CANCEL                 = 1
	RENEWAL                = 2
	INTERACTIVE_RENEWAL    = 3
	NEW_RENEWAL_PREF       = 4
	RENEWAL_STOPPED        = 5
	RENEWAL_RESTORED       = 6
	RENEWAL_RECURRING      = 7
	ON_HOLD                = 9
	PAUSED                 = 10
	PAUSE_PLAN_CHANGED     = 11
	PRICE_CHANGE_CONFIRMED = 12
	DEFERRED               = 13
)

type NotificationServer struct {
}

var NotificationDemo = &NotificationServer{}

type NotificationRequest struct {
	StatusUpdateNotification string `json:"purchaseTokenData"`
	NotificationSignature    string `json:"dataSignature"`
}

type StatusUpdateNotification struct {
	AutoRenewing        bool   `json:"autoRenewing"`
	OrderId             string `json:"orderId"`
	PackageName         string `json:"packageName"`
	ApplicationId       int    `json:"applicationId"`
	ApplicationIdString string `json:"applicationIdString"`
	Kind                int    `json:"kind"`
	ProductId           string `json:"productId"`
	ProductName         string `json:"productName"`
	PurchaseTime        int64  `json:"purchaseTime"`
	PurchaseTimeMillis  int64  `json:"purchaseTimeMillis"`
	PurcHaseState       int    `json:"purchaseState"`
	PurchaseToken       string `json:"purchaseToken"`
	ResponseCode        string `json:"responseCode"`
	ConsumptionState    int    `json:"consumptionState"`
	Confirmed           int    `json:"confirmed"`
	PurchaseType        int    `json:"purchaseType"`
	Currency            string `json:"currency"`
	Price               int    `json:"price"`
	Country             string `json:"country"`
	PayOrderId          string `json:"payOrderId"`
	PayType             string `json:"payType"`
	SdkChannel          string `json:"sdkChannel"`
	DeveloperPayload    string `json:"developerPayload"`
	//-----订阅------------------------------
	SubIsvalid           bool   `json:"subIsvalid"`
	LastOrderId          string `json:"lastOrderId"`
	ProductGroup         string `json:"productGroup"`
	OriPurchaseTime      int64  `json:"oriPurchaseTime"`
	SubscriptionId       string `json:"subscriptionId"`
	Quantity             int    `json:"quantity"`
	DaysLasted           int    `json:"daysLasted"`
	NumOfPeriods         int    `json:"numOfPeriods"`
	NumOfDiscount        int    `json:"numOfDiscount"`
	ExpirationDate       int64  `json:"expirationDate"`
	RetryFlag            int    `json:"retryFlag"`
	IntroductoryFlag     int    `json:"introductoryFlag"`
	TrialFlag            int    `json:"trialFlag"`
	RenewStatus          int    `json:"renewStatus"`
	CancelledSubKeepDays int    `json:"cancelledSubKeepDays"`
}

type SubscriptionNotification struct {
	Environment                string `json:"environment"`
	NotificationType           int    `json:"notificationType"`
	SubscriptionId             string `json:"subscriptionId"`
	PurchaseToken              string `json:"purchaseToken"`
	OrderId                    string `json:"orderId"`
	LatestReceipt              string `json:"latestReceipt"`
	LatestReceiptInfo          string `json:"latestReceiptInfo"`
	LatestReceiptInfoSignature string `json:"latestReceiptInfoSignature"`
	SignatureAlgorithm         string `json:"signatureAlgorithm"`
	AutoRenewStatus            int    `json:"autoRenewStatus"`
	ProductId                  string `json:"productId"`
	ApplicationId              string `json:"applicationId"`
}

func (eventServer *NotificationServer) DealNotification(information []byte, appId string) (*StatusUpdateNotification, error) {
	var request NotificationRequest
	err := json.Unmarshal(information, &request)
	if err != nil {
		return nil, err
	}
	err = VerifyRsaSign(request.StatusUpdateNotification, request.NotificationSignature, appId)
	if err != nil {
		return nil, err
	}
	var info StatusUpdateNotification
	json.Unmarshal([]byte(request.StatusUpdateNotification), &info)
	return &info, nil
}

type SubscriptionRequest struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	InappPurchaseData  string `json:"inappPurchaseData"`
	DataSignature      string `json:"dataSignature"`
	SignatureAlgorithm string `json:"signatureAlgorithm"`
}
