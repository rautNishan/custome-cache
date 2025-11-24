package core

import (
	"log"
	"net"
	"syscall"

	"github.com/rautNishan/custome-cache/config"
)

func CreateAndHandelConnection(cnfg *config.Config) {
	currentlyUserOs := DetectOS()
	log.Println("Client os: ", currentlyUserOs) //For debugging
	switch currentlyUserOs {
	case 1:
		runForLinux(cnfg)
	}
}

func runForLinux(cnfg *config.Config) {
	max_client := 10000
	// var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_client)
	serverFileDescriptor, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	if err != nil {
		PanicOnErr("failed to create listing file descriptor", err)
	}
	defer syscall.Close(serverFileDescriptor)

	err = syscall.SetNonblock(serverFileDescriptor, true)
	if err != nil {
		PanicOnErr("failed to set non-blocking on listning file descriptor", err)
	}

	ips, err := net.LookupIP(cnfg.Host)
	if err != nil {
		PanicOnErr("DNS lookup failed", err)
	}

	var ipv4 net.IP
	for _, ip := range ips {
		if v4 := ip.To4(); v4 != nil {
			ipv4 = v4
			break
		}
	}

	if ipv4 == nil {
		PanicOnErr("no IPv4 found for host", nil)
	}

	log.Println("IP: ", ipv4)

	err = syscall.Bind(serverFileDescriptor, &syscall.SockaddrInet4{
		Port: cnfg.Port,
		Addr: [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
	})
	if err != nil {
		PanicOnErr("bind failed", err)
	}

	err = syscall.Listen(serverFileDescriptor, max_client)
	if err != nil {
		PanicOnErr("listen failed", err)
	}
	log.Println("Running server on host: ", cnfg.Host, "and port: ", cnfg.Port)

}
