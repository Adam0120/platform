package wt_platform

import (
	"bytes"
	"fmt"
	"github.com/Adam0120/sdk-utils/helpers"
	"net/url"
	"sort"
)

type SessionVerifyRequest struct {
	Session   string `json:"session"`
	AppId     string `json:"app_id,omitempty"`
	ChannelId int    `json:"channel_id"`
	Uid       string `json:"uid"`
	LoginType string `json:"login_type,omitempty"`
}

func Sort(params url.Values) bytes.Buffer {
	var keys []string
	for k := range params {
		if k == "sign" {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys) // 对键进行排序
	buf := bytes.Buffer{}
	for _, k := range keys {
		buf.WriteString(fmt.Sprintf("%s=%v&", k, params.Get(k)))
	}
	buf.Truncate(buf.Len() - 1)
	return buf
}

func Sign(params url.Values, secretKey string) string {
	s := Sort(params)
	s.WriteString(secretKey)
	return helpers.Md5String(s.String())
}

func MapSign(params map[string]any, secretKey string) string {
	paramsUrl := url.Values{}
	for k, v := range params {
		paramsUrl.Set(k, v.(string))
	}
	s := Sort(paramsUrl)
	s.WriteString(secretKey)
	return helpers.Md5String(s.String())
}
