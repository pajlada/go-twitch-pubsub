package twitchpubsub

import (
	"crypto/tls"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/gorilla/websocket"
)

func TestConnectionNoPingResponse(t *testing.T) {
	c := qt.New(t)

	// Run https://github.com/Chatterino/twitch-pubsub-server-test
	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	setConnectionDialer(&dialer)
	messageBus := make(chan sharedMessage, 10)
	conn := newConnection("wss://127.0.0.1:9050/dont-respond-to-ping", messageBus)
	conn.pingInterval = 2 * time.Second
	conn.pongDeadlineTime = 1 * time.Second
	conn.pongDeadlineTime = 1 * time.Second

	c.Assert(conn, qt.IsNotNil)

	err := conn.connect()
	c.Assert(err, qt.IsNil)

	time.Sleep(10 * time.Second)

	c.Assert(conn.numConnects.Load(), qt.Equals, uint64(2))
}
