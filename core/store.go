package core

import (
	"strings"

	"github.com/postgres/wal"
)

const (
	RMGRKV = 1

	OpPut = "PUT"
	OpDel = "DEL"
)

var (
	index map[string]RID

	w   *wal.WAL
	bpm *BufferPool

	nextPageID int
)

type RID struct {
	PageID int
	SlotID int
}

func Init() {
	w = wal.NewWAL()

	disk, err := NewDiskManager("data.db")
	if err != nil {
		panic(err)
	}

	bpm, err = NewBufferPool(1024, disk)
	if err != nil {
		panic(err)
	}

	index = make(map[string]RID)

	nextPageID = 0
}

func Get(key string) string {
	rid, ok := index[key]

	if !ok {
		return "-1"
	}

	frame, err := bpm.FetchPage(rid.PageID)
	if err != nil {
		return "-1"
	}

	data, err := frame.Page.Get(rid.SlotID)

	bpm.UnpinPage(rid.PageID, false)

	if err != nil {
		return "-1"
	}

	parts := strings.SplitN(string(data), "|", 2)

	if len(parts) != 2 {
		return "-1"
	}

	return parts[1]

}

func Put(key, value string) {

	record := []byte("PUT|" + key + "|" + value)

	w.Append(
		1,
		RMGRKV,
		record,
	)

	if oldRID, ok := index[key]; ok {

		frame, err := bpm.FetchPage(oldRID.PageID)

		if err == nil {
			frame.Page.Delete(oldRID.SlotID)

			bpm.UnpinPage(
				oldRID.PageID,
				true,
			)
		}
	}

	data := []byte(key + "|" + value)

	pageID := nextPageID

	frame, err := bpm.FetchPage(pageID)

	if err != nil {
		return
	}

	slotID, err := frame.Page.Insert(data)

	if err != nil {

		bpm.UnpinPage(pageID, false)

		nextPageID++

		pageID = nextPageID

		frame, err = bpm.FetchPage(pageID)

		if err != nil {
			return
		}

		slotID, err = frame.Page.Insert(data)

		if err != nil {
			return
		}
	}

	index[key] = RID{
		PageID: pageID,
		SlotID: slotID,
	}

	bpm.UnpinPage(pageID, true)
}

func Del(key string) {

	record := []byte("DEL|" + key)

	w.Append(
		1,
		RMGRKV,
		record,
	)

	rid, ok := index[key]

	if !ok {
		return
	}

	frame, err := bpm.FetchPage(rid.PageID)

	if err != nil {
		return
	}

	frame.Page.Delete(rid.SlotID)

	bpm.UnpinPage(
		rid.PageID,
		true,
	)

	delete(index, key)
}
