package core

type EventMultiPlexer interface {
	Add(fd int) error
	Remove(fd int) error
	Wait() ([]Event, error)
	Close() error
}

type Event struct {
	Fd int
}
