package comet

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)


type AcceptFunc = func(string) error
// tcp conn
type tConn struct {
	net.Conn
	readLimit int64
	handlePong func(string) error
	handlePing func(string) error
	handleClose func(int,string) error
}

func NewTConn(conn net.Conn) Conn {
	c := &tConn{Conn:conn}
	c.SetPongHandler(func(s string) error {return nil})
	c.SetPingHandler(func(s string) error {return nil})
	c.SetCloseHandler(func(i int, s string) error {return nil})
	return c
}

func HandelTcp(m *Comet)  AcceptFunc {
	return func(addr string) error {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return err
		}
		for {
			conn, err := listener.Accept()
			if err != nil {
				return err
			}
			go m.Handle(NewTConn(conn), map[string]interface{}{})
		}
	}
}

func (c *tConn) WriteMessage(_type int, data []byte) error {
	size := len(data)
	buffer := make([]byte, 4+4+size)
	binary.BigEndian.PutUint32(buffer[:4], uint32(size))
	binary.BigEndian.PutUint32(buffer[4:8], uint32(_type))
	copy(buffer[8:], data)
	_, err := c.Write(data)
	return err
}

func (c *tConn) SetReadLimit(size int64) {
	c.readLimit = size
}

func (c *tConn) ReadMessage() (int, []byte, error) {
	header := make([]byte, 4+4)
	_ , err := io.ReadFull(c, header)
	if err !=nil {
		return NoFrame, nil, fmt.Errorf("read header err: %s", err)
	}

	size := binary.BigEndian.Uint32(header[:4])
	if c.readLimit>0 && int64(size) > c.readLimit {
		return NoFrame, nil, fmt.Errorf("the data size %d is beyond the max: %d", size, c.readLimit)
	}

	_type := binary.BigEndian.Uint32(header[4:8])

	data := make([]byte, size)
	_, err = io.ReadFull(c.Conn, data)
	if err != nil {
		return NoFrame, nil, fmt.Errorf("read data err: %s", err)
	}
	switch _type {
	case PingMessage:
		err := c.handlePing(string(data))
		if err != nil {
			return NoFrame, nil, fmt.Errorf("handle ping err: %s", err)
		}
	case PongMessage:
		err := c.handlePong(string(data))
		if err != nil {
			return NoFrame, nil, fmt.Errorf("handle pong err: %s", err)
		}
	case CloseMessage:
		code := CloseNoStatusReceived
		text := ""
		if size >= 2 {
			code = int(binary.BigEndian.Uint16(data))
			text = string(data[2:])
		}
		err := c.handleClose(code, text)
		if err != nil {
			return NoFrame, nil, fmt.Errorf("handle close err: %+v", err)
		}
	}
	return int(_type), data, nil
}

func (c *tConn) SetPongHandler(f func(string) error) {
	if f == nil {
		f = func(s string) error {return nil}
	}
	c.handlePong = f
}

func (c *tConn) SetPingHandler(f func(string) error)  {
	if f == nil {
		f = func(msg string) error {
			_ = c.SetWriteDeadline(time.Now().Add(time.Second))
			err := c.WriteMessage(PingMessage, []byte(msg))
			if e, ok := err.(net.Error); ok && e.Temporary() {
				return nil
			}
			return err
		}
	}
	c.handlePing = f
}

func (c *tConn) SetCloseHandler(f func(int, string) error) {
	if f == nil {
		f = func(code int, text string) error {
			msg := FormatCloseMessage(code, text)
			_ = c.SetWriteDeadline(time.Now().Add(time.Second))
			_ = c.WriteMessage(CloseMessage, msg)
			return nil
		}
	}
	c.handleClose = f
}

// FormatCloseMessage formats closeCode and text as a WebSocket close message.
// An empty message is returned for code CloseNoStatusReceived.
func FormatCloseMessage(code int, text string) []byte {
	if code == CloseNoStatusReceived {
		return []byte{}
	}
	buf := make([]byte, 2+len(text))
	binary.BigEndian.PutUint16(buf, uint16(code))
	copy(buf[2:], text)
	return buf
}