package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Host      string
	Port      int
	MaxEvents int
}

func InitializeConfig() Config {
	config := Config{}

	config.Host = os.Getenv("HOST")
	portStr := os.Getenv("PORT")

	if config.Host == "" {
		config.Host = "localhost"
	}

	if portStr == "" {
		portStr = "3000"
	}
	intPort, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("Error while initializeing port")
	}
	config.Port = intPort

	if config.MaxEvents == 0 {
		config.MaxEvents = 1000
	}

	return config
}
