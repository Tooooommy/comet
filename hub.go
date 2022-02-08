package comet

import (
	"errors"
	"github.com/gorilla/websocket"
	"sync"
)

type Hub struct {
	sessions   map[*Session]bool
	broadcast  chan *envelope
	register   chan *Session
	unregister chan *Session
	exit       chan *envelope
	open       bool
	rwmutex    *sync.RWMutex
}

func NewHub(broadcast uint64) *Hub {
	hub := &Hub{
		sessions:   make(map[*Session]bool),
		broadcast:  make(chan *envelope, broadcast),
		register:   make(chan *Session),
		unregister: make(chan *Session),
		exit:       make(chan *envelope),
		open:       true,
		rwmutex:    &sync.RWMutex{},
	}
	go hub.run()
	return hub
}

func (h *Hub) run() {
loop:
	for {
		select {
		case s := <-h.register:
			h.rwmutex.Lock()
			h.sessions[s] = true
			h.rwmutex.Unlock()
		case s := <-h.unregister:
			if _, ok := h.sessions[s]; ok {
				h.rwmutex.Lock()
				delete(h.sessions, s)
				h.rwmutex.Unlock()
			}
		case m := <-h.broadcast:
			h.Range(func(s *Session) {
				if m.filter != nil && !m.filter(s) {
					return
				}
				s.writeMessage(m)
			})
		case m := <-h.exit:
			h.Range(func(s *Session) {
				s.writeMessage(m)
			})

			h.rwmutex.Lock()
			h.sessions = map[*Session]bool{}
			h.open = false
			h.rwmutex.Unlock()
			break loop
		}
	}
}

func (h *Hub) Register(s *Session) {
	if h.Closed() {
		return
	}

	h.register <- s
}

func (h *Hub) Unregister(s *Session) {
	if h.Closed() {
		return
	}

	h.unregister <- s
}

// Closed returns the status of the hub instance.
func (h *Hub) Closed() bool {
	h.rwmutex.RLock()
	closed := !h.open
	h.rwmutex.RUnlock()
	return closed
}

// Online return the number of connected sessions.
func (h *Hub) Online() int {
	h.rwmutex.RLock()
	online := len(h.sessions)
	h.rwmutex.RUnlock()
	return online
}

// Broadcast broadcasts a text message to all sessions.
func (h *Hub) Broadcast(msg []byte) error {
	if h.Closed() {
		return errors.New("hub instance is Closed")
	}

	message := &envelope{t: websocket.TextMessage, msg: msg}
	h.broadcast <- message
	return nil
}

// BroadcastFilter broadcasts a text message to all sessions that fn returns true for.
func (h *Hub) BroadcastFilter(msg []byte, fn func(*Session) bool) error {
	if h.Closed() {
		return errors.New("hub instance is Closed")
	}

	message := &envelope{t: websocket.TextMessage, msg: msg, filter: fn}
	h.broadcast <- message
	return nil
}

// BroadcastOthers broadcasts a text message to all sessions except session s.
func (h *Hub) BroadcastOthers(msg []byte, s *Session) error {
	return h.BroadcastFilter(msg, func(q *Session) bool {
		return s != q
	})
}

// BroadcastMultiple broadcasts a text message to multiple sessions given in the sessions slice.
func (h *Hub) BroadcastMultiple(msg []byte, sessions []*Session) error {
	for _, sess := range sessions {
		if writeErr := sess.Write(msg); writeErr != nil {
			return writeErr
		}
	}
	return nil
}

// BroadcastBinary broadcasts a binary message to all sessions.
func (h *Hub) BroadcastBinary(msg []byte) error {
	if h.Closed() {
		return errors.New("hub instance is Closed")
	}

	message := &envelope{t: websocket.BinaryMessage, msg: msg}
	h.broadcast <- message
	return nil
}

// BroadcastBinaryFilter broadcasts a binary message to all sessions that fn returns true for.
func (h *Hub) BroadcastBinaryFilter(msg []byte, fn func(*Session) bool) error {
	if h.Closed() {
		return errors.New("hub instance is Closed")
	}

	message := &envelope{t: websocket.BinaryMessage, msg: msg, filter: fn}
	h.broadcast <- message
	return nil
}

// BroadcastBinaryOthers broadcasts a binary message to all sessions except session s.
func (h *Hub) BroadcastBinaryOthers(msg []byte, s *Session) error {
	return h.BroadcastBinaryFilter(msg, func(q *Session) bool {
		return s != q
	})
}

// Close closes the hub instance and all connected sessions.
func (h *Hub) Close() error {
	if h.Closed() {
		return errors.New("hub instance is already Closed")
	}

	h.exit <- &envelope{t: websocket.CloseMessage, msg: []byte{}}
	return nil
}

// CloseWithMsg closes the hub instance with the given close payload and all connected sessions.
// Use the FormatCloseMessage function to format a proper close message payload.
func (h *Hub) CloseWithMsg(msg []byte) error {
	if h.Closed() {
		return errors.New("hub instance is already Closed")
	}

	h.exit <- &envelope{t: websocket.CloseMessage, msg: msg}
	return nil
}

func (h *Hub) Range(fn func(s *Session)) {
	h.rwmutex.RLock()
	for session := range h.sessions {
		fn(session)
	}
	h.rwmutex.RUnlock()
}
