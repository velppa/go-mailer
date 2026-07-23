package model

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
)

// SendRequest is a body of POST /send request.
type SendRequest struct {
	MJML string
	Message
}

func (sr SendRequest) GetReader() io.Reader {
	b, err := json.Marshal(sr)
	if err != nil {
		slog.
			With("module", "go-mailer/model").
			Error("failed to marshal", "value", sr, "err", err)
		return nil
	}
	return bytes.NewReader(b)
}

// Response is a response body from the API.
type Response struct{ Message string }
