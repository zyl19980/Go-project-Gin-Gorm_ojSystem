package test

import (
	"crypto/tls"
	"github.com/jordan-wright/email"
	"net/smtp"
	"testing"
)

func TestSendEmail(t *testing.T) {
	e := email.NewEmail()
	e.From = "Get <zyl02work@163.com>"
	e.To = []string{"1013277321@qq.com"}
	e.Subject = "验证码发送测试"
	e.HTML = []byte("验证码：<h1>Fancy HTML is supported, too!</h1>")
	err := e.Send("smtp.163.com:456", smtp.PlainAuth("", "zyl02work@163.com>", "NBeN4UuWtBqDvGmc", "smtp.163.com"))
	//如果返回 EOF 错误
	// 返回 EOF 时，关闭SSL重试
	err = e.SendWithTLS("smtp.163.com:465",
		smtp.PlainAuth("", "zyl02work@163.com", "NBeN4UuWtBqDvGmc", "smtp.163.com"),
		&tls.Config{InsecureSkipVerify: true, ServerName: "smtp.163.com"})
	if err != nil {
		t.Fatal(err)
	}

}
