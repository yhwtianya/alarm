package api

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/open-falcon/alarm/g"
	"github.com/toolkits/container/set"
	"github.com/toolkits/net/httplib"
)

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Phone string `json:"phone"`
}

// uic响应获取User的结构
type UsersWrap struct {
	Msg   string  `json:"msg"`
	Users []*User `json:"users"`
}

type UsersCache struct {
	sync.RWMutex
	M map[string][]*User
}

// 缓存User信息
var Users = &UsersCache{M: make(map[string][]*User)}

func (this *UsersCache) Get(team string) []*User {
	this.RLock()
	defer this.RUnlock()
	val, exists := this.M[team]
	if !exists {
		return nil
	}

	return val
}

// 更新User
func (this *UsersCache) Set(team string, users []*User) {
	this.Lock()
	defer this.Unlock()
	this.M[team] = users
}

// 获取team下的User信息,优先通过uic接口获取，失败则直接从缓存获取
func UsersOf(team string) []*User {
	users := CurlUic(team)

	if users != nil {
		Users.Set(team, users)
	} else {
		users = Users.Get(team)
	}

	return users
}

// 获取teams列表关联的所有User信息
func GetUsers(teams string) map[string]*User {
	userMap := make(map[string]*User)
	arr := strings.Split(teams, ",")
	for _, team := range arr {
		if team == "" {
			continue
		}

		users := UsersOf(team)
		if users == nil {
			continue
		}

		for _, user := range users {
			userMap[user.Name] = user
		}
	}
	return userMap
}

// 返回teams关联的所有phone和mail地址
// return phones, emails
func ParseTeams(teams string) ([]string, []string) {
	if teams == "" {
		return []string{}, []string{}
	}

	userMap := GetUsers(teams)
	// 防止重复，使用Set保存
	phoneSet := set.NewStringSet()
	mailSet := set.NewStringSet()
	for _, user := range userMap {
		phoneSet.Add(user.Phone)
		mailSet.Add(user.Email)
	}
	return phoneSet.ToSlice(), mailSet.ToSlice()
}

// 通过uic获取team下的User信息
func CurlUic(team string) []*User {
	if team == "" {
		return []*User{}
	}

	uri := fmt.Sprintf("%s/team/users", g.Config().Api.Uic)
	req := httplib.Get(uri).SetTimeout(2*time.Second, 10*time.Second)
	req.Param("name", team)
	req.Param("token", g.Config().UicToken)

	var usersWrap UsersWrap
	err := req.ToJson(&usersWrap)
	if err != nil {
		log.Printf("curl %s fail: %v", uri, err)
		return nil
	}

	if usersWrap.Msg != "" {
		log.Printf("curl %s return msg: %v", uri, usersWrap.Msg)
		return nil
	}

	return usersWrap.Users
}
