package main

import (
	"flag"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/mail"
	"os"
	"os/exec"
	"path"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/labstack/echo"
	"github.com/labstack/echo/engine/standard"
	"github.com/labstack/echo/middleware"
	"github.com/schmooser/go-echolog15"
	log "gopkg.in/inconshreveable/log15.v2"

	"github.com/schmooser/go-mailer/message"
	"github.com/schmooser/go-mailer/provider"
	"github.com/schmooser/go-mailer/provider/mailgun"
	"github.com/schmooser/go-mailer/provider/mandrill"
	"github.com/schmooser/go-mailer/provider/sparkpost"
)

var (
	// Config is a configuration struct.
	Config struct {
		Main struct {
			Provider string `toml:"provider"`
		}
		Log struct {
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

	// Log is a logger variable.
	Log = log.New()
	// LogHandler handles log format.
	LogHandler log.Handler
)

func init() {
	configFile := flag.String("config", "config.toml", "configuration file")
	flag.Parse()

	configData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		Log.Crit("Reading config file", "err", err)
		os.Exit(1)
	}

	_, err = toml.Decode(string(configData), &Config)
	if err != nil {
		Log.Crit("Decoding config file", "err", err)
		os.Exit(1)
	}

	// setting log handlers from configuration parameters
	logLvlFile, err := log.LvlFromString(Config.Log.FileLogLevel)
	if err != nil {
		Log.Crit("Wrong log file level", "err", err)
		os.Exit(1)
	}

	logLvlTerm, err := log.LvlFromString(Config.Log.TermLogLevel)
	if err != nil {
		Log.Crit("Wrong log term level", "err", err)
		os.Exit(1)
	}

	LogHandler = log.MultiHandler(
		log.LvlFilterHandler(
			logLvlTerm,
			//log.StreamHandler(os.Stderr, log.LogfmtFormat())),
			log.StdoutHandler),
		log.LvlFilterHandler(
			logLvlFile,
			log.Must.FileHandler(Config.Log.FileName, log.JsonFormat())))
	Log.SetHandler(LogHandler)

	Log.Debug("Configuration parameters", "config", Config)

	// seeding random
	rand.Seed(time.Now().Unix())
}

// JSONResponse is a struct which is returned to user.
type JSONResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// JSONError sends json error to context.
func JSONError(c echo.Context, code int, message string) error {
	c.JSON(code, JSONResponse{
		Status:  "error",
		Message: message,
	})
	err := echo.NewHTTPError(code, message)
	return err
}

// CheckToken is a middleware which checks token passed either in header or in post data.
func CheckToken() echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			token := c.Request().Header().Get("Authorization")
			if token == "" {
				return JSONError(c, http.StatusBadRequest, "token not provided")
			}
			if _, ok := Config.Tokens[token]; !ok {
				return JSONError(c, http.StatusUnauthorized, "invalid token")
			}
			return next.Handle(c)
		})
	}
}

func sendHandler(p provider.Provider) echo.HandlerFunc {
	return func(c echo.Context) error {
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

		if err := c.Bind(&data); err != nil {
			Log.Error("Binding data failed", "err", err)
			return JSONError(c, http.StatusBadRequest, "can't understand provided data")
		}

		Log.Debug("Incoming data", "data", data)

		// Sending message asynchrounously
		go func() {
			// handling mjml template
			if data.HTML == "" && data.MJML != "" {

				// save mjml content as temporary file
				filename := path.Join(os.TempDir(), "mail-"+strconv.Itoa(rand.Int()))
				ioutil.WriteFile(filename+".mjml", []byte(data.MJML), 0644)

				// run mjml to convert to html
				cmd := exec.Command("mjml", "-r", filename+".mjml", "-o", filename+".html")
				err := cmd.Run()
				if err != nil {
					Log.Error("mjml conversion failed", "err", err, "mjml", filename+".mjml")
				} else {
					var b []byte
					b, err := ioutil.ReadFile(filename + ".html")
					if err != nil {
						Log.Error("reading mjml-converted file failed", "err", err, "html", filename+".html")
					} else {
						Log.Debug("mjml converted to html successfully", "mjml", filename+".mjml", "html", filename+".html")
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
					Log.Error("Sending message failed", "err", err, "iteration", i+1)
					time.Sleep(time.Duration(10*i) * time.Second)
					Log.Debug("Trying to send message again", "iteration", i+2)
					continue
				}
				Log.Debug("Provider response", "resp", resp)
				return
			}
			Log.Error("Message was not sent")
		}()

		// message was sent - returning
		return c.JSON(http.StatusOK, JSONResponse{
			Status:  "ok",
			Message: "email added to the queue",
		})
	}
}

func main() {

	// provider
	var p provider.Provider
	var err error

	switch Config.Main.Provider {

	case "mailgun":
		p = mailgun.New(Config.Mailgun.User, Config.Mailgun.Pass, Config.Mailgun.Server)
	case "mandrill":
		p, err = mandrill.New(Config.Mandrill.Key)
		if err != nil {
			Log.Error("Mandrill instance creation failed", "err", err)
			os.Exit(1)
		}
	case "sparkpost":
		p, err = sparkpost.New(Config.SparkPost.Key)
		if err != nil {
			Log.Error("SparkPost instance creation failed", "err", err)
			os.Exit(1)
		}
		sparkpost.Log.SetHandler(LogHandler)
	default:
		Log.Error("Provider is not supported", "provider", Config.Main.Provider)
		os.Exit(1)
	}

	// echo
	e := echo.New()

	e.Use(echolog15.Logger(Log))
	e.Use(middleware.Recover())
	e.Use(CheckToken())
	e.SetHTTPErrorHandler(echolog15.HTTPErrorHandler(Log))

	e.Post("/send", sendHandler(p))
	e.Run(standard.New(Config.WebServer.Host + ":" + Config.WebServer.Port))
}
