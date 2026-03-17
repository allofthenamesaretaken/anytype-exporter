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

	// client init
	anytypeClient := client.NewAnytypeClient(logInstance)

	// client request spaces
	var spacesResponse *client.SpacesResponse
	var params = client.NewQueryParams().WithOffset(0).WithLimit(1)
	spacesResponse, err = anytypeClient.RequestSpaces(params)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("spaces: %v\n", *spacesResponse)

	var targetSpaceId string
	targetSpace := os.Getenv("ANYTYPE_TARGET_SPACE")
	targetSpaceExists := false
	for _, v := range spacesResponse.Data {
		if v.Name == targetSpace {
			targetSpaceExists = true
			targetSpaceId = v.ID
		}
	}

	var spaceResponse *client.SpaceResponse
	spaceResponse, err = anytypeClient.RequestSpace(targetSpaceId)
	if err != nil {
		os.Exit(1)
	}

	fmt.Printf("target space: %v\n", spaceResponse)

	var objectsResponse *client.ObjectsResponse
	if targetSpaceExists {
		objectsResponse, err = anytypeClient.RequestObjects(targetSpaceId, nil)
		if err != nil {
			os.Exit(1)
		}
	} else {
		msg := fmt.Sprintf("Target space %s does not exist", targetSpace)
		logInstance.Error(msg, err)
		os.Exit(1)
	}

	var objectIds []string
	for _, v := range objectsResponse.Data {
		objectIds = append(objectIds, v.ID)
	}

	// fmt.Printf("objects: %v\n", objects)
	for _, v := range objectIds {
		object, err := anytypeClient.RequestObject(targetSpaceId, v)
		if err != nil {
			os.Exit(1)
		}

		fmt.Printf("%s\n", object.Object.ID)
	}

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
