package main

import (
	"encoding/json"
	"flag"
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
	"github.com/velppa/go-mailer/provider/sparkpost"
)

var (
	// Config is a configuration struct.
	Config struct {
		Main struct {
			Provider string `toml:"provider"`
		}
		slog struct {
			FileName     string `toml:"logfile"`
			FileLogLevel string `toml:"file-level"`
			TermLogLevel string `toml:"term-level"`
		}
		Mailgun struct {
			User   string
			Pass   string
			Server string
		}
		Mandrill struct {
			Key string
		}
		SparkPost struct {
			Key string
		}
		// Tokens is a map with token as a key and description as a value.
		Tokens    map[string]string
		WebServer struct {
			Host string
			Port string
		}
	}
)

func init() {
	configFile := flag.String("config", "config.toml", "configuration file")
	flag.Parse()

	configData, err := os.ReadFile(*configFile)
	if err != nil {
		slog.Error("Reading config file", "err", err)
		os.Exit(1)
	}

	_, err = toml.Decode(string(configData), &Config)
	if err != nil {
		slog.Error("Decoding config file", "err", err)
		os.Exit(1)
	}

	slog.Debug("Configuration parameters", "config", Config)
}

// AuthorizationMiddleware is a middleware which checks token passed either in header or in post data.
func AuthorizationMiddleware(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		token := req.Header.Get("Authorization")
		if _, ok := Config.Tokens[token]; !ok {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(Response{Message: "invalid token"})
			return
		}
		f(w, req)
	}
}

// Provider is an interface for transactional mail providers.
type Provider interface {
	// Send sends email message in sync or async regime.
	Send(msg *message.Message, async bool) (any, error)
}

type Response struct{ Message string }

func sendHandler(p Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Add("Content-Type", "encoding/json")

		// binding incoming data
		data := struct {
			Subject string          `json:"subject"`
			Text    string          `json:"text"`
			HTML    string          `json:"html"`
			MJML    string          `json:"mjml"`
			From    *mail.Address   `json:"from"`
			To      []*mail.Address `json:"to"`
			CC      []*mail.Address `json:"cc"`
			BCC     []*mail.Address `json:"bcc"`
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
					slog.Error("Sending message failed", "err", err, "iteration", i+1)
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

func main() {

	// provider
	var p Provider
	var err error

	switch Config.Main.Provider {

	case "mailgun":
		p = mailgun.New(Config.Mailgun.User, Config.Mailgun.Pass, Config.Mailgun.Server)
	case "mandrill":
		p, err = mandrill.New(Config.Mandrill.Key)
		if err != nil {
			slog.Error("Mandrill instance creation failed", "err", err)
			os.Exit(1)
		}
	case "sparkpost":
		p, err = sparkpost.New(Config.SparkPost.Key)
		if err != nil {
			slog.Error("SparkPost instance creation failed", "err", err)
			os.Exit(1)
		}
		sparkpost.Log.SetHandler(slog.Handler)
	default:
		slog.Error("Provider is not supported", "provider", Config.Main.Provider)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/send", AuthorizationMiddleware(sendHandler(p)))
	server := http.Server{Addr: Config.WebServer.Host + ":" + Config.WebServer.Port, Handler: mux}
	server.ListenAndServe()
}
