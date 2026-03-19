package main

import (
	"log"
	"os"

	"github.com/allofthenamesaretaken/anytype-exporter/client"
	"github.com/allofthenamesaretaken/anytype-exporter/utils"
	"github.com/joho/godotenv"
)

func initLogger() *utils.Logger {
	logInstance, err := utils.NewLogger("messages.log")
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}

	return logInstance
}

func main() {
	logInstance := initLogger()
	defer logInstance.Close()

	err := godotenv.Load()
	if err != nil {
		logInstance.Error("Failed to load .env file", err)
		os.Exit(1)
	}

	anytypeClient := client.NewAnytypeClient(logInstance)

	err = anytypeClient.ExportTargetObjects()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
