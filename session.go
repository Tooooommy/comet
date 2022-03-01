package comet

import (
	"errors"
	"github.com/labstack/gommon/log"
	"net/http"
	"sync"
	"time"
)

// Session wrapper around websocket connections.
type Session struct {
	req     *http.Request
	keys    map[string]interface{}
	conn    Conn
	buffer  *RingBuffer
	comet   *Comet
	open    bool
	rwmutex *sync.RWMutex
}

func (s *Session) writeMessage(message *envelope) {
	if s.closed() {
		s.comet.errorHandler(s, errors.New("tried to write to Closed a session"))
		return
	}
	if err := s.buffer.Put(message); err != nil {
		s.comet.errorHandler(s, errors.New("session message buffer is full"))
	}
}

func (s *Session) writeRaw(message *envelope) error {
	if s.closed() {
		return errors.New("tried to write to a Closed session")
	}

	s.conn.SetWriteDeadline(time.Now().Add(s.comet.Config.WriteWait))
	err := s.conn.WriteMessage(message.t, message.msg)

	if err != nil {
		log.Error("WriteRaw")
		return err
	}

	return nil
}

func (s *Session) closed() bool {
	s.rwmutex.RLock()
	closed := !s.open
	s.rwmutex.RUnlock()
	return closed
}

func (s *Session) close() {
	if !s.closed() {
		s.rwmutex.Lock()
		s.open = false
		s.buffer.Dispose()
		_ = s.conn.Close()
		s.rwmutex.Unlock()
	}
}

func (s *Session) ping() {
	_ = s.writeRaw(&envelope{t: PingMessage, msg: []byte{}})
}

func (s *Session) writePump() {
	ticker := time.NewTicker(s.comet.Config.PingPeriod)
	defer ticker.Stop()
	for {
		msg, err := s.buffer.Get(s.comet.Config.PingPeriod)
		if err == ErrDisposed || err == ErrPanic {
			break
		}
		if err == ErrTimeout {
			s.ping()
			continue
		}

		err = s.writeRaw(msg)

		if err != nil {
			s.comet.errorHandler(s, err)
			break
		}

		if msg.t == CloseMessage {
			break
		}

		if msg.t == TextMessage {
			s.comet.messageSentHandler(s, msg.msg)
		}

		if msg.t == BinaryMessage {
			s.comet.messageSentHandlerBinary(s, msg.msg)
		}
	}
}

func (s *Session) readPump() {
	s.conn.SetReadLimit(s.comet.Config.MaxMessageSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(s.comet.Config.PongWait))

	s.conn.SetPongHandler(func(string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(s.comet.Config.PongWait))
		s.comet.pongHandler(s)
		return nil
	})

	if s.comet.closeHandler != nil {
		s.conn.SetCloseHandler(func(code int, text string) error {
			return s.comet.closeHandler(s, code, text)
		})
	}

	for {
		t, message, err := s.conn.ReadMessage()

		if err != nil {
			s.comet.errorHandler(s, err)
			break
		}

		if t == TextMessage {
			s.comet.messageHandler(s, message)
		}

		if t == BinaryMessage {
			s.comet.messageHandlerBinary(s, message)
		}
	}
}

// Write writes message to session.
func (s *Session) Write(msg []byte) error {
	if s.closed() {
		return errors.New("session is Closed")
	}

	s.writeMessage(&envelope{t: TextMessage, msg: msg})

	return nil
}

// WriteBinary writes a binary message to session.
func (s *Session) WriteBinary(msg []byte) error {
	if s.closed() {
		return errors.New("session is Closed")
	}

	s.writeMessage(&envelope{t: BinaryMessage, msg: msg})

	return nil
}

// Close closes session.
func (s *Session) Close() error {
	if s.closed() {
		return errors.New("session is already Closed")
	}

	s.writeMessage(&envelope{t: CloseMessage, msg: []byte{}})

	return nil
}

// CloseWithMsg closes the session with the provided payload.
// Use the FormatCloseMessage function to format a proper close message payload.
func (s *Session) CloseWithMsg(msg []byte) error {
	if s.closed() {
		return errors.New("session is already Closed")
	}

	s.writeMessage(&envelope{t: CloseMessage, msg: msg})

	return nil
}

// Request is http original Request.
func (s *Session) Request() *http.Request {
	return s.req
}

// Set is used to store a new key/value pair exclusivelly for this session.
// It also lazy initializes s.keys if it was not used previously.
func (s *Session) Set(key string, value interface{}) {
	if s.keys == nil {
		s.keys = make(map[string]interface{})
	}

	s.keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (s *Session) Get(key string) (value interface{}, exists bool) {
	if s.keys != nil {
		value, exists = s.keys[key]
	}

	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (s *Session) MustGet(key string) interface{} {
	if value, exists := s.Get(key); exists {
		return value
	}

	panic("Key \"" + key + "\" does not exist")
}

// IsClosed returns the status of the connection.
func (s *Session) IsClosed() bool {
	return s.closed()
}
