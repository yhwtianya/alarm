package cron

import (
	"fmt"

	"github.com/open-falcon/alarm/g"
	"github.com/open-falcon/common/model"
	"github.com/open-falcon/common/utils"
)

// 构造通用sms告警内容
func BuildCommonSMSContent(event *model.Event) string {
	return fmt.Sprintf(
		"[P%d][%s][%s][][%s %s %s %s %s%s%s][O%d %s]",
		event.Priority(),
		event.Status,
		event.Endpoint,
		event.Note(),
		event.Func(),
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.CurrentStep,
		event.FormattedTime(),
	)
}

// 构造通用mail告警内容
func BuildCommonMailContent(event *model.Event) string {
	// 生成查看对应template或expression的url
	link := g.Link(event)
	return fmt.Sprintf(
		"%s\r\nP%d\r\nEndpoint:%s\r\nMetric:%s\r\nTags:%s\r\n%s: %s%s%s\r\nNote:%s\r\nMax:%d, Current:%d\r\nTimestamp:%s\r\n%s\r\n",
		event.Status,
		event.Priority(),
		event.Endpoint,
		event.Metric(),
		utils.SortedTags(event.PushedTags),
		event.Func(),
		utils.ReadableFloat(event.LeftValue),
		event.Operator(),
		utils.ReadableFloat(event.RightValue()),
		event.Note(),
		event.MaxStep(),
		event.CurrentStep,
		event.FormattedTime(),
		link,
	)
}

// 生成sms告警内容
func GenerateSmsContent(event *model.Event) string {
	return BuildCommonSMSContent(event)
}

// 生成mail告警内容
func GenerateMailContent(event *model.Event) string {
	return BuildCommonMailContent(event)
}
