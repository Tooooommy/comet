package comet

import "time"

// Conf comet configuration struct.
type (
	Conf struct {
		WriteWait         time.Duration // Milliseconds until write times out.
		PongWait          time.Duration // Timeout for waiting on pong.
		PingPeriod        time.Duration // Milliseconds between pings.
		MaxMessageSize    int64         // Maximum size in bytes of a message.
		MessageBufferSize uint64        // The max amount of messages that can be in a sessions buffer before it starts dropping them.
	}
)

func newConf() *Conf {
	return &Conf{
		WriteWait:         10 * time.Second,
		PongWait:          60 * time.Second,
		PingPeriod:        (60 * time.Second * 9) / 10,
		MaxMessageSize:    1024,
		MessageBufferSize: 1024,
	}
}
