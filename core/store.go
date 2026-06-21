package core

import (
	"log"
	"strings"

	"github.com/postgres/wal"
)

const (
	RMGRKV = 1

	OpPut = "PUT"
	OpDel = "DEL"
)

var (
	pages []*Page
	index map[string]RID
	w     *wal.WAL
)

type RID struct {
	PageID int
	SlotID int
}

func Init() {
	w = wal.NewWAL()

	pages = []*Page{
		NewPage(),
	}

	index = make(map[string]RID)
}

func Get(key string) string {
	rid, ok := index[key]

	if !ok {
		return "-1"
	}

	log.Println(rid.PageID, rid.SlotID)

	page := pages[rid.PageID]

	data, err := page.Get(rid.SlotID)
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
		pages[oldRID.PageID].Delete(oldRID.SlotID)
	}

	data := []byte(key + "|" + value)

	pageID := len(pages) - 1
	page := pages[pageID]

	slotID, err := page.Insert(data)

	if err != nil {

		page = NewPage()

		pages = append(pages, page)

		pageID = len(pages) - 1

		slotID, _ = page.Insert(data)
	}

	log.Println(pageID, slotID)

	index[key] = RID{
		PageID: pageID,
		SlotID: slotID,
	}
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

	page := pages[rid.PageID]

	page.Delete(rid.SlotID)

	delete(index, key)
}
