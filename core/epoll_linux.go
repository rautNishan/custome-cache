package core

import (
	"syscall"
)

type EpollEventMultiPlexing struct {
	epollFd int
	events  []syscall.EpollEvent
}

func MultiPlexerInit(maxEvents int) (EventMultiPlexer, error) {
	epFd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, err
	}
	return &EpollEventMultiPlexing{
		epollFd: epFd,
		events:  make([]syscall.EpollEvent, maxEvents),
	}, nil
}

func (mp *EpollEventMultiPlexing) Add(fd int) error {
	ev := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(fd),
	}
	return syscall.EpollCtl(mp.epollFd, syscall.EPOLL_CTL_ADD, fd, &ev)
}

func (mp *EpollEventMultiPlexing) Remove(fd int) error {
	return syscall.EpollCtl(mp.epollFd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (mp *EpollEventMultiPlexing) Wait() ([]Event, error) {
	n, err := syscall.EpollWait(mp.epollFd, mp.events[:], 1)
	if err != nil {
		if err == syscall.EINTR { //Source https://stackoverflow.com/questions/19186711/after-call-epoll-wait-linux-system-is-returning-interrupted-system-call-error
			return nil, nil
		}
		return nil, err
	}
	var out []Event
	for i := 0; i < n; i++ {
		out = append(out, Event{
			Fd: int(mp.events[i].Fd),
		})
	}
	return out, nil
}

func (mp *EpollEventMultiPlexing) Close() error {
	return syscall.Close(mp.epollFd)
}
