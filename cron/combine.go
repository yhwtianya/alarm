package cron

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/open-falcon/alarm/api"
	"github.com/open-falcon/alarm/g"
	redi "github.com/open-falcon/alarm/redis"
)

// 周期性读取Sms待告警合并队列，合并,并放入Sms队列
func CombineSms() {
	for {
		// 每分钟读取处理一次
		time.Sleep(time.Minute)
		combineSms()
	}
}

// 周期性读取Mail待告警合并队列，合并,并放入Mail队列
func CombineMail() {
	for {
		// 每分钟读取处理一次
		time.Sleep(time.Minute)
		combineMail()
	}
}

// Mail告警压缩,并放入Mail队列
func combineMail() {
	dtos := popAllMailDto()
	count := len(dtos)
	if count == 0 {
		return
	}

	dtoMap := make(map[string][]*MailDto)
	for i := 0; i < count; i++ {
		key := fmt.Sprintf("%d%s%s%s", dtos[i].Priority, dtos[i].Status, dtos[i].Email, dtos[i].Metric)
		if _, ok := dtoMap[key]; ok {
			dtoMap[key] = append(dtoMap[key], dtos[i])
		} else {
			dtoMap[key] = []*MailDto{dtos[i]}
		}
	}

	// 不要在这处理，继续写回redis，否则重启alarm很容易丢数据
	for _, arr := range dtoMap {
		size := len(arr)
		if size == 1 {
			redi.WriteMail([]string{arr[0].Email}, arr[0].Subject, arr[0].Content)
			continue
		}

		subject := fmt.Sprintf("[P%d][%s] %d %s", arr[0].Priority, arr[0].Status, size, arr[0].Metric)
		contentArr := make([]string, size)
		for i := 0; i < size; i++ {
			contentArr[i] = arr[i].Content
		}
		// 多个告警内容拼接到一个邮件中
		content := strings.Join(contentArr, "\r\n")

		redi.WriteMail([]string{arr[0].Email}, subject, content)
	}
}

// Sms告警压缩,并放入Sms队列
func combineSms() {
	dtos := popAllSmsDto()
	count := len(dtos)
	if count == 0 {
		return
	}

	dtoMap := make(map[string][]*SmsDto)
	for i := 0; i < count; i++ {
		//Priority+Status+Phone+Metric共同作为key
		key := fmt.Sprintf("%d%s%s%s", dtos[i].Priority, dtos[i].Status, dtos[i].Phone, dtos[i].Metric)
		if _, ok := dtoMap[key]; ok {
			dtoMap[key] = append(dtoMap[key], dtos[i])
		} else {
			dtoMap[key] = []*SmsDto{dtos[i]}
		}
	}

	for _, arr := range dtoMap {
		size := len(arr)
		if size == 1 {
			redi.WriteSms([]string{arr[0].Phone}, arr[0].Content)
			continue
		}

		// 把多个sms内容写入数据库，只给用户提供一个链接
		contentArr := make([]string, size)
		for i := 0; i < size; i++ {
			contentArr[i] = arr[i].Content
		}
		content := strings.Join(contentArr, ",,")

		first := arr[0].Content
		t := strings.Split(first, "][")
		// 故障描述信息
		eg := ""
		if len(t) >= 3 {
			eg = t[2]
		}

		// 多个告警内容生成一个url，供用户访问
		path, err := api.LinkToSMS(content)
		sms := ""
		if err != nil || path == "" {
			sms = fmt.Sprintf("[P%d][%s] %d %s.  e.g. %s detail in email", arr[0].Priority, arr[0].Status, size, arr[0].Metric, eg)
			log.Println("get link fail", err)
		} else {
			links := g.Config().Api.Links
			sms = fmt.Sprintf("[P%d][%s] %d %s e.g. %s %s/%s ", arr[0].Priority, arr[0].Status, size, arr[0].Metric, eg, links, path)
		}

		redi.WriteSms([]string{arr[0].Phone}, sms)
	}

}

// 读取Sms信息队列中所有数据
func popAllSmsDto() []*SmsDto {
	ret := []*SmsDto{}
	queue := g.Config().Redis.UserSmsQueue

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				log.Println("get SmsDto fail", err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var smsDto SmsDto
		err = json.Unmarshal([]byte(reply), &smsDto)
		if err != nil {
			log.Printf("json unmarshal SmsDto: %s fail: %v", reply, err)
			continue
		}

		ret = append(ret, &smsDto)
	}

	return ret
}

// 读取Mail信息队列中所有数据
func popAllMailDto() []*MailDto {
	ret := []*MailDto{}
	queue := g.Config().Redis.UserMailQueue

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	for {
		reply, err := redis.String(rc.Do("RPOP", queue))
		if err != nil {
			if err != redis.ErrNil {
				log.Println("get MailDto fail", err)
			}
			break
		}

		if reply == "" || reply == "nil" {
			continue
		}

		var mailDto MailDto
		err = json.Unmarshal([]byte(reply), &mailDto)
		if err != nil {
			log.Printf("json unmarshal MailDto: %s fail: %v", reply, err)
			continue
		}

		ret = append(ret, &mailDto)
	}

	return ret
}
