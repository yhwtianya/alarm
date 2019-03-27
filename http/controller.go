package http

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/open-falcon/alarm/g"
	"github.com/toolkits/file"
)

// 通过内嵌beego.Controller，实现了beego.Controller所有方法
// MainController可以重写beego.Controller已有方法修改beego默认行为
// MainController可以新写处理函数，通过beego.Router("/", &MainController{}, "get:Index")这种形式注册即可
type MainController struct {
	beego.Controller
}

func (this *MainController) Version() {
	this.Ctx.WriteString(g.VERSION)
}

func (this *MainController) Health() {
	this.Ctx.WriteString("ok")
}

func (this *MainController) Workdir() {
	this.Ctx.WriteString(fmt.Sprintf("%s", file.SelfDir()))
}

func (this *MainController) ConfigReload() {
	remoteAddr := this.Ctx.Request.RemoteAddr
	if strings.HasPrefix(remoteAddr, "127.0.0.1") {
		g.ParseConfig(g.ConfigFile)
		this.Data["json"] = g.Config()
		this.ServeJSON()
	} else {
		this.Ctx.WriteString("no privilege")
	}
}

// 返回按时间排序的未恢复告警列表
func (this *MainController) Index() {
	events := g.Events.Clone()

	defer func() {
		this.Data["Now"] = time.Now().Unix()
		this.TplName = "index.html"
	}()

	count := len(events)
	if count == 0 {
		this.Data["Events"] = []*g.EventDto{}
		return
	}

	// 按照持续时间排序
	beforeOrder := make([]*g.EventDto, count)
	i := 0
	for _, event := range events {
		beforeOrder[i] = event
		i++
	}

	sort.Sort(g.OrderedEvents(beforeOrder))
	this.Data["Events"] = beforeOrder
}

// 删除内存中对应id的EventDto
func (this *MainController) Solve() {
	ids := this.GetString("ids")
	if ids == "" {
		this.Ctx.WriteString("")
		return
	}

	idArr := strings.Split(ids, ",,")
	for i := 0; i < len(idArr); i++ {
		g.Events.Delete(idArr[i])
	}

	this.Ctx.WriteString("")
}
