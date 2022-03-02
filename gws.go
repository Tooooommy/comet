package comet

import (
	"github.com/gorilla/websocket"
	"net/http"
)

type HandlerFunc func(http.ResponseWriter, *http.Request) error

// gorilla websocket conn
type gConn struct {
	*websocket.Conn
}

func NewGConn(conn *websocket.Conn) Conn  {
	return &gConn{conn}
}

func HandleGws(m *Comet) HandlerFunc  {
	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}
	return func(writer http.ResponseWriter, request *http.Request) error {
		conn, err := upgrader.Upgrade(writer, request, writer.Header())
		if err != nil {
			return err
		}

		keys := map[string]interface{}{}
		for k, v := range request.Header {
			keys[k] = v
		}

		m.Handle(NewGConn(conn), keys)
		return nil
	}
}