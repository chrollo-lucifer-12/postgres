package main

import (
	"github.com/postgres/core"
	"github.com/postgres/server"
)

func main() {
	core.Init()
	core.Put("k", "v")
	server.RunServer()
}
