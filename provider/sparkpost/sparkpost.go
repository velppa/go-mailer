package sparkpost

import (
	//"bytes"
	"errors"
	"net/http"
	"net/mail"

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

func fromRecipient(addr mail.Address, to message.AddressList) Recipient {
	return Recipient{
		Address: Address{
			Name:     addr.Name,
			Email:    addr.Address,
			HeaderTo: to.String(),
		},
	}
}

// New returns new SparkPost instance. API key is validated upon creation,
// returning error if key is not valid.
func New(key string) (*SparkPost, error) {
	return &SparkPost{key: key}, nil
}

// Send sends provided message in async or sync way.
func (sp *SparkPost) Send(msg *message.Message, async bool) (interface{}, error) {

	var emptyAddrList message.AddressList
	var recipients []Recipient
	for _, r := range msg.To {
		recipients = append(recipients, fromRecipient(*r, emptyAddrList))
	}
	for _, r := range append(msg.CC, msg.BCC...) {
		recipients = append(recipients, fromRecipient(*r, msg.To))
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
			Headers: make(map[string]string),
		},
	}

	m.Content.Headers["CC"] = msg.CC.String()

	Log.Debug("Message to send", "msg", m)

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

	Log.Debug("Response", "body", resp.RawText())
	return nil, errors.New("Something wrong happened")

	/*
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(b))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", s.key)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if err != nil {
			return nil, err
		}

		return resp.Body, nil
	*/
}

func init() {
	Log.SetHandler(log.DiscardHandler())
}
