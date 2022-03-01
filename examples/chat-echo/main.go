package main

import (
	"github.com/Tooooommy/comet"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"net/http"
)

func main() {
	e := echo.New()
	m := comet.New()
	h := comet.NewHub()
	m.HandleConnect(func(session *comet.Session) {
		h.Register(session)
	})

	m.HandleDisconnect(func(session *comet.Session) {
		h.Unregister(session)
	})

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		http.ServeFile(c.Response(), c.Request(), "index.html")
		return nil
	})

	e.GET("/ws", func(c echo.Context) error {
		return comet.HandleRequest(m)(c.Response(), c.Request())
	})

	m.HandleMessage(func(s *comet.Session, msg []byte) {
		_ = h.Broadcast(msg)
	})

	_ = e.Start(":5000")
}
