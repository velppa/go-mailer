package mandrill

import (
	mmandrill "github.com/mostafah/mandrill"

	"github.com/schmooser/go-mailer/message"
)

type Mandrill struct {
	apiKey string
}

// New returns new Mandrill instance. API key is validated upon creation,
// returning error if key is not valid.
func New(key string) (*Mandrill, error) {
	m := &Mandrill{
		apiKey: key,
	}
	mmandrill.Key = m.apiKey
	if err := mmandrill.Ping(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Mandrill) Send(msg *message.Message, async bool) (interface{}, error) {
	mm := mmandrill.NewMessage()
	mm.Subject = msg.Subject
	mm.Text = msg.Text
	mm.HTML = msg.HTML
	mm.FromEmail = msg.FromEmail
	mm.FromName = msg.FromName
	for _, rcpt := range msg.Recipients {
		mm.AddRecipient(rcpt.Email, rcpt.Name)
	}
	res, err := mm.Send(async)
	return res, err
}
