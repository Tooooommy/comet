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

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	r.GET("/channel/:name", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "chan.html")
	})

	r.GET("/channel/:name/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleConnect(func(session *comet.Session) {
		h.Register(session)
	})

	m.HandleDisconnect(func(session *comet.Session) {
		h.Unregister(session)
	})

	m.HandleMessage(func(s *comet.Session, msg []byte) {
		h.BroadcastFilter(msg, func(q *comet.Session) bool {
			return q.Request().URL.Path == s.Request().URL.Path
		})
	})

	r.Run(":5000")
}
