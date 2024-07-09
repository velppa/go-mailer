package main

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/velppa/go-mailer/message"
)

// echoSender "sends" a message by logging it.
type echoSender struct{ *testing.T }

var expectedTestSubject = "test message"

func (s echoSender) Send(msg *message.Message, async bool) (any, error) {
	if msg.Subject != expectedTestSubject {
		s.Errorf("subject parsed incorrectly, %s != %s", expectedTestSubject, msg.Subject)
	}
	slog.Info("sending message", "msg", msg)
	return "ok", nil
}

func Test_sendHandler(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /send", sendHandler(echoSender{t}))
	svr := httptest.NewServer(mux)

	defer svr.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		svr.URL+"/send",
		bytes.NewBufferString(fmt.Sprintf(`{"subject": "%s"}`, expectedTestSubject)),
	)
	if err != nil {
		t.Fatal("NewRequest failed", "err", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal("request failed", "err", err)
	}

	t.Logf("response: %+v", resp)
}
