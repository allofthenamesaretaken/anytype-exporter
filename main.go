package main

import (
	"fmt"
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
	// inits
	logInstance := initLogger()
	defer logInstance.Close()

	// dotenv vars
	err := godotenv.Load()
	if err != nil {
		logInstance.Error("Failed to load .env file", err)
		os.Exit(1)
	}

	// targetSpace := os.Getenv("ANYTYPE_TARGET_SPACE")

	// client init
	anytypeClient := client.NewAnytypeClient(logInstance)

	// client request spaces
	var spaces *client.SpacesResponse
	var params = client.NewQueryParams().WithOffset(0).WithLimit(1)
	spaces, err = anytypeClient.RequestSpaces(params)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("spaces: %v\n", *spaces)

	// var spaceIds = make(map[string]string)
	// for _, v := range spaces.Objects {
	// 	spaceIds[v.Name] = v.ID
	// }
	//
	// var targetSpaceId string
	// for key := range spaceIds {
	// 	if key == targetSpace {
	// 		targetSpaceId = spaceIds[key]
	// 	}
	// }
	//
	// body, err := anytypeClient.RequestObjects(targetSpaceId)
	//
	// fmt.Printf("%s", body)
}
