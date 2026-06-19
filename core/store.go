package core

import (
	"github.com/postgres/wal"
)

const (
	RMGRKV = 1

	OpPut = "PUT"
	OpDel = "DEL"
)

var Store map[string]string
var w *wal.WAL

func Init() {
	Store = make(map[string]string)
	w = wal.NewWAL()
}

func Get(key string) string {
	val, exists := Store[key]
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

	Store[key] = value
}

func Del(key string) {
	record := []byte("DEL|" + key)

	w.Append(
		1,
		RMGRKV,
		record,
	)

	delete(Store, key)
}
