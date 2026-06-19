package wal

import (
	"log"
)

const MAX_ENTRIES = 2

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
			Length:   uint32(len(data)),
			Checksum: 0,
			PrevLsn:  w.lastLSN,
		},
		Data: data,
	}

	w.entries = append(w.entries, entry)

	w.lastLSN = recordLSN
	w.lsn += LSN(entrySize)

	if len(w.entries) == MAX_ENTRIES {
		log.Println("triggering a disc save")
		w.Flush()
	}
}

func (w *WAL) Flush() {
	wf := NewWALFile()

	for _, entry := range w.entries {
		err := wf.Write(entry)
		if err != nil {
			log.Println(err)
		}
	}

	SaveControlFile("pg_temp/pg_control.json", ControlFile{
		LastLSN: w.lastLSN,
	})

	w.entries = nil
	w.lsn = 0
	w.lastLSN = 0
}

func Restore(apply func(WALEntry)) error {
	err := ReadAll("pg_temp/pgdata/pg_wal", apply)

	if err != nil {
		return err
	}

	return nil
}
