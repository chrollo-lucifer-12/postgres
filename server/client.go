package server

import (
	"fmt"
	"strings"

	"github.com/postgres/core"
	"golang.org/x/sys/unix"
)

type Cmd struct {
	Command string
	Args    []string
}

type Backend struct {
	fd      int
	recvBuf []byte
}

func handleClient(connFd int) {

	b := &Backend{
		fd:      connFd,
		recvBuf: make([]byte, 0, 8192),
	}

	b.readLoop()
}

func (b *Backend) readLoop() {
	tmp := make([]byte, 4096)

	for {
		n, err := unix.Read(b.fd, tmp)

		if err != nil {
			fmt.Printf("[fd=%d] read error: %v\n", b.fd, err)
			return
		}

		if n == 0 {
			fmt.Printf("[fd=%d] client disconnected\n", b.fd)
			return
		}

		b.recvBuf = append(b.recvBuf, tmp[:n]...)

		b.processMessages()

	}
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

func (b *Backend) processMessages() {

	for {
		idx := -1

		for i, c := range b.recvBuf {
			if c == '\n' {
				idx = i
				break
			}
		}

		if idx == -1 {
			return
		}

		rawMsg := b.recvBuf[:idx]

		parsedCommand := getCmd(rawMsg)

		res := eval(parsedCommand)

		_, err := unix.Write(b.fd, []byte(res))
		if err != nil {
			fmt.Printf("[fd=%d] write error: %v\n", b.fd, err)
			return
		}

		b.recvBuf = b.recvBuf[idx+1:]
	}
}
