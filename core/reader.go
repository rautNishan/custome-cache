package core

import "syscall"

type Reader interface {
	Read(buf []byte) (int error)
}

type SyscallReader struct {
	fd int
}

func NewSyscallReader(fd int) *SyscallReader {
	return &SyscallReader{fd: fd}
}

func (r *SyscallReader) Read(buf []byte) (int, error) {
	return syscall.Read(r.fd, buf)
}
