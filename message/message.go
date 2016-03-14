package message

import (
	"net/mail"
	"strings"
)

// Message defines email message.
type Message struct {
	Subject string
	Text    string
	HTML    string
	From    *mail.Address
	To      []*mail.Address
	CC      []*mail.Address
	BCC     []*mail.Address
}

// NewMessage returns new empty Message instance.
func NewMessage() *Message {
	return &Message{}
}

// AddTo adds new recipient to slice of recipients.
func (msg *Message) AddTo(addr *mail.Address) {
	msg.To = append(msg.To, addr)
}

// AddCC adds new recipient to slice of recipients.
func (msg *Message) AddCC(addr *mail.Address) {
	msg.CC = append(msg.CC, addr)
}

// AddBCC adds new recipient to slice of recipients.
func (msg *Message) AddBCC(addr *mail.Address) {
	msg.CC = append(msg.CC, addr)
}

func list(addrs []*mail.Address) string {
	ss := make([]string, len(addrs))
	for i, addr := range addrs {
		ss[i] = addr.String()
	}
	return strings.Join(ss, ", ")
}

// To returns comma-separated recepient emails.
func (msg *Message) To() string {
	return list(msg.To)
}

// CC is the same as To but for CC field.
func (msg *Message) CC() string {
	return list(msg.CC)
}

// BCC is the same as To but for BCC field.
func (msg *Message) BCC() string {
	return list(msg.BCC)
}
