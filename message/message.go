package message

type Provider interface {
	Send(msg Message, async bool) (string, error)
}

// Message defines email message.
type Message struct {
	Subject    string
	Text       string
	HTML       string
	FromEmail  string
	FromName   string
	Recipients []Recipient
}

// NewMessage returns new empty Message instance.
func NewMessage() *Message {
	return &Message{}
}

// AddRecipient adds new recipient to slice of recipients.
func (msg *Message) AddRecipient(email, name string) {
	msg.Recipients = append(msg.Recipients, Recipient{Name: name, Email: email})
}

// Recipient defines recipient of an email.
type Recipient struct {
	Name  string
	Email string
}
