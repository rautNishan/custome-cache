package main

import (
	"github.com/rautNishan/custome-cache/config"
	"github.com/rautNishan/custome-cache/core"
)

func main() {
	startServer()
}

func startServer() {
	config := config.InitializeConfig()
	core.CreateAndHandelConnection(&config)
}
