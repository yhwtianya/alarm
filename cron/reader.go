package cron

import (
	"encoding/json"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/open-falcon/alarm/g"
	"github.com/open-falcon/common/model"
)

// 持续读取HighQueues,进行保存
func ReadHighEvent() {
	queues := g.Config().Redis.HighQueues
	if len(queues) == 0 {
		return
	}

	for {
		// 无限循环
		event, err := popEvent(queues)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		consume(event, true)
	}
}

// 持续读取LowQueues,进行保存
func ReadLowEvent() {
	queues := g.Config().Redis.LowQueues
	if len(queues) == 0 {
		return
	}

	for {
		event, err := popEvent(queues)
		if err != nil {
			time.Sleep(time.Second)
			continue
		}
		consume(event, false)
	}
}

// 取出并返回一个event。如果event状态为OK，从内存Events中删除，否则保存或更新到内存Events
func popEvent(queues []string) (*model.Event, error) {

	count := len(queues)

	params := make([]interface{}, count+1)
	for i := 0; i < count; i++ {
		params[i] = queues[i]
	}
	// set timeout 0
	params[count] = 0

	rc := g.RedisConnPool.Get()
	defer rc.Close()

	// brpop(key1, key2,… key N, timeout)返回其中任一key中的尾元素,如果所有key都没值，则超时
	// redis.Strings将redis查询结果转为[]string，返回值为[key,value]
	reply, err := redis.Strings(rc.Do("BRPOP", params...))
	if err != nil {
		log.Printf("get alarm event from redis fail: %v", err)
		return nil, err
	}

	var event model.Event
	err = json.Unmarshal([]byte(reply[1]), &event)
	if err != nil {
		log.Printf("parse alarm event fail: %v", err)
		return nil, err
	}

	if g.Config().Debug {
		log.Println("======>>>>")
		log.Println(event.String())
	}

	// save in memory. display in dashboard
	g.Events.Put(&event)

	return &event, nil
}
