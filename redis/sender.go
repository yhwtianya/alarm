package redis

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/open-falcon/alarm/g"
	"github.com/open-falcon/sender/model"
)

// 发送msg到redis列表
func LPUSH(queue, message string) {
	rc := g.RedisConnPool.Get()
	defer rc.Close()
	_, err := rc.Do("LPUSH", queue, message)
	if err != nil {
		log.Println("LPUSH redis", queue, "fail:", err, "message:", message)
	}
}

// 检查sms合法性，减少空格，发送sms到redis队列
func WriteSmsModel(sms *model.Sms) {
	if sms == nil {
		return
	}

	//检查sms合法性，减少空格
	bs, err := json.Marshal(sms)
	if err != nil {
		log.Println(err)
		return
	}

	LPUSH(g.Config().Queue.Sms, string(bs))
}

// 检查mail合法性，减少空格，发送mail到redis队列
func WriteMailModel(mail *model.Mail) {
	if mail == nil {
		return
	}

	bs, err := json.Marshal(mail)
	if err != nil {
		log.Println(err)
		return
	}

	LPUSH(g.Config().Queue.Mail, string(bs))
}

// 将Sms放入redis队列
func WriteSms(tos []string, content string) {
	if len(tos) == 0 {
		return
	}

	sms := &model.Sms{Tos: strings.Join(tos, ","), Content: content}
	WriteSmsModel(sms)
}

// 将mail放入redis队列
func WriteMail(tos []string, subject, content string) {
	if len(tos) == 0 {
		return
	}

	mail := &model.Mail{Tos: strings.Join(tos, ","), Subject: subject, Content: content}
	WriteMailModel(mail)
}
