package core

import (
	"log"
	"strings"
)

const (
	RMGRKV = 1

	OpPut = "PUT"
	OpDel = "DEL"
)

var (
	bpm *BufferPool
)

type RID struct {
	PageID int
	SlotID int
}

func Init() {
	//	w = wal.NewWAL()

	disk, err := NewDiskManager("pg_temp/data.db")
	if err != nil {
		panic(err)
	}

	bpm, err = NewBufferPool(1024, disk)
	if err != nil {
		panic(err)
	}

	//	index = make(map[string]RID)

	// nextPageID = 0
}

func Get(key string) string {
	rid, ok := bpm.IndexGet(key)

	if !ok {
		return "-1"
	}

	frame, err := bpm.FetchPage(rid.PageID)
	if err != nil {
		log.Println("no frame")
		return "-1"
	}

	data, err := frame.Page.Get(rid.SlotID)

	bpm.UnpinPage(rid.PageID, false)

	if err != nil {
		log.Println("error reading data")
		return "-1"
	}

	parts := strings.SplitN(string(data), "|", 2)

	if len(parts) != 2 {
		return "-1"
	}

	return parts[1]

}

func Put(key, value string) {

	//	record := []byte("PUT|" + key + "|" + value)

	// w.Append(
	// 	1,
	// 	RMGRKV,
	// 	record,
	// )

	if oldRID, ok := bpm.IndexGet(key); ok {

		frame, err := bpm.FetchPage(oldRID.PageID)

		if err == nil {
			log.Println("page already existing :", err)

			frame.Page.Delete(oldRID.SlotID)

			bpm.UnpinPage(
				oldRID.PageID,
				true,
			)
		}
	}

	data := []byte(key + "|" + value)

	pageID := bpm.NewPageID()

	frame, err := bpm.FetchPage(pageID)

	if err != nil {
		log.Println("error when fetching page :", err)
		return
	}

	slotID, err := frame.Page.Insert(data)

	if err != nil {

		log.Println(err)

		bpm.UnpinPage(pageID, false)

		pageID = bpm.NewPageID()

		frame, err = bpm.FetchPage(pageID)

		if err != nil {

			log.Println(err)
			return
		}

		slotID, err = frame.Page.Insert(data)

		if err != nil {
			log.Panicln(err)
			return
		}
	}

	bpm.IndexSet(key, RID{PageID: pageID, SlotID: slotID})

	bpm.UnpinPage(pageID, true)
}

func Del(key string) {

	//	record := []byte("DEL|" + key)

	// w.Append(
	// 	1,
	// 	RMGRKV,
	// 	record,
	// )

	rid, ok := bpm.IndexGet(key)

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

	bpm.IndexDelete(key)
}
