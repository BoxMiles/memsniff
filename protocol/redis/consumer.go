package redis

import (
	"github.com/box/memsniff/protocol/model"
	goredis "github.com/go-redis/redis"
	"strings"
	"time"
)

type Consumer struct {
	model.Consumer
	client *Reader
	server *Reader
}

func New(c model.Consumer) *Consumer {
	return &Consumer{
		Consumer: c,
		client:   NewReader(c.ClientReader),
		server:   NewReader(c.ServerReader),
	}
}

func (c *Consumer) Run() {
	c.Log("processing as redis protocol")
	defer c.FlushEvents()
	for {
		c.Log("awaiting command")
		x, err := c.client.ReadArrayReply(stringSliceParser)
		if err != nil {
			return
		}
		start := c.ClientReader.Seen()
		req := x.([]string)
		c.Log("got command", req)

		if len(req) < 1 {
			return
		}

		switch strings.ToUpper(req[0]) {
		case "GET":
			err = c.handleGet(req, start)
		default:
			err = c.discardResponse(req, start)
		}

		if err != nil {
			c.Log("error while discarding response:", err)
			return
		}
	}
}

func (c *Consumer) handleGet(req []string, start time.Time) error {
	bytes, err := c.server.ReadTmpBytesReply()
	after := c.ServerReader.Seen()
	if err == goredis.Nil {
		c.AddEvent(model.Event{
			Type:  model.EventGetMiss,
			Key:   req[1],
			Start: start,
			End:   after,
		})

		return nil
	}

	if err != nil {
		return err
	}

	c.AddEvent(model.Event{
		Type:  model.EventGetHit,
		Key:   req[1],
		Size:  len(bytes),
		Start: start,
		End:   after,
	})

	return nil
}

func (c *Consumer) discardResponse(req []string, start time.Time) error {
	// read and discard all forms of reply
	_, err := c.server.ReadReply(sliceParser)
	if err != nil {
		return err
	}

	after := c.ServerReader.Seen()

	evt := model.Event{
		Type:  model.EventCommandExecuted,
		Key:   strings.Join(req[1:], " "),
		Start: start,
		End:   after,
	}
	c.AddEvent(evt)

	return nil
}
