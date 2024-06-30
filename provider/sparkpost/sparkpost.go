package sparkpost

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jmcvetta/napping"

	"github.com/velppa/go-mailer/message"
)

// Client defines Client transactional mail provider.
type Client struct {
	Key string
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

// Send sends provided message in async or sync way.
func (sp *Client) Send(msg *message.Message, async bool) (any, error) {

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
			Headers: make(map[string]string),
		},
	}

	if msg.CC.String() != "" {
		m.Content.Headers["CC"] = msg.CC.String()
	}

	b, _ := json.Marshal(m)
	slog.Debug("Message to send", "msg", string(b))

	headers := make(http.Header)
	headers.Add("Authorization", sp.Key)
	s := napping.Session{
		Header: &headers,
	}

	var result any
	resp, err := s.Post(apiURL, &m, &result, nil)
	if err != nil {
		return nil, err
	}

	if resp.Status() == 200 {
		return result, nil
	}

	slog.Error("Response", "status", resp.Status(), "body", resp.RawText())
	return nil, errors.New("Non-200 status returned")
}
