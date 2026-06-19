package wal

import (
	"encoding/binary"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
)

type WALFile struct {
	dir string
}

func NewWALFile() *WALFile {
	return &WALFile{
		dir: "pg_temp/pgdata/pg_wal",
	}
}

func ReadAll(dir string, apply func(WALEntry)) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	var walFiles []string

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		walFiles = append(walFiles, file.Name())
	}

	sort.Strings(walFiles)

	for _, filename := range walFiles {
		path := filepath.Join(dir, filename)

		f, err := os.Open(path)
		if err != nil {
			return err
		}

		for {
			var header WALHeader

			err := binary.Read(
				f,
				binary.LittleEndian,
				&header,
			)

			if err == io.EOF {
				break
			}

			if err != nil {
				f.Close()
				return err
			}

			data := make([]byte, header.Length)

			_, err = io.ReadFull(f, data)
			if err != nil {
				f.Close()
				return err
			}

			entry := WALEntry{
				Header: header,
				Data:   data,
			}

			apply(entry)
		}

		f.Close()
	}

	return nil
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
