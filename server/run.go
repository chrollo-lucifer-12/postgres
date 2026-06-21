package server

import (
	"golang.org/x/sys/unix"
)

func RunServer() {

	pm := NewPostMaster()

	events := make([]unix.EpollEvent, 128)

	for {
		n, err := unix.EpollWait(pm.epollFd, events, -1)

		if err != nil {
			if err == unix.EINTR {
				continue
			}
			panic(err)
		}

		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)

			if fd == pm.listenFd {
				pm.accept()
			}
		}
	}

}
