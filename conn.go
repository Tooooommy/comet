package comet

import (
	"net"
	"time"
)

// The message types are defined in RFC 6455, section 11.8.
const (
	// NoFrame
	NoFrame = -1

	// TextMessage denotes a text data message. The text message payload is
	// interpreted as UTF-8 encoded text data.
	TextMessage = 1

	// BinaryMessage denotes a binary data message.
	BinaryMessage = 2

	// CloseMessage denotes a close control message. The optional message
	// payload contains a numeric code and text. Use the FormatCloseMessage
	// function to format a close message payload.
	CloseMessage = 8

	// PingMessage denotes a ping control message. The optional message payload
	// is UTF-8 encoded text.
	PingMessage = 9

	// PongMessage denotes a pong control message. The optional message payload
	// is UTF-8 encoded text.
	PongMessage = 10

)

const (
	CloseNoStatusReceived        = 0
)



// 仿照WebSocket Conn
type Conn interface {
	LocalAddr() net.Addr
	RemoteAddr() net.Addr
	SetWriteDeadline(time.Time) error
	WriteMessage(int, []byte) error
	Close() error
	SetReadLimit(int64)
	SetReadDeadline(time.Time) error
	ReadMessage()(int, []byte, error)
	SetPongHandler(func(string) error)
	SetPingHandler(func(string) error)
	SetCloseHandler(func(int, string) error)
}
