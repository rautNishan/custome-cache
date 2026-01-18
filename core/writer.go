package core

import "syscall"

type Writer interface {
	Write(data []byte) error
}

type SyscallWriter struct {
	fd int
}

func NewSysCallWriter(fd int) *SyscallWriter {
	return &SyscallWriter{fd: fd}
}

func (s *SyscallWriter) Write(data []byte) error {
	_, err := syscall.Write(s.fd, data)
	return err
}
