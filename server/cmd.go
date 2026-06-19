package server

import (
	"strings"

	"github.com/postgres/core"
)

type Cmd struct {
	Command string
	Args    []string
}

func getCmd(rawMsg []byte) Cmd {
	cmd := Cmd{}

	splitMsg := strings.Split(string(rawMsg), " ")

	cmd.Command = splitMsg[0]
	cmd.Args = splitMsg[1:]

	return cmd
}

func eval(cmd Cmd) []byte {
	switch cmd.Command {
	case "SET":
		key := cmd.Args[0]
		val := cmd.Args[1]
		core.Put(key, val)
		return []byte("ok")
	case "GET":
		key := cmd.Args[0]
		val := core.Get(key)
		return []byte(val)
	case "DEL":
		key := cmd.Args[0]
		core.Del(key)
		return []byte("ok")
	default:
		return []byte("-1")
	}
}
