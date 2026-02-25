package main

import (
	"log"

	"github.com/allofthenamesaretaken/anytype-exporter/utils"
)

func initLogger() *utils.Logger {
	logInstance, err := utils.NewLogger("messages.log")
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	return logInstance
}

func main() {
	// inits
	logInstance := initLogger()
	defer logInstance.Close()

	print("hello world")
}
