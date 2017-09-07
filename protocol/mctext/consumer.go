package mctext

import (
	"bytes"
	"github.com/box/memsniff/protocol/model"
	"io"
	"strconv"
	"time"
)

const (
	crlf = "\r\n"
)

// Consumer generates events based on a memcached text protocol conversation.
type Consumer struct {
	model.Consumer
}

func New(c model.Consumer) *Consumer {
	return &Consumer{
		Consumer: c,
	}
}

// Run reads the conversation and returns at the close of the conversation.
// The client- and server-side readers are closed before Run returns.
func (c *Consumer) Run() {
	c.Log("processing as memcache text protocol")
	defer c.FlushEvents()
	for {
		c.Log("awaiting command")
		line, err := c.ClientReader.ReadLine()
		if err == io.EOF {
			return
		}
		if err != nil {
			return
		}
		c.Log("read command:", string(line))
		if len(line) > 0 {
			err = c.handleCommand(line, c.ClientReader.Seen())
			if err != nil {
				return
			}
		}
	}
}

func (c *Consumer) handleCommand(line []byte, start time.Time) (err error) {
	fields := bytes.Split(line, []byte(" "))
	switch string(fields[0]) {
	case "get", "gets":
		if err = c.handleGet(fields[1:], start); err != nil {
			c.Log("error processing stream:", err)
		}
	case "set", "add", "replace", "append", "prepend":
		if err = c.handleSet(fields[1:]); err != nil {
			c.Log("error processing stream:", err)
		}
	case "quit":
	default:
		if err = c.discardResponse(); err != nil {
			c.Log("error processing unknown command", string(line), ":", err)
		}
	}
	return
}

func (c *Consumer) handleGet(fields [][]byte, start time.Time) error {
	if len(fields) < 1 {
		return c.discardResponse()
	}
	for {
		c.Log("awaiting server reply to get")
		line, err := c.ServerReader.ReadLine()
		if err != nil {
			return err
		}
		c.Log("server reply:", string(line))
		fields := bytes.Split(line, []byte(" "))
		if len(fields) >= 4 && bytes.Equal(fields[0], []byte("VALUE")) {
			key := fields[1]
			size, err := strconv.Atoi(string(fields[3]))
			if err != nil {
				return err
			}
			c.Log("discarding value")
			_, err = c.ServerReader.Discard(size + len(crlf))
			if err != nil {
				return err
			}
			c.Log("discarded value")
			evt := model.Event{
				Type:  model.EventGetHit,
				Key:   string(key),
				Size:  size,
				Start: start,
				End:   c.ServerReader.Seen(),
			}
			c.Log("sending event:", evt)
			c.AddEvent(evt)
		} else {
			return nil
		}
	}
}

func (c *Consumer) handleSet(fields [][]byte) error {
	if len(fields) < 4 {
		return c.discardResponse()
	}
	size, err := strconv.Atoi(string(fields[3]))
	if err != nil {
		return c.discardResponse()
	}
	c.Log("discarding", size+len(crlf), "from client")
	_, err = c.ClientReader.Discard(size + len(crlf))
	if err != nil {
		return nil
	}
	c.Log("discarding response from server")
	return c.discardResponse()
}

func (c *Consumer) discardResponse() error {
	line, err := c.ServerReader.ReadLine()
	c.Log("discarded response from server:", string(line))
	return err
}
