package provider

import (
	"github.com/schmooser/go-mailer/message"
)

// Provider is an interface for transactional mail providers.
type Provider interface {
	// Send sends email message in sync or async regime.
	Send(msg *message.Message, async bool) (interface{}, error)
}
