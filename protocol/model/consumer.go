package model

import (
	"io"
	"sync"

	"github.com/box/memsniff/assembly/reader"
	"github.com/box/memsniff/log"
	"github.com/google/gopacket/tcpassembly"
)

var (
	bufferPool = sync.Pool{New: func() interface{} { return reader.New() }}
	eofSource  *reader.Reader
)

func init() {
	eofSource = reader.New()
	eofSource.ReassemblyComplete()
}

// Reader represents a subset of the bufio.Reader interface.
type Reader interface {
	// Discard skips the next n bytes, returning the number of bytes discarded.
	// If Discard skips fewer than n bytes, it also returns an error.
	Discard(n int) (discarded int, err error)

	// ReadN returns the next n bytes.
	//
	// If EOF is encountered before reading n bytes, the available bytes are returned
	// along with ErrUnexpectedEOF.
	//
	// The returned buffer is only valid until the next call to ReadN or ReadLine.
	ReadN(n int) ([]byte, error)

	// IndexAny returns the result of bytes.IndexAny invoked on the available buffer.
	// If the delimiters are not found and the stream is at its end, returns io.UnexpectedEOF.
	IndexAny(chars string) (int, error)

	// PeekN returns the next n bytes, not advancing the read cursor.
	//
	// If EOF is encountered before reading n bytes, the available bytes are returned
	// along with ErrUnexpectedEOF.
	//
	// The returned buffer is only valid until the next call to ReadN or ReadLine.
	PeekN(n int) ([]byte, error)

	// ReadLine returns a single line, not including the end-of-line bytes.
	// The returned buffer is only valid until the next call to ReadN or ReadLine.
	// ReadLine either returns a non-nil line or it returns an error, never both.
	//
	// The text returned from ReadLine does not include the line end ("\r\n" or "\n").
	// No indication or error is given if the input ends without a final line end.
	ReadLine() ([]byte, error)

	// Reset discards all state, preparing the Reader to receive data from a new connection.
	Reset()

	// Truncate discards all buffered data from the reader, leaving other state intact.
	Truncate()
}

// ConsumerSource buffers tcpassembly.Stream data and exposes it as a closeable Reader.
type ConsumerSource interface {
	Reader
	io.Closer
	tcpassembly.Stream
}

// Consumer is a generic reader of a datastore conversation.
type Consumer struct {
	// Handler receives events derived from the conversation.
	Handler EventHandler
	// ClientReader exposes data sent by the client to the server.
	ClientReader *reader.Reader
	// ServerReader exposes data send by the server to the client.
	ServerReader *reader.Reader

	Fsm Fsm

	logger   log.Logger
	eventBuf []Event
}

func New(logger log.Logger, handler EventHandler, fsm Fsm) *Consumer {
	cr := bufferPool.Get().(*reader.Reader)
	// logger.Log("using Reader", cr.ID, "for client")
	sr := bufferPool.Get().(*reader.Reader)
	// logger.Log("using Reader", sr.ID, "for server")
	c := &Consumer{
		Handler:      handler,
		ClientReader: cr,
		ServerReader: sr,
		Fsm:          fsm,
		logger:       logger,
	}
	fsm.SetConsumer(c)
	return c
}

func (c *Consumer) AddEvent(evt Event) {
	if c.eventBuf == nil {
		c.eventBuf = make([]Event, 0, 8)
	}
	c.eventBuf = append(c.eventBuf, evt)
	if len(c.eventBuf) == cap(c.eventBuf) {
		c.FlushEvents()
	}
}

func (c *Consumer) FlushEvents() {
	c.Handler(c.eventBuf)
	c.eventBuf = c.eventBuf[:0]
}

func (c *Consumer) Close() {
	if c.ClientReader != eofSource {
		// c.ClientReader.Reset()
		// c.logger.Log("returning", c.ClientReader.ID, "to pool after Close")
		// bufferPool.Put(c.ClientReader)
		c.ClientReader = eofSource
	}
	if c.ServerReader != eofSource {
		// c.ServerReader.Reset()
		// c.logger.Log("returning", c.ServerReader.ID, "to pool after Close")
		// bufferPool.Put(c.ServerReader)
		c.ServerReader = eofSource
	}
	c.Fsm = noopFsm{}
}

func (c *Consumer) ClientStream() tcpassembly.Stream {
	return (*ClientStream)(c)
}

func (c *Consumer) ServerStream() tcpassembly.Stream {
	return (*ServerStream)(c)
}

// ClientStream is a view on a Consumer that consumes tcpassembly data from the client
type ClientStream Consumer

func (cs *ClientStream) Reassembled(rs []tcpassembly.Reassembly) {
	for _, r := range rs {
		cs.ClientReader.Reassembled([]tcpassembly.Reassembly{r})
		(*Consumer)(cs).Fsm.Run()
	}
}

func (cs *ClientStream) ReassemblyComplete() {
	cs.ClientReader.ReassemblyComplete()
	(*Consumer)(cs).FlushEvents()
	if cs.ClientReader != eofSource {
		// cs.ClientReader.Reset()
		// cs.logger.Log("returning", cs.ClientReader.ID, "to pool after ReassemblyComplete")
		// bufferPool.Put(cs.ClientReader)
		cs.ClientReader = eofSource
	}
}

// ServerStream is a view on a Consumer that consumes tcpassembly data from the server
type ServerStream Consumer

func (ss *ServerStream) Reassembled(rs []tcpassembly.Reassembly) {
	for _, r := range rs {
		ss.ServerReader.Reassembled([]tcpassembly.Reassembly{r})
		(*Consumer)(ss).Fsm.Run()
	}
}

func (ss *ServerStream) ReassemblyComplete() {
	ss.ServerReader.ReassemblyComplete()
	(*Consumer)(ss).FlushEvents()
	if ss.ServerReader != eofSource {
		// ss.ServerReader.Reset()
		// ss.logger.Log("returning", ss.ServerReader.ID, "to pool after ReassemblyComplete")
		// bufferPool.Put(ss.ServerReader)
		ss.ServerReader = eofSource
	}
}
