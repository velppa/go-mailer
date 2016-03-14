package sparkpost

import (
	sp "github.com/SparkPost/gosparkpost"

	"github.com/schmooser/go-mailer/message"
)

// SparkPost defines SparkPost transactional mail provider.
type SparkPost struct {
	client *sp.Client
}

// New returns new SparkPost instance. API key is validated upon creation,
// returning error if key is not valid.
func New(key string) (*SparkPost, error) {
	api := &SparkPost{
		client: &sp.Client{},
	}
	err := api.client.Init(&sp.Config{
		ApiKey: key,
	})
	if err != nil {
		return nil, err
	}

	return api, nil
}

// Send sends provided message in async or sync way.
func (m *SparkPost) Send(msg *message.Message, async bool) (interface{}, error) {

	// Create a Transmission using an inline Recipient List
	// and inline email Content.
	tx := &sp.Transmission{
		Recipients: append(msg.To.Slice(), msg.CC.Slice()...),
		Content: sp.Content{
			HTML:    msg.HTML,
			From:    msg.From.String(),
			Subject: msg.Subject,
		},
	}

	id, _, err := m.client.Send(tx)
	if err != nil {
		return nil, err
	}

	return id, nil
}
