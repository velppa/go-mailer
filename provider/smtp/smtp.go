package smtp

import (
	"net/mail"

	"github.com/velppa/go-mailer/message"
	"gopkg.in/gomail.v2"
)

type Client struct {
	Host     string
	Port     int
	UserName string
	FromName string
	Password string
}

func (c *Client) Send(msg *message.Message, _ bool) (any, error) {
	m := gomail.NewMessage()
	from := &mail.Address{Address: c.UserName, Name: c.FromName}
	m.SetHeader("From", from.String())
	m.SetHeader("To", msg.To.Slice()...)
	m.SetHeader("Cc", msg.CC.Slice()...)
	m.SetHeader("Bcc", msg.BCC.Slice()...)
	m.SetHeader("Subject", msg.Subject)
	m.SetBody("text/html", msg.HTML)
	m.SetBody("text", msg.Text)

	d := gomail.NewDialer(c.Host, c.Port, c.UserName, c.Password)

	err := d.DialAndSend(m)
	return nil, err
}
