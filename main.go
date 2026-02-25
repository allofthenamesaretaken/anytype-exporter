package main

import (
	"fmt"
	"log"
	"os"

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
	// inits
	logInstance := initLogger()
	defer logInstance.Close()

	// dotenv vars
	err := godotenv.Load()
	if err != nil {
		logInstance.Error("Failed to load .env file", err)
	}

	anytypeBaseUrl := os.Getenv("ANYTYPE_BASE_URL")
	anytypeApiKey := os.Getenv("ANYTYPE_API_KEY")
	anytypeVersion := os.Getenv("ANYTYPE_VERSION")
	exportDir := os.Getenv("EXPORT_DIR")
	fmt.Printf("%v\n", anytypeBaseUrl)
	fmt.Printf("%v\n", anytypeApiKey)
	fmt.Printf("%v\n", anytypeVersion)
	fmt.Printf("%v\n", exportDir)
	print("hello world")
}
