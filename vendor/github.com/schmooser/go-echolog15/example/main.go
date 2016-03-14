package main

import (
	"errors"
	"net/http"

	"github.com/labstack/echo"
	"github.com/schmooser/go-echolog15"
	"gopkg.in/inconshreveable/log15.v2"
)

// Handler
func hello(c *echo.Context) error {
	return c.String(http.StatusOK, "Hello, World!\n")
}

func main() {
	// Echo instance
	e := echo.New()

	// logger
	log := log15.New()
	// Logger middleware
	e.Use(echolog15.Logger(log))

	// Set error handler
	e.SetHTTPErrorHandler(echolog15.HTTPErrorHandler(log))

	// Routes
	e.Get("/", hello)

	// Routes
	e.Get("/error", func(c *echo.Context) error {
		err := errors.New("Some test error")
		return err
	})

	// Start server
	e.Run(":1323")
}
