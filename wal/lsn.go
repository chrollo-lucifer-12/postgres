package wal

import "fmt"

type LSN uint64

const WALSegmentSize = 16 * 1024 * 1024

func (l LSN) String() string {
	hi := uint32(l >> 32)
	lo := uint32(l & 0xffffffff)

	return fmt.Sprintf("%X/%X", hi, lo)
}

func MakeLSN(hi, lo uint32) LSN {
	return LSN(uint64(hi)<<32 | uint64(lo))
}

func (l LSN) SegmentOffset() uint64 {
	return uint64(l) % WALSegmentSize
}

func (l LSN) SegmentNo() uint64 {
	return uint64(l) / WALSegmentSize
}

func (l LSN) WALFileName() string {
	segment := l.SegmentNo()

	return fmt.Sprintf(
		"%08X%016X",
		1,
		segment,
	)
}

func (l LSN) Location() (string, uint64) {
	return l.WALFileName(), l.SegmentOffset()
}
