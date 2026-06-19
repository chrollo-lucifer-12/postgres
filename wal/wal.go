package wal

type WALHeader struct {
	LSN      LSN
	TxID     uint64
	RMGR     uint32
	Length   uint32
	Checksum uint32
	PrevLsn  LSN
}

type WALEntry struct {
	Header WALHeader
	Data   []byte
}

type WAL struct {
	entries []WALEntry
	lsn     LSN
	lastLSN LSN
}

func NewWAL() *WAL {
	return &WAL{
		entries: make([]WALEntry, 0),
		lsn:     0,
		lastLSN: 0,
	}
}

func (w *WAL) Append(txID uint64, rmgr uint32, data []byte) {

	recordLSN := w.lsn

	entrySize := uint64(len(data)) + uint64(40)

	entry := WALEntry{
		Header: WALHeader{
			LSN:      recordLSN,
			TxID:     txID,
			RMGR:     rmgr,
			Length:   uint32(entrySize),
			Checksum: 0,
			PrevLsn:  w.lastLSN,
		},
		Data: data,
	}

	w.entries = append(w.entries, entry)

	w.lastLSN = recordLSN
	w.lsn += LSN(entrySize)
}
