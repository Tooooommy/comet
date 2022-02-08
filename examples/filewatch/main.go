package main

import (
	"github.com/Tooooommy/comet"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func main() {
	file := "file.txt"

	r := gin.Default()
	m := comet.New()
	h := comet.NewHub(1024)
	m.HandleConnect(func(session *comet.Session) {
		h.Register(session)
	})
	m.HandleDisconnect(func(session *comet.Session) {
		h.Unregister(session)
	})
	w, _ := fsnotify.NewWatcher()

	r.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		_ = m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleConnect(func(s *comet.Session) {
		content, _ := ioutil.ReadFile(file)
		s.Write(content)
	})

	go func() {
		for {
			ev := <-w.Events
			if ev.Op == fsnotify.Write {
				content, _ := ioutil.ReadFile(ev.Name)
				h.Broadcast(content)
			}
		}
	}()

	w.Add(file)

	r.Run(":5000")
}
