package comet

import "time"

// Config comet configuration struct.
type Config struct {
	WriteWait           time.Duration // Milliseconds until write times out.
	PongWait            time.Duration // Timeout for waiting on pong.
	PingPeriod          time.Duration // Milliseconds between pings.
	MaxMessageSize      int64         // Maximum size in bytes of a message.
	MessageBufferSize   uint64        // The max amount of messages that can be in a sessions buffer before it starts dropping them.
	BroadcastBufferSize uint64        // The max amount of broadcast that can be in a broadcast buffer before it starts dropping them.
}

func newConfig() *Config {
	return &Config{
		WriteWait:           10 * time.Second,
		PongWait:            60 * time.Second,
		PingPeriod:          (60 * time.Second * 9) / 10,
		MaxMessageSize:      1024,
		MessageBufferSize:   1024,
		BroadcastBufferSize: 1024,
	}
}
