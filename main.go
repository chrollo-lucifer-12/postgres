package main

import (
	"github.com/postgres/core"
	"github.com/postgres/server"
	"github.com/postgres/wal"
)

func apply(wal wal.WALEntry) {
	cmd := server.GetCmd(wal.Data)
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
