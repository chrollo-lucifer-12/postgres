package wal

import (
	"encoding/binary"
	"os"
	"path"
)

type WALFile struct {
	dir string
}

func NewWALFile() *WALFile {
	return &WALFile{
		dir: "pgdata/pg_wal",
	}
}

func (wf *WALFile) Write(entry WALEntry) error {
	filename := entry.Header.LSN.WALFileName()

	path := path.Join(wf.dir, filename)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	offset := entry.Header.LSN.SegmentOffset()

	if _, err := f.Seek(int64(offset), 0); err != nil {
		return err
	}

	if err := binary.Write(
		f,
		binary.LittleEndian,
		entry.Header,
	); err != nil {
		return err
	}

	if _, err := f.Write(entry.Data); err != nil {
		return err
	}

	return f.Sync()
}
