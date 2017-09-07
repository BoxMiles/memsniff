package inferer

import (
	"github.com/box/memsniff/protocol/mctext"
	"github.com/box/memsniff/protocol/model"
	"github.com/box/memsniff/protocol/redis"
)

// Consumer inspects the first byte of a connection and attempts to guess the correct protocol.
type Consumer model.Consumer

// Run reads the conversation and returns at the close of the conversation.
// The client- and server-side readers are closed before Run returns.
func (c *Consumer) Run() {
	defer c.ClientReader.Close()
	defer c.ServerReader.Close()

	first, err := c.ClientReader.Peek(1)
	if err != nil || len(first) < 1 {
		return
	}

	switch first[0] {
	case 0x80:
		(*model.Consumer)(c).Log("memcache binary protocol not currently handled")

	case '*':
		redis.New(model.Consumer(*c)).Run()

	default:
		mctext.New(model.Consumer(*c)).Run()
	}
}
