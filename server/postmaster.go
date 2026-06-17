package server

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func RunServer() {

	listenFd, err := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if err != nil {
		panic(err)
	}

	if err := unix.SetsockoptInt(listenFd, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		panic(err)
	}

	if err := unix.SetNonblock(listenFd, true); err != nil {
		panic(err)
	}

	addr := &unix.SockaddrInet4{
		Port: 5432,
		Addr: [4]byte{0, 0, 0, 0},
	}

	if err := unix.Bind(listenFd, addr); err != nil {
		panic(err)
	}

	if err := unix.Listen(listenFd, 128); err != nil {
		panic(err)
	}

	epollFd, err := unix.EpollCreate1(0)
	if err != nil {
		panic(err)
	}

	event := &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(listenFd),
	}

	if err := unix.EpollCtl(epollFd, unix.EPOLL_CTL_ADD, listenFd, event); err != nil {
		panic(err)
	}

	fmt.Println("Listening on :8080")

	events := make([]unix.EpollEvent, 128)

	for {
		n, err := unix.EpollWait(epollFd, events, -1)

		if err != nil {
			panic(err)
		}

		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)

			if fd == listenFd {
				for {
					connFd, _, err := unix.Accept(listenFd)

					if err != nil {
						if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
							break
						}
						panic(err)
					}

					unix.SetNonblock(connFd, true)

					unix.Close(connFd)
				}
			}
		}
	}

}
