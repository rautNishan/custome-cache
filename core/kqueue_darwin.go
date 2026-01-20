package core

import "syscall"

type KqueueMultiPlexing struct {
	kq     int
	events []syscall.Kevent_t
}

func MultiPlexerInit(maxEvents int) (EventMultiPlexer, error) {
	kq, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	return &KqueueMultiPlexing{
		kq:     kq,
		events: make([]syscall.Kevent_t, maxEvents),
	}, nil
}

func (mp *KqueueMultiPlexing) Add(fd int) error {
	ev := syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_ADD,
	}
	_, err := syscall.Kevent(mp.kq, []syscall.Kevent_t{ev}, nil, nil)
	return err
}

func (mp *KqueueMultiPlexing) Remove(fd int) error {
	ev := syscall.Kevent_t{
		Ident:  uint64(fd),
		Filter: syscall.EVFILT_READ,
		Flags:  syscall.EV_DELETE,
	}
	_, err := syscall.Kevent(mp.kq, []syscall.Kevent_t{ev}, nil, nil)
	return err
}

func (mp *KqueueMultiPlexing) Wait() ([]Event, error) {
	events := make([]syscall.Kevent_t, 1024)
	timeOut := &syscall.Timespec{
		Nsec: 1,
	}
	n, err := syscall.Kevent(mp.kq, nil, events, timeOut)

	if err != nil {
		if err == syscall.EINTR { //Source https://stackoverflow.com/questions/19186711/after-call-epoll-wait-linux-system-is-returning-interrupted-system-call-error
			return nil, nil
		}
		return nil, err
	}

	var out []Event
	for i := 0; i < n; i++ {
		out = append(out, Event{
			Fd: int(events[i].Ident),
		})
	}
	return out, nil
}

func (mp *KqueueMultiPlexing) Close() error {
	return syscall.Close(mp.kq)
}
