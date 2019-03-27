package g

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/toolkits/file"
)

type HttpConfig struct {
	Enabled bool   `json:"enabled"`
	Listen  string `json:"listen"`
}

type QueueConfig struct {
	Sms  string `json:"sms"`
	Mail string `json:"mail"`
}

type RedisConfig struct {
	Addr          string   `json:"addr"`
	MaxIdle       int      `json:"maxIdle"`
	HighQueues    []string `json:"highQueues"`    //高优先级event队列，e.g. p0-p5
	LowQueues     []string `json:"lowQueues"`     //低优先级event队列，e.g. p6
	UserSmsQueue  string   `json:"userSmsQueue"`  //低优先级sms待告警合并队列
	UserMailQueue string   `json:"userMailQueue"` //低优先级mail待告警合并队列
}

type ApiConfig struct {
	Portal string `json:"portal"`
	Uic    string `json:"uic"`
	Links  string `json:"links"`
}

type GlobalConfig struct {
	Debug    bool         `json:"debug"`
	UicToken string       `json:"uicToken"`
	Http     *HttpConfig  `json:"http"`
	Queue    *QueueConfig `json:"queue"`
	Redis    *RedisConfig `json:"redis"`
	Api      *ApiConfig   `json:"api"`
}

var (
	ConfigFile string
	config     *GlobalConfig
	configLock = new(sync.RWMutex)
)

func Config() *GlobalConfig {
	configLock.RLock()
	defer configLock.RUnlock()
	return config
}

func ParseConfig(cfg string) {
	if cfg == "" {
		log.Fatalln("use -c to specify configuration file")
	}

	if !file.IsExist(cfg) {
		log.Fatalln("config file:", cfg, "is not existent")
	}

	ConfigFile = cfg

	configContent, err := file.ToTrimString(cfg)
	if err != nil {
		log.Fatalln("read config file:", cfg, "fail:", err)
	}

	var c GlobalConfig
	err = json.Unmarshal([]byte(configContent), &c)
	if err != nil {
		log.Fatalln("parse config file:", cfg, "fail:", err)
	}

	configLock.Lock()
	defer configLock.Unlock()
	config = &c
	log.Println("read config file:", cfg, "successfully")
}
