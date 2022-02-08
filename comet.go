package comet

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type handleMessageFunc func(*Session, []byte)
type handleErrorFunc func(*Session, error)
type handleCloseFunc func(*Session, int, string) error
type handleSessionFunc func(*Session)
type filterFunc func(*Session) bool

// Comet implements a websocket manager.
type (
	Comet struct {
		Config                   *Conf
		Upgrader                 *websocket.Upgrader
		messageHandler           handleMessageFunc
		messageHandlerBinary     handleMessageFunc
		messageSentHandler       handleMessageFunc
		messageSentHandlerBinary handleMessageFunc
		errorHandler             handleErrorFunc
		closeHandler             handleCloseFunc
		connectHandler           handleSessionFunc
		disconnectHandler        handleSessionFunc
		pongHandler              handleSessionFunc
	}

	CometOpt func(*Conf)
)

// New creates a new comet instance with default Upgrader and CometConf.
func New(options ...CometOpt) *Comet {
	cfg := newConf()
	for _, option := range options {
		option(cfg)
	}

	upgrader := &websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	return &Comet{
		Config:                   cfg,
		Upgrader:                 upgrader,
		messageHandler:           func(*Session, []byte) {},
		messageHandlerBinary:     func(*Session, []byte) {},
		messageSentHandler:       func(*Session, []byte) {},
		messageSentHandlerBinary: func(*Session, []byte) {},
		errorHandler:             func(*Session, error) {},
		closeHandler:             nil,
		connectHandler:           func(*Session) {},
		disconnectHandler:        func(*Session) {},
		pongHandler:              func(*Session) {},
	}
}

// HandleConnect fires fn when a session connects.
func (m *Comet) HandleConnect(fn func(*Session)) {
	m.connectHandler = fn
}

// HandleDisconnect fires fn when a session disconnects.
func (m *Comet) HandleDisconnect(fn func(*Session)) {
	m.disconnectHandler = fn
}

// HandlePong fires fn when a pong is received from a session.
func (m *Comet) HandlePong(fn func(*Session)) {
	m.pongHandler = fn
}

// HandleMessage fires fn when a text message comes in.
func (m *Comet) HandleMessage(fn func(*Session, []byte)) {
	m.messageHandler = fn
}

// HandleMessageBinary fires fn when a binary message comes in.
func (m *Comet) HandleMessageBinary(fn func(*Session, []byte)) {
	m.messageHandlerBinary = fn
}

// HandleSentMessage fires fn when a text message is successfully sent.
func (m *Comet) HandleSentMessage(fn func(*Session, []byte)) {
	m.messageSentHandler = fn
}

// HandleSentMessageBinary fires fn when a binary message is successfully sent.
func (m *Comet) HandleSentMessageBinary(fn func(*Session, []byte)) {
	m.messageSentHandlerBinary = fn
}

// HandleError fires fn when a session has an error.
func (m *Comet) HandleError(fn func(*Session, error)) {
	m.errorHandler = fn
}

// HandleClose sets the handler for close messages received from the session.
// The code argument to h is the received close code or CloseNoStatusReceived
// if the close message is empty. The default close handler sends a close frame
// back to the session.
//
// The application must read the connection to process close messages as
// described in the section on Control Frames above.
//
// The connection read methods return a CloseError when a close frame is
// received. Most applications should handle close messages as part of their
// normal error handling. Applications should only set a close handler when the
// application must perform some action before sending a close frame back to
// the session.
func (m *Comet) HandleClose(fn func(*Session, int, string) error) {
	if fn != nil {
		m.closeHandler = fn
	}
}

// HandleRequest upgrades http requests to websocket connections and dispatches them to be handled by the comet instance.
func (m *Comet) HandleRequest(w http.ResponseWriter, r *http.Request) error {
	return m.HandleRequestWithKeys(w, r, nil)
}

// HandleRequestWithKeys does the same as HandleRequest but populates session.keys with keys.
func (m *Comet) HandleRequestWithKeys(w http.ResponseWriter, r *http.Request, keys map[string]interface{}) error {
	conn, err := m.Upgrader.Upgrade(w, r, w.Header())

	if err != nil {
		return err
	}

	session := &Session{
		req:     r,
		keys:    keys,
		conn:    conn,
		buffer:  NewRingBuffer(m.Config.MessageBufferSize),
		comet:   m,
		open:    true,
		rwmutex: &sync.RWMutex{},
	}

	m.connectHandler(session)

	go session.writePump()

	session.readPump()

	session.close()

	m.disconnectHandler(session)

	return nil
}
