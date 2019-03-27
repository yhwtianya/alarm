package api

import (
	"fmt"
	"time"

	"github.com/open-falcon/alarm/g"
	"github.com/toolkits/net/httplib"
)

// 根据Sms的Content请求link模块生成url连接
func LinkToSMS(content string) (string, error) {
	links := g.Config().Api.Links
	uri := fmt.Sprintf("%s/store", links)
	req := httplib.Post(uri).SetTimeout(3*time.Second, 10*time.Second)
	req.Body([]byte(content))
	return req.String()
}
