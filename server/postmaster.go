package server

import (
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

type PostMaster struct {
	listenFd int
	epollFd  int
}

func NewPostMaster() *PostMaster {
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

	fmt.Println("Listening on :5432")

	return &PostMaster{
		listenFd: listenFd,
		epollFd:  epollFd,
	}
}

func (p *PostMaster) accept() {
	for {
		connFd, _, err := unix.Accept(p.listenFd)

		if err != nil {
			if err == unix.EAGAIN || err == unix.EWOULDBLOCK {
				break
			}
			panic(err)
		}

		p.spawnProcess(connFd)
	}
}

func (p *PostMaster) spawnProcess(connFd int) {
	fmt.Println("before fork", os.Getpid())

	r1, _, errno := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
	if errno != 0 {
		panic(errno)
	}

	if r1 == 0 {
		fmt.Println("child pid:", os.Getpid())
		unix.Close(p.listenFd)

		unix.SetNonblock(connFd, false)
		handleClient(connFd)

		unix.Close(connFd)
		syscall.Exit(0)
	} else {
		fmt.Println("parent pid:", os.Getpid(), "child pid:", r1)
	}

	unix.Close(connFd)
}
