package mailgun

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/schmooser/go-mailer/message"
)

type Mailgun struct {
	user   string
	pass   string
	server string
}

func New(user, pass, server string) *Mailgun {
	return &Mailgun{
		user:   user,
		pass:   pass,
		server: server,
	}
}

// Send sends the message.
func (mg *Mailgun) Send(msg *message.Message, async bool) (interface{}, error) {

	apiURL := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", mg.server)

	data := url.Values{}
	data.Add("from", fmt.Sprintf("%s <%s>", msg.FromName, msg.FromEmail))
	for _, rcpt := range msg.Recipients {
		data.Add("to", fmt.Sprintf("%s <%s>", rcpt.Name, rcpt.Email))
	}
	data.Add("subject", msg.Subject)
	data.Add("text", msg.Text)
	data.Add("html", msg.HTML)

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(mg.user, mg.pass)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(content), nil
}
