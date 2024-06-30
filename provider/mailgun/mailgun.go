package mailgun

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"

	"github.com/velppa/go-mailer/message"
)

// Client defines Client transactional mail provider.
type Client struct {
	User   string
	Pass   string
	Server string
}

// Send sends the message.
func (mg *Client) Send(msg *message.Message, async bool) (any, error) {

	apiURL := fmt.Sprintf("https://api.mailgun.net/v3/%s/messages", mg.Server)

	data := url.Values{}
	data.Add("from", msg.From.String())

	data.Add("to", msg.To.String())
	data.Add("cc", msg.CC.String())
	data.Add("bcc", msg.BCC.String())

	data.Add("subject", msg.Subject)
	data.Add("text", msg.Text)
	data.Add("html", msg.HTML)

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return "", err
	}

	req.SetBasicAuth(mg.User, mg.Pass)
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
