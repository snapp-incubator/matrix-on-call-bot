package http

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func NewServer() *server {
	s := new(server)
	s.e = echo.New()
	s.e.HideBanner = true
	s.e.HidePort = true

	return s
}

type server struct {
	e *echo.Echo
}

func (s *server) Run(addr string) {
	s.e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	s.e.Logger.Fatal(s.e.Start(addr))
}
