package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func encodeCommand(parts []string) []byte {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("*%d\r\n", len(parts)))
	for _, p := range parts {
		sb.WriteString(fmt.Sprintf("$%d\r\n%s\r\n", len(p), p))
	}
	return []byte(sb.String())
}

func main() {
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connected to server")

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Split(line, " ")
		payload := encodeCommand(parts)

		_, err = conn.Write(payload)
		if err != nil {
			fmt.Println("write error:", err)
			return
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read error:", err)
			return
		}

		fmt.Println(string(buf[:n]))
	}
}
