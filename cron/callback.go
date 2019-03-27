package cron

import (
	"fmt"
	"strings"
	"time"

	"github.com/open-falcon/alarm/api"
	"github.com/open-falcon/alarm/redis"
	"github.com/open-falcon/common/model"
	"github.com/toolkits/net/httplib"
)

// 告警回调流程
func HandleCallback(event *model.Event, action *api.Action) {

	// falcon,dinp
	teams := action.Uic
	phones := []string{}
	mails := []string{}

	if teams != "" {
		phones, mails = api.ParseTeams(teams)
		smsContent := GenerateSmsContent(event)
		mailContent := GenerateMailContent(event)

		// 回调前发送Sms或Mail
		if action.BeforeCallbackSms == 1 {
			redis.WriteSms(phones, smsContent)
		}

		if action.BeforeCallbackMail == 1 {
			redis.WriteMail(mails, smsContent, mailContent)
		}
	}

	// 执行url回调
	message := Callback(event, action)

	if teams != "" {
		// 回调后发送Sms或Mail
		// 消息内容为回调函数的返回值
		if action.AfterCallbackSms == 1 {
			redis.WriteSms(phones, message)
		}

		if action.AfterCallbackMail == 1 {
			redis.WriteMail(mails, message, message)
		}
	}

}

// 执行url回调,使用get方法，返回get结果
func Callback(event *model.Event, action *api.Action) string {
	if action.Url == "" {
		return "callback url is blank"
	}

	L := make([]string, 0)
	if len(event.PushedTags) > 0 {
		for k, v := range event.PushedTags {
			L = append(L, fmt.Sprintf("%s:%s", k, v))
		}
	}

	tags := ""
	if len(L) > 0 {
		tags = strings.Join(L, ",")
	}

	req := httplib.Get(action.Url).SetTimeout(3*time.Second, 20*time.Second)

	req.Param("endpoint", event.Endpoint)
	req.Param("metric", event.Metric())
	req.Param("status", event.Status)
	req.Param("step", fmt.Sprintf("%d", event.CurrentStep))
	req.Param("priority", fmt.Sprintf("%d", event.Priority()))
	req.Param("time", event.FormattedTime())
	req.Param("tpl_id", fmt.Sprintf("%d", event.TplId()))
	req.Param("exp_id", fmt.Sprintf("%d", event.ExpressionId()))
	req.Param("stra_id", fmt.Sprintf("%d", event.StrategyId()))
	req.Param("tags", tags)

	resp, e := req.String()

	success := "success"
	if e != nil {
		success = fmt.Sprintf("fail:%s", e.Error())
	}
	message := fmt.Sprintf("curl %s %s. resp: %s", action.Url, success, resp)

	return message
}
