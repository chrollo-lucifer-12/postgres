package server

import (
	"fmt"
	"strings"

	"github.com/postgres/core"
	"golang.org/x/sys/unix"
)

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

		msg := b.recvBuf[:idx]

		core.Put("k", "v")

		fmt.Println(strings.Split(string(msg), " "))

		value := core.Get("k")

		_, err := unix.Write(b.fd, []byte(value))
		if err != nil {
			fmt.Printf("[fd=%d] write error: %v\n", b.fd, err)
			return
		}

		b.recvBuf = b.recvBuf[idx+1:]
	}
}
