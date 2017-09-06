// Copied directly from github.com/go-redis/redis/internal/proto
// since internal packages cannot be accessed from another project.

package redis

import (
	"errors"
	"fmt"
	"github.com/box/memsniff/protocol/model"
	"reflect"
	"strconv"
	"unsafe"
)

const bytesAllocLimit = 1024 * 1024 // 1mb

const (
	ErrorReply  = '-'
	StatusReply = '+'
	IntReply    = ':'
	StringReply = '$'
	ArrayReply  = '*'
)

var Nil = errors.New("redis: nil")

type MultiBulkParse func(*Reader, int64) (interface{}, error)

type Reader struct {
	src model.Reader
	buf []byte
}

func NewReader(rd model.Reader) *Reader {
	return &Reader{
		src: rd,
		buf: make([]byte, 4096),
	}
}

func (p *Reader) ReadN(n int) ([]byte, error) {
	return p.src.ReadN(n)
}

func (p *Reader) ReadLine() ([]byte, error) {
	line, err := p.src.ReadLine()
	if err != nil {
		return nil, err
	}
	if len(line) == 0 {
		return nil, errors.New("redis: reply is empty")
	}
	if isNilReply(line) {
		return nil, Nil
	}
	return line, nil
}

func (p *Reader) ReadReply(m MultiBulkParse) (interface{}, error) {
	line, err := p.ReadLine()
	if err != nil {
		return nil, err
	}

	switch line[0] {
	case ErrorReply:
		return nil, ParseErrorReply(line)
	case StatusReply:
		return parseStatusValue(line), nil
	case IntReply:
		return parseInt(line[1:], 10, 64)
	case StringReply:
		return p.readTmpBytesValue(line)
	case ArrayReply:
		n, err := parseArrayLen(line)
		if err != nil {
			return nil, err
		}
		return m(p, n)
	}
	return nil, fmt.Errorf("redis: can't parse %.100q", line)
}

func (p *Reader) ReadIntReply() (int64, error) {
	line, err := p.ReadLine()
	if err != nil {
		return 0, err
	}
	switch line[0] {
	case ErrorReply:
		return 0, ParseErrorReply(line)
	case IntReply:
		return parseInt(line[1:], 10, 64)
	default:
		return 0, fmt.Errorf("redis: can't parse int reply: %.100q", line)
	}
}

func (p *Reader) ReadTmpBytesReply() ([]byte, error) {
	line, err := p.ReadLine()
	if err != nil {
		return nil, err
	}
	switch line[0] {
	case ErrorReply:
		return nil, ParseErrorReply(line)
	case StringReply:
		return p.readTmpBytesValue(line)
	case StatusReply:
		return parseStatusValue(line), nil
	default:
		return nil, fmt.Errorf("redis: can't parse string reply: %.100q", line)
	}
}

func (r *Reader) ReadBytesReply() ([]byte, error) {
	b, err := r.ReadTmpBytesReply()
	if err != nil {
		return nil, err
	}
	cp := make([]byte, len(b))
	copy(cp, b)
	return cp, nil
}

func (p *Reader) ReadStringReply() (string, error) {
	b, err := p.ReadTmpBytesReply()
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (p *Reader) ReadFloatReply() (float64, error) {
	b, err := p.ReadTmpBytesReply()
	if err != nil {
		return 0, err
	}
	return parseFloat(b, 64)
}

func (p *Reader) ReadArrayReply(m MultiBulkParse) (interface{}, error) {
	line, err := p.ReadLine()
	if err != nil {
		return nil, err
	}
	switch line[0] {
	case ErrorReply:
		return nil, ParseErrorReply(line)
	case ArrayReply:
		n, err := parseArrayLen(line)
		if err != nil {
			return nil, err
		}
		return m(p, n)
	default:
		return nil, fmt.Errorf("redis: can't parse array reply: %.100q", line)
	}
}

func (p *Reader) ReadArrayLen() (int64, error) {
	line, err := p.ReadLine()
	if err != nil {
		return 0, err
	}
	switch line[0] {
	case ErrorReply:
		return 0, ParseErrorReply(line)
	case ArrayReply:
		return parseArrayLen(line)
	default:
		return 0, fmt.Errorf("redis: can't parse array reply: %.100q", line)
	}
}

func (p *Reader) ReadScanReply() ([]string, uint64, error) {
	n, err := p.ReadArrayLen()
	if err != nil {
		return nil, 0, err
	}
	if n != 2 {
		return nil, 0, fmt.Errorf("redis: got %d elements in scan reply, expected 2", n)
	}

	cursor, err := p.ReadUint()
	if err != nil {
		return nil, 0, err
	}

	n, err = p.ReadArrayLen()
	if err != nil {
		return nil, 0, err
	}

	keys := make([]string, n)
	for i := int64(0); i < n; i++ {
		key, err := p.ReadStringReply()
		if err != nil {
			return nil, 0, err
		}
		keys[i] = key
	}

	return keys, cursor, err
}

func (p *Reader) readTmpBytesValue(line []byte) ([]byte, error) {
	if isNilReply(line) {
		return nil, Nil
	}

	replyLen, err := strconv.Atoi(bytesToString(line[1:]))
	if err != nil {
		return nil, err
	}

	b, err := p.ReadN(replyLen + 2)
	if err != nil {
		return nil, err
	}
	return b[:replyLen], nil
}

func (r *Reader) ReadInt() (int64, error) {
	b, err := r.ReadTmpBytesReply()
	if err != nil {
		return 0, err
	}
	return parseInt(b, 10, 64)
}

func (r *Reader) ReadUint() (uint64, error) {
	b, err := r.ReadTmpBytesReply()
	if err != nil {
		return 0, err
	}
	return parseUint(b, 10, 64)
}

// --------------------------------------------------------------------

func formatInt(n int64) string {
	return strconv.FormatInt(n, 10)
}

func formatUint(u uint64) string {
	return strconv.FormatUint(u, 10)
}

func formatFloat(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func isNilReply(b []byte) bool {
	return len(b) == 3 &&
		(b[0] == StringReply || b[0] == ArrayReply) &&
		b[1] == '-' && b[2] == '1'
}

func ParseErrorReply(line []byte) error {
	return errors.New(string(line[1:]))
}

func parseStatusValue(line []byte) []byte {
	return line[1:]
}

func parseArrayLen(line []byte) (int64, error) {
	if isNilReply(line) {
		return 0, Nil
	}
	return parseInt(line[1:], 10, 64)
}

func atoi(b []byte) (int, error) {
	return strconv.Atoi(bytesToString(b))
}

func parseInt(b []byte, base int, bitSize int) (int64, error) {
	return strconv.ParseInt(bytesToString(b), base, bitSize)
}

func parseUint(b []byte, base int, bitSize int) (uint64, error) {
	return strconv.ParseUint(bytesToString(b), base, bitSize)
}

func parseFloat(b []byte, bitSize int) (float64, error) {
	return strconv.ParseFloat(bytesToString(b), bitSize)
}

func bytesToString(b []byte) string {
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	strHeader := reflect.StringHeader{
		Data: bytesHeader.Data,
		Len:  bytesHeader.Len,
	}
	return *(*string)(unsafe.Pointer(&strHeader))
}
