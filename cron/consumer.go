package cron

import (
	"encoding/json"
	"log"

	"github.com/open-falcon/alarm/api"
	"github.com/open-falcon/alarm/g"
	"github.com/open-falcon/alarm/redis"
	"github.com/open-falcon/common/model"
)

func consume(event *model.Event, isHigh bool) {
	actionId := event.ActionId()
	if actionId <= 0 {
		return
	}

	// 获取action
	action := api.GetAction(actionId)
	if action == nil {
		return
	}

	if action.Callback == 1 {
		// 执行回调流程，立即返回，不再走默认的Sms和Mail流程
		// 回调配置中可以设置回调前后是否发送Sms或Mail
		HandleCallback(event, action)
		return
	}

	if isHigh {
		consumeHighEvents(event, action)
	} else {
		consumeLowEvents(event, action)
	}
}

// 消费高优先级event，构造告警结构，放入对应告警队列
// 高优先级的不做报警合并
func consumeHighEvents(event *model.Event, action *api.Action) {
	if action.Uic == "" {
		return
	}

	phones, mails := api.ParseTeams(action.Uic)

	smsContent := GenerateSmsContent(event)
	mailContent := GenerateMailContent(event)

	// 这里不做告警合并，直接写入发送队列
	if event.Priority() < 3 {
		// Priority小于3，触发发送SMS
		redis.WriteSms(phones, smsContent)
	}

	redis.WriteMail(mails, smsContent, mailContent)
}

// 消费低优先级event，构造告警结构，放入对应告警合并队列
// 低优先级的做报警合并
func consumeLowEvents(event *model.Event, action *api.Action) {
	if action.Uic == "" {
		return
	}

	if event.Priority() < 3 {
		// Priority小于3，触发发送SMS
		ParseUserSms(event, action)
	}

	ParseUserMail(event, action)
}

// 构造sms结构，放入低优先级sms待告警合并队列
func ParseUserSms(event *model.Event, action *api.Action) {
	userMap := api.GetUsers(action.Uic)

	content := GenerateSmsContent(event)
	metric := event.Metric()
	status := event.Status
	priority := event.Priority()

	// 低优先级sms待告警合并队列
	queue := g.Config().Redis.UserSmsQueue

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for _, user := range userMap {
		dto := SmsDto{
			Priority: priority,
			Metric:   metric,
			Content:  content,
			Phone:    user.Phone,
			Status:   status,
		}
		bs, err := json.Marshal(dto)
		if err != nil {
			log.Println("json marshal SmsDto fail:", err)
			continue
		}

		_, err = rc.Do("LPUSH", queue, string(bs))
		if err != nil {
			log.Println("LPUSH redis", queue, "fail:", err, "dto:", string(bs))
		}
	}
}

// 构造mail结构，放入低优先级mail待告警合并队列
func ParseUserMail(event *model.Event, action *api.Action) {
	userMap := api.GetUsers(action.Uic)

	metric := event.Metric()
	subject := GenerateSmsContent(event)
	content := GenerateMailContent(event)
	status := event.Status
	priority := event.Priority()

	// 低优先级mail待告警合并队列
	queue := g.Config().Redis.UserMailQueue

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for _, user := range userMap {
		dto := MailDto{
			Priority: priority,
			Metric:   metric,
			Subject:  subject,
			Content:  content,
			Email:    user.Email,
			Status:   status,
		}
		bs, err := json.Marshal(dto)
		if err != nil {
			log.Println("json marshal MailDto fail:", err)
			continue
		}

		_, err = rc.Do("LPUSH", queue, string(bs))
		if err != nil {
			log.Println("LPUSH redis", queue, "fail:", err, "dto:", string(bs))
		}
	}
}
