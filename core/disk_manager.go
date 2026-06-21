package core

import (
	"io"
	"os"
)

type DiskManager struct {
	file *os.File
}

func NewDiskManager(path string) (*DiskManager, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0644)

	if err != nil {
		return nil, err
	}

	return &DiskManager{
		file: f,
	}, nil
}

func (d *DiskManager) ReadPage(pageID int, data []byte) error {

	offset := int64(pageID * PageSize)

	_, err := d.file.ReadAt(data, offset)

	if err == io.EOF {
		return nil
	}

	return err
}

func (d *DiskManager) WritePage(
	pageID int,
	data []byte,
) error {

	offset := int64(pageID * PageSize)

	_, err := d.file.WriteAt(data, offset)

	return err
}
