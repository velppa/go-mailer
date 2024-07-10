package mandrill

import (
	"fmt"

	mmandrill "github.com/mostafah/mandrill"

	"github.com/velppa/go-mailer/model"
)

// Client defines Client transactional mail provider.
type Client struct {
	ApiKey string
}

func (m Client) Ping() error {
	mmandrill.Key = m.ApiKey
	return mmandrill.Ping()
}

// SendResult encapsulates mostafah/mandrill SendResult.
type SendResult struct {
	*mmandrill.SendResult
}

func (sr SendResult) String() string {
	return fmt.Sprintf("%+v", sr.SendResult)
}

// Send sends provided message in async or sync way.
func (m *Client) Send(msg *model.Message, async bool) (any, error) {
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
	sr := make([]SendResult, len(res))
	for i, r := range res {
		sr[i] = SendResult{r}
	}
	return sr, err
}
