package echolog15

import (
	"net"
	"net/http"
	//"net/http/httputil"
	"time"

	"github.com/labstack/echo"
	"gopkg.in/inconshreveable/log15.v2"
)

// Logger is a logger middleware for log15 package.
func Logger(l log15.Logger) echo.MiddlewareFunc {
	return func(next echo.Handler) echo.Handler {
		return echo.HandlerFunc(func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			/*
				l.Debug("Echo request", "req", func() string {
					dump, _ := httputil.DumpRequest(req, true)
					return string(dump)
				}())
			*/

			remoteAddr := req.RemoteAddress()
			if ip := req.Header().Get(echo.XRealIP); ip != "" {
				remoteAddr = ip
			} else if ip = req.Header().Get(echo.XForwardedFor); ip != "" {
				remoteAddr = ip
			} else {
				remoteAddr, _, _ = net.SplitHostPort(remoteAddr)
			}

			start := time.Now()
			if err := next.Handle(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			path := req.URL().Path()
			if path == "" {
				path = "/"
			}

			l.Info("Echo response",
				"remoteAddr", remoteAddr,
				"method", req.Method(),
				"path", path,
				"response", res.Status(),
				"time", stop.Sub(start),
				"size", res.Size())
			return nil
		})
	}
}

// HTTPErrorHandler is an error handler with log15 support.
func HTTPErrorHandler(l log15.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		code := http.StatusInternalServerError
		msg := http.StatusText(code)
		if he, ok := err.(*echo.HTTPError); ok {
			code = he.Code
			msg = he.Message
		}
		if !c.Response().Committed() {
			c.String(code, msg)
		}
		if err.Error() != msg {
			l.Error("Echo error", "err", err, "code", code, "msg", msg)
		} else {
			l.Error("Echo error", "code", code, "msg", msg)
		}
	}
}
