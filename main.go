package main

import (
	"github.com/postgres/core"
	"github.com/postgres/server"
	"github.com/postgres/wal"
)

func apply(entry wal.WALEntry) {
	cmd := server.ParseCmd(string(entry.Data))
	server.Eval(cmd)
}

func main() {
	core.Init()

	err := wal.Restore(apply)
	if err != nil {
		panic(err)
	}

	server.RunServer()
}
