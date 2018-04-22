package model

import (
	"fmt"
	"net"
	"time"
)

type NetClient struct {
	LastMessageTime time.Time
	Out             chan []byte
	In              chan []byte
	Conn            net.Conn
}

func (c *NetClient) SendString(message string) {
	c.Out <- []byte(message)
}

// SendString bytes to client
func (c *NetClient) SendBytes(b []byte) {
	c.In <- b
}

func (c *NetClient) CloseChannels() {
	fmt.Println("closing channels")
	close(c.Out)
	close(c.In)
}
