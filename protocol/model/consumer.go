package model

import (
	"fmt"
	"github.com/box/memsniff/log"
	"github.com/google/gopacket/tcpassembly"
	"io"
	"time"
)

// EventType described what sort of event has occurred.
type EventType int

//go:generate stringer -type=EventType
const (
	// EventUnknown is an unhandled event.
	EventUnknown EventType = iota
	// EventCommandExecuted is any command, useful for latency analysis
	EventCommandExecuted
	// EventGetHit is a successful data retrieval that returned data.
	EventGetHit
	// EventGetMiss is a data retrieval that did not result in data.
	EventGetMiss
)

// Reader represents a subset of the bufio.Reader interface.
type Reader interface {
	io.Reader

	// Discard skips the next n bytes, returning the number of bytes discarded.
	// If Discard skips fewer than n bytes, it also returns an error.
	Discard(n int) (discarded int, err error)

	// Peek returns up to n bytes of the next data from r, without
	// advancing the stream.
	Peek(n int) ([]byte, error)

	// ReadN returns the next n bytes.
	//
	// If EOF is encountered before reading n bytes, the available bytes are returned
	// along with ErrUnexpectedEOF.
	//
	// The returned buffer is only valid until the next call to ReadN or ReadLine.
	ReadN(n int) ([]byte, error)

	// ReadLine returns a single line, not including the end-of-line bytes.
	// The returned buffer is only valid until the next call to ReadN or ReadLine.
	// ReadLine either returns a non-nil line or it returns an error, never both.
	//
	// The text returned from ReadLine does not include the line end ("\r\n" or "\n").
	// No indication or error is given if the input ends without a final line end.
	ReadLine() ([]byte, error)

	// Seen returns the timestamp at which the last packet read was captured.
	// This usually reflects the end of the data returned from the most recent read operation.
	Seen() time.Time
}

// ConsumerSource buffers tcpassembly.Stream data and exposes it as a closeable Reader.
type ConsumerSource interface {
	Reader
	io.Closer
	tcpassembly.Stream
}

// Consumer is a generic reader of a datastore conversation.
type Consumer struct {
	// A Logger instance for debugging.  No logging is done if nil.
	Logger log.Logger
	// Handler receives events derived from the conversation.
	Handler EventHandler
	// ClientReader exposes data sent by the client to the server.
	ClientReader ConsumerSource
	// ServerReader exposes data send by the server to the client.
	ServerReader ConsumerSource

	eventBuf []Event
}

// AddEvent buffers an event to be sent to the event handler.
// If the buffer fills, all events in the buffer are flushed immediately.
func (c *Consumer) AddEvent(evt Event) {
	if c.eventBuf == nil {
		c.eventBuf = make([]Event, 0, 128)
	}
	c.eventBuf = append(c.eventBuf, evt)
	if len(c.eventBuf) == cap(c.eventBuf) {
		c.FlushEvents()
	}
}

// FlushEvents immediately sends all events in the buffer to the event handler.
func (c *Consumer) FlushEvents() {
	if len(c.eventBuf) > 0 {
		c.Handler(c.eventBuf)
		c.eventBuf = c.eventBuf[:0]
	}
}

func (c *Consumer) Log(items ...interface{}) {
	if c.Logger != nil {
		c.Logger.Log(items...)
	}
}

// Event is a single event in a datastore conversation
type Event struct {
	// Type of the event.
	Type EventType
	// Datastore key affected by this event.
	Key string
	// Size of the datastore value affected by this event.
	Size int
	// StartTime is the timestamp when the end of the request was captured.
	Start time.Time
	// EndTime is the timestamp when the end of the response was captured.
	End time.Time
}

// Duration returns the amount of time the event took from end of request to end of response.
func (e Event) Duration() time.Duration {
	return e.End.Sub(e.Start)
}

func (e Event) String() string {
	return fmt.Sprintf("{%v %v(%d) %v}", e.Type, e.Key, e.Size, e.Duration())
}

// EventHandler consumes a single event.
type EventHandler func(evts []Event)
