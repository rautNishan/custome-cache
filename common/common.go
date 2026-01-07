package common

import (
	"log"
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
