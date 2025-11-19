package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	Host string
	Port int
}

func InitializeConfig() Config {
	config := Config{}

	config.Host = os.Getenv("HOST")
	portStr := os.Getenv("PORT")

	if config.Host == "" {
		log.Println("Initializing default Host: Localhost")
		config.Host = "localhost"
	}

	if portStr == "" {
		log.Println("Initializing default Port: 3769")
		portStr = "3769"
	}
	intPort, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal("Error while initializeing port")
	}
	config.Port = intPort
	return config
}
