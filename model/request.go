package model

// SendRequest is a body of POST /send request.
type SendRequest struct {
	MJML    string
	Message
}

// Response is a response body from the API.
type Response struct{ Message string }
