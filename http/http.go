package http

import (
	"fmt"
	"log"
	_ "net/http/pprof"

	"github.com/astaxie/beego"
	"github.com/open-falcon/alarm/g"
)

// 注册MainController里的处理函数
func configRoutes() {
	// 返回按时间排序的未恢复告警列表
	beego.Router("/", &MainController{}, "get:Index")
	// 版本
	beego.Router("/version", &MainController{}, "get:Version")
	// 连通性
	beego.Router("/health", &MainController{}, "get:Health")
	// workdir
	beego.Router("/workdir", &MainController{}, "get:Workdir")
	// 热加载配置
	beego.Router("/config/reload", &MainController{}, "get:ConfigReload")
	// 删除内存中对应id的EventDto
	beego.Router("/event/solve", &MainController{}, "post:Solve")
}

func Duration(now, before int64) string {
	d := now - before
	if d <= 60 {
		return "just now"
	}

	if d <= 120 {
		return "1 minute ago"
	}

	if d <= 3600 {
		return fmt.Sprintf("%d minutes ago", d/60)
	}

	if d <= 7200 {
		return "1 hour ago"
	}

	if d <= 3600*24 {
		return fmt.Sprintf("%d hours ago", d/3600)
	}

	if d <= 3600*24*2 {
		return "1 day ago"
	}

	return fmt.Sprintf("%d days ago", d/3600/24)
}

func init() {
	configRoutes()
	// beego.AddFuncMap模板自定义方法，使用方法:beego.AddFuncMap("模版中调用的方法名", 具体函数)
	beego.AddFuncMap("duration", Duration)
}

func Start() {
	if !g.Config().Http.Enabled {
		return
	}

	addr := g.Config().Http.Listen
	if addr == "" {
		return
	}

	beego.Run(addr)

	log.Println("http listening", addr)
}
