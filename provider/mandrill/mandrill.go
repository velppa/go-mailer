package mandrill

import (
	mmandrill "github.com/mostafah/mandrill"

	"github.com/schmooser/go-mailer/message"
)

// Mandrill defines Mandrill transactional mail provider.
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

// Send sends provided message in async or sync way.
func (m *Mandrill) Send(msg *message.Message, async bool) (interface{}, error) {
	mm := mmandrill.NewMessage()
	mm.Subject = msg.Subject
	mm.Text = msg.Text
	mm.HTML = msg.HTML
	mm.FromEmail = msg.From.Address
	mm.FromName = msg.From.Name
	for _, rcpt := range msg.To {
		mm.AddRecipientType(rcpt.Address, rcpt.Name, mmandrill.RecipientTo)
	}
	for _, rcpt := range msg.CC {
		mm.AddRecipientType(rcpt.Address, rcpt.Name, mmandrill.RecipientCC)
	}
	for _, rcpt := range msg.BCC {
		mm.AddRecipientType(rcpt.Address, rcpt.Name, mmandrill.RecipientBCC)
	}
	res, err := mm.Send(async)
	return res, err
}
