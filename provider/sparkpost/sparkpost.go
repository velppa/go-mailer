package sparkpost

import (
	"encoding/json"
	"errors"
	"net/http"

	log "gopkg.in/inconshreveable/log15.v2"
	"gopkg.in/jmcvetta/napping.v3"

	"github.com/schmooser/go-mailer/message"
)

// Log represents logger.
var Log = log.New()

// SparkPost defines SparkPost transactional mail provider.
type SparkPost struct {
	key string
}

const apiURL = "https://api.sparkpost.com/api/v1/transmissions?num_rcpt_errors=30"

// Address is SparkPost' address part.
type Address struct {
	Name     string `json:"name,omitempty"`
	Email    string `json:"email"`
	HeaderTo string `json:"header_to,omitempty"`
}

// Recipient defines SparkPost' recipient.
type Recipient struct {
	Address Address `json:"address"`
}

// Content defines SparkPost' content.
type Content struct {
	From    Address           `json:"from,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
	Subject string            `json:"subject,omitempty"`
	Text    string            `json:"text,omitempty"`
	HTML    string            `json:"html,omitempty"`
}

// Message defines SparkPost' message.
type Message struct {
	ReturnPath string      `json:"return_path,omitempty"`
	Recipients []Recipient `json:"recipients,omitempty"`
	Content    Content     `json:"content,omitempt"`
}

// New returns new SparkPost instance. API key is validated upon creation,
// returning error if key is not valid.
func New(key string) (*SparkPost, error) {
	return &SparkPost{key: key}, nil
}

// Send sends provided message in async or sync way.
func (sp *SparkPost) Send(msg *message.Message, async bool) (interface{}, error) {

	var recipients []Recipient
	for _, r := range msg.To {
		recipients = append(recipients, Recipient{
			Address: Address{
				Email: r.Address,
				Name:  r.Name,
			},
		})
	}
	for _, r := range append(msg.CC, msg.BCC...) {
		recipients = append(recipients, Recipient{
			Address: Address{
				Email:    r.Address,
				HeaderTo: msg.To.String(),
			},
		})
	}

	m := Message{
		ReturnPath: msg.From.Address,
		Recipients: recipients,
		Content: Content{
			From: Address{
				Name:  msg.From.Name,
				Email: msg.From.Address,
			},
			Subject: msg.Subject,
			Text:    msg.Text,
			HTML:    msg.HTML,
			Headers: map[string]string{"CC": msg.CC.String()},
		},
	}

	b, _ := json.Marshal(m)
	Log.Debug("Message to send", "msg", string(b))

	headers := make(http.Header)
	headers.Add("Authorization", sp.key)
	s := napping.Session{
		Header: &headers,
	}

	var result interface{}
	resp, err := s.Post(apiURL, &m, &result, nil)
	if err != nil {
		return nil, err
	}

	if resp.Status() == 200 {
		return result, nil
	}

	Log.Error("Response", "status", resp.Status(), "body", resp.RawText())
	return nil, errors.New("Non-200 status returned")
}

func init() {
	Log.SetHandler(log.DiscardHandler())
}
