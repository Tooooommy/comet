package main

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Tooooommy/comet"
	"github.com/gin-gonic/gin"
)

// GopherInfo contains information about the gopher on screen
type GopherInfo struct {
	ID, X, Y string
}

func main() {
	router := gin.Default()
	m := comet.New()
	h := comet.NewHub()
	gophers := make(map[*comet.Session]*GopherInfo)
	lock := new(sync.Mutex)
	counter := 0

	router.GET("/", func(c *gin.Context) {
		http.ServeFile(c.Writer, c.Request, "index.html")
	})

	router.GET("/ws", func(c *gin.Context) {
		m.HandleRequest(c.Writer, c.Request)
	})

	m.HandleConnect(func(s *comet.Session) {
		h.Register(s)
		lock.Lock()
		for _, info := range gophers {
			s.Write([]byte("set " + info.ID + " " + info.X + " " + info.Y))
		}
		gophers[s] = &GopherInfo{strconv.Itoa(counter), "0", "0"}
		s.Write([]byte("iam " + gophers[s].ID))
		counter++
		lock.Unlock()
	})

	m.HandleDisconnect(func(s *comet.Session) {
		h.Unregister(s)
		lock.Lock()
		h.BroadcastOthers([]byte("dis "+gophers[s].ID), s)
		delete(gophers, s)
		lock.Unlock()
	})

	m.HandleMessage(func(s *comet.Session, msg []byte) {
		p := strings.Split(string(msg), " ")
		lock.Lock()
		info := gophers[s]
		if len(p) == 2 {
			info.X = p[0]
			info.Y = p[1]
			h.BroadcastOthers([]byte("set "+info.ID+" "+info.X+" "+info.Y), s)
		}
		lock.Unlock()
	})

	router.Run(":5000")
}
