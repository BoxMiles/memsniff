package capture

import (
	"errors"
	"github.com/google/gopacket"
)

var (
	// ErrBytesFull is returned if the buffer cannot hold enough
	// data bytes to store this block.
	ErrBytesFull = errors.New("buffer out of space for more bytes")
	// ErrBlocksFull is returned if the buffer cannot hold another block.
	ErrBlocksFull = errors.New("buffer out of space for more packets")
)

// BlockBuffer stores a sequence of byte slices in compact form.
type BlockBuffer struct {
	// stores where the nth block ends in data
	offsets []int
	// block n is located at data[offsets[n-1]:offsets[n]]
	data []byte
}

// NewBlockBuffer creates a new BlockBuffer that can hold a maximum of
// maxBlocks slices or a total of at most maxBytes.
func NewBlockBuffer(maxBlocks, maxBytes int) BlockBuffer {
	return BlockBuffer{
		offsets: make([]int, 0, maxBlocks),
		data:    make([]byte, 0, maxBytes),
	}
}

// BytesRemaining returns the number of additional bytes this buffer can hold.
func (b *BlockBuffer) BytesRemaining() int {
	return cap(b.data) - len(b.data)
}

// Block returns a slice referencing block n, indexed from 0.
// Any modifications to the returned slice modify the contents of the buffer.
func (b *BlockBuffer) Block(n int) []byte {
	start := 0
	if n > 0 {
		start = b.offsets[n-1]
	}
	end := b.offsets[n]
	return b.data[start:end:end]
}

// Cap returns the maximum number of blocks (not bytes) this buffer can hold.
func (b *BlockBuffer) Cap() int {
	return cap(b.offsets)
}

// Len returns the number of blocks currently stored in the buffer.
func (b *BlockBuffer) Len() int {
	return len(b.offsets)
}

// Clear removes all blocks.
func (b *BlockBuffer) Clear() {
	b.offsets = b.offsets[:0]
	b.data = b.data[:0]
}

// Append adds data to the buffer as a new block.
// Makes a copy of data, which may be modified freely after Append returns.
// Returns ErrBytesFull or ErrBlocksFull if there is insufficient space.
func (b *BlockBuffer) Append(data []byte) error {
	if cap(b.offsets) == 0 {
		panic("attempted to Append to uninitialized BlockBuffer")
	}
	if len(b.offsets) >= cap(b.offsets) {
		return ErrBlocksFull
	}
	if len(b.data)+len(data) > cap(b.data) {
		return ErrBytesFull
	}

	b.data = append(b.data, data...)
	b.offsets = append(b.offsets, len(b.data))

	return nil
}

// PacketBuffer stores captured packets in a compact format.
type PacketBuffer struct {
	BlockBuffer
	cis []gopacket.CaptureInfo
}

// NewPacketBuffer creates a PacketBuffer with the specified limits.
func NewPacketBuffer(maxPackets, maxBytes int) *PacketBuffer {
	return &PacketBuffer{
		BlockBuffer: NewBlockBuffer(maxPackets, maxBytes),
		cis:         make([]gopacket.CaptureInfo, maxPackets),
	}
}

// Append adds a packet to the PacketBuffer.
// Makes a copy of pd.Data, which may be modified after Append returns.
// Returns ErrBytesFull or ErrBlocksFull if there is insufficient space.
func (pb *PacketBuffer) Append(pd PacketData) error {
	err := pb.BlockBuffer.Append(pd.Data)
	if err != nil {
		return err
	}
	pb.cis[pb.Len()-1] = pd.Info
	return nil
}

// Packet returns the packet at the specified index, which must be
// less than Len.  The Data field of the returned PacketData
// points into this PacketBuffer and must not be modified.
func (pb *PacketBuffer) Packet(n int) PacketData {
	return PacketData{
		Info: pb.cis[n],
		Data: pb.BlockBuffer.Block(n),
	}
}

// Clear removes all packets.
func (pb *PacketBuffer) Clear() {
	pb.BlockBuffer.Clear()
}
