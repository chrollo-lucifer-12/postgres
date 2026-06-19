package core

import "github.com/postgres/wal"

const (
	RMGRKV = 1

	OpPut = "PUT"
	OpDel = "DEL"
)

var store map[string]string
var w *wal.WAL

func Init() {
	store = make(map[string]string)
	w = wal.NewWAL()
}

func Get(key string) string {
	val, exists := store[key]
	if !exists {
		return "-1"
	}
	return val
}

func Put(key, value string) {
	record := []byte("PUT|" + key + "|" + value)

	w.Append(
		1,
		RMGRKV,
		record,
	)

	store[key] = value
}

func Del(key string) {
	record := []byte("DEL|" + key)

	w.Append(
		1,
		RMGRKV,
		record,
	)

	delete(store, key)
}
