package core

import (
	"log"
	"runtime"
)

type OS int

const (
	UnknownOS OS = iota
	Linux
	Windows
	MacOS
)

func PanicOnErr(msg string, err error) {
	if err != nil {
		log.Printf("%s: %v", msg, err)
		panic(err)
	}
}

func DetectOS() OS {
	switch runtime.GOOS {
	case "linux":
		return Linux //1 for linux
	case "windows":
		return Windows //2 for windows
	case "darwin":
		return MacOS //3 for mac
	default:
		return UnknownOS
	}
}
