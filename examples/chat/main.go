package main

import (
	"github.com/Tooooommy/comet"
	"github.com/gin-gonic/gin"
	"net/http"
)

func main() {
	r := gin.Default()
	m := comet.New()
	h := comet.NewHub()
	m.HandleConnect(func(session *comet.Session) {
		h.Register(session)
	})
	m.HandleDisconnect(func(session *comet.Session) {
		h.Unregister(session)
	})

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		_ = m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleMessage(func(s *comet.Session, msg []byte) {
		_ = h.Broadcast(msg)
	})

	_ = r.Run(":5000")
}
