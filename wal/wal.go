package wal

import "fmt"

type LSN uint64

func (l LSN) String() string {
	hi := uint32(l >> 32)
	lo := uint32(l & 0xffffffff)

	return fmt.Sprintf("%X/%X", hi, lo)
}

func MakeLSN(hi, lo uint32) LSN {
	return LSN(uint64(hi)<<32 | uint64(lo))
}

type WALHeader struct {
	LSN      LSN
	TxID     uint64
	RMGR     uint32
	Length   uint32
	Checksum uint32
	PrevLsn  uint64
}

type WALEntry struct {
	Header WALHeader
	Data   []byte
}

type WAL struct {
	entries []WALEntry
	lsn     LSN
}

func NewWAL() *WAL {
	return &WAL{
		entries: make([]WALEntry, 0),
		lsn:     0,
	}
}

func (w *WAL) Append(txID uint64, rmgr uint32, data []byte, prevLsn uint64) {
	entrySize := uint64(len(data)) + 32

	w.lsn += LSN(entrySize)

	entry := WALEntry{
		Header: WALHeader{},
		Data:   data,
	}

	w.entries = append(w.entries, entry)

}
