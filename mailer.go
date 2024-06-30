package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"net/mail"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/velppa/go-mailer/message"
	"github.com/velppa/go-mailer/provider/mailgun"
	"github.com/velppa/go-mailer/provider/mandrill"
	"github.com/velppa/go-mailer/provider/smtp"
	"github.com/velppa/go-mailer/provider/sparkpost"
)

type Config struct {
	Main struct {
		Provider string `toml:"provider"`
	}

	Mailgun   mailgun.Client
	SMTP      smtp.Client
	Mandrill  mandrill.Client
	SparkPost sparkpost.Client

	// Tokens is a map with token as a key and description as a value.
	Tokens    map[string]string
	WebServer struct {
		Host string
		Port string
	}
}

func newConfig() (*Config, error) {
	configFile := flag.String("config", "config.toml", "configuration file")
	flag.Parse()

	configData, err := os.ReadFile(*configFile)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var c Config
	_, err = toml.Decode(string(configData), &c)
	if err != nil {
		return nil, fmt.Errorf("decoding config file: %w", err)
	}

	return &c, nil
}

// jsonMiddleware sets Content-Type header to encoding/jsonis a middleware which checks token passed either in header or in post data.
func jsonMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "encoding/json")
		next(w, req)
	}
}

// authorizationMiddleware is a middleware which checks token passed either in header or in post data.
func authorizationMiddleware(tokens map[string]string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("Authorization")
		if _, ok := tokens[token]; !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Message: "invalid token"})
			return
		}
		next(w, req)
	}
}

// Sender is an interface for transactional mail providers.
type Sender interface {
	// Send sends email message in sync or async regime.
	Send(msg *message.Message, async bool) (any, error)
}

type Response struct{ Message string }

func sendHandler(p Sender) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		// binding incoming data
		data := struct {
			Subject string
			Text    string
			HTML    string
			MJML    string
			From    *mail.Address
			To      []*mail.Address
			CC      []*mail.Address
			BCC     []*mail.Address
		}{}

		if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
			slog.Error("Binding data failed", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(Response{Message: "can't understand provided data"})
			return
		}

		slog.Debug("Incoming data", "data", data)

		// Sending message asynchrounously
		go func() {
			// handling mjml template
			if data.HTML == "" && data.MJML != "" {

				// save mjml content as temporary file
				filename := path.Join(os.TempDir(), "mail-"+strconv.Itoa(rand.Int()))
				os.WriteFile(filename+".mjml", []byte(data.MJML), 0644)

				// run mjml to convert to html
				cmd := exec.Command("mjml", "-r", filename+".mjml", "-o", filename+".html")
				err := cmd.Run()
				if err != nil {
					slog.Error("mjml conversion failed", "err", err, "mjml", filename+".mjml")
				} else {
					var b []byte
					b, err := os.ReadFile(filename + ".html")
					if err != nil {
						slog.Error("reading mjml-converted file failed", "err", err, "html", filename+".html")
					} else {
						slog.Debug("mjml converted to html successfully", "mjml", filename+".mjml", "html", filename+".html")
						data.HTML = string(b)
					}
				}
			}

			msg := &message.Message{
				Subject: data.Subject,
				Text:    data.Text,
				HTML:    data.HTML,
				From:    data.From,
				To:      data.To,
				CC:      data.CC,
				BCC:     data.BCC,
			}

			// trying to send message 10 times
			for i := 0; i < 10; i++ {
				// sending message
				resp, err := p.Send(msg, true)
				if err != nil {
					slog.Error("Sending message failed", "err", err, "iteration", fmt.Sprintf("%d/10", i+1))
					time.Sleep(time.Duration(10*i) * time.Second)
					slog.Debug("Trying to send message again", "iteration", i+2)
					continue
				}
				slog.Info("Provider response", "resp", resp)
				return
			}
			slog.Error("Message was not sent")
		}()

		// message was sent - returning
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Response{Message: "email added to the queue"})
		return
	}
}

func newProvider(config Config) (Sender, error) {
	switch config.Main.Provider {
	case "mailgun":
		return &config.Mailgun, nil
	case "mandrill":
		client := config.Mandrill
		return &client, client.Ping()
	case "sparkpost":
		return &config.SparkPost, nil
	case "smtp":
		return &config.SMTP, nil
	default:
		return nil, fmt.Errorf("Provider %s is not supported", config.Main.Provider)
	}
}

func main() {
	config, err := newConfig()
	if err != nil {
		slog.Error("parsing configuration failed", "err", err)
		os.Exit(1)
	}
	slog.Debug("Configuration parameters", "config", config)

	provider, err := newProvider(*config)
	if err != nil {
		slog.Error("creating provider failed", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /send",
		jsonMiddleware(
			authorizationMiddleware(config.Tokens,
				sendHandler(provider))))
	addr := config.WebServer.Host + ":" + config.WebServer.Port
	server := http.Server{Addr: addr, Handler: mux}

	slog.Info("Mailer started", "address", addr)
	server.ListenAndServe()
}
