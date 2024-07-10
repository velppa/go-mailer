package model

import (
	"net/mail"
	"strings"
)

// AddressList is a slice of Addresses.
type AddressList []mail.Address

// String returns list as comma-separated values.
func (list AddressList) String() string {
	ss := make([]string, len(list))
	for i, addr := range list {
		ss[i] = addr.String()
	}
	return strings.Join(ss, ", ")
}

// Slice returns slice of email strings.
func (list AddressList) Slice() []string {
	ss := make([]string, len(list))
	for i, addr := range list {
		ss[i] = addr.String()
	}
	return ss
}

type Header struct {
	Key string
	Values []string
}

// Message defines email message.
type Message struct {
	Subject string
	Text    string
	HTML    string
	From    mail.Address
	To      AddressList
	CC      AddressList
	BCC     AddressList
	Headers []Header
}

// NewMessage returns new empty Message instance.
func NewMessage() *Message {
	return &Message{}
}

// AddTo adds new recipient to slice of recipients.
func (msg *Message) AddTo(addr mail.Address) {
	msg.To = append(msg.To, addr)
}

// AddCC adds new recipient to slice of recipients.
func (msg *Message) AddCC(addr mail.Address) {
	msg.CC = append(msg.CC, addr)
}

// AddBCC adds new recipient to slice of recipients.
func (msg *Message) AddBCC(addr mail.Address) {
	msg.CC = append(msg.CC, addr)
}
