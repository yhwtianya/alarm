package cron

// Mail信息结构
type MailDto struct {
	Priority int    `json:"priority"`
	Metric   string `json:"metric"`
	Subject  string `json:"subject"`
	Content  string `json:"content"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

// Sms信息结构
type SmsDto struct {
	Priority int    `json:"priority"`
	Metric   string `json:"metric"`
	Content  string `json:"content"`
	Phone    string `json:"phone"`
	Status   string `json:"status"`
}
