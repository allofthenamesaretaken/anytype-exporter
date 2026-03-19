package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/allofthenamesaretaken/anytype-exporter/utils"
)

type AnytypeClient struct {
	target   string
	self     *http.Client
	request  *AnytypeRequest
	exporter *AnytypeExporter
	logger   *utils.Logger
}

func NewAnytypeClient(logger *utils.Logger) *AnytypeClient {
	return &AnytypeClient{
		target: os.Getenv("ANYTYPE_TARGET_SPACE"),
		self: &http.Client{
			Timeout: time.Second * 10,
		},
		request:  NewAnytypeRequest(logger),
		exporter: NewAnytypeExporter(logger),
		logger:   logger,
	}
}

type JSONResponse interface {
	SpacesResponse | SpaceResponse | ObjectsResponse | ObjectResponse
}

func JSONRequest[T JSONResponse](client *AnytypeClient, request *http.Request, err error) (*T, error) {
	if err != nil {
		client.logger.Error("Anytype client failed to create new request", err)
		return nil, err
	}

	response, err := client.self.Do(request)
	if err != nil {
		client.logger.Error("Anytype client failed to exec request", err)
		return nil, err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("Anytype client response got unexpected code (%d)", response.StatusCode)

		client.logger.Error(msg, err)
		err = errors.New(msg)

		return nil, err
	}

	var data T
	err = json.NewDecoder(response.Body).Decode(&data)
	if err != nil {
		client.logger.Error("Anytype client failed to decode response into json", err)
		return nil, err
	}

	return &data, nil
}

func (client *AnytypeClient) RequestSpaces(queryParams *QueryParams) (*SpacesResponse, error) {
	request, err := client.request.GetSpaces(queryParams)
	return JSONRequest[SpacesResponse](client, request, err)
}

func (client *AnytypeClient) RequestSpace(spaceId string) (*SpaceResponse, error) {
	request, err := client.request.GetSpace(spaceId)
	return JSONRequest[SpaceResponse](client, request, err)
}

func (client *AnytypeClient) RequestObjects(spaceId string, queryParams *QueryParams) (*ObjectsResponse, error) {
	request, err := client.request.GetObjects(spaceId, queryParams)
	return JSONRequest[ObjectsResponse](client, request, err)
}

func (client *AnytypeClient) RequestObject(spaceId string, objectId string) (*ObjectResponse, error) {
	request, err := client.request.GetObject(spaceId, objectId)
	return JSONRequest[ObjectResponse](client, request, err)
}

func (client *AnytypeClient) getTargetSpaceId() (*string, error) {
	// WARN: pagination surfing has not been implemented yet
	params := NewQueryParams().WithOffset(0).WithLimit(10)
	spacesResponse, err := client.RequestSpaces(params)
	if err != nil {
		return nil, err
	}

	var targetSpaceId string
	for _, v := range spacesResponse.Data {
		if v.Name == client.target {
			targetSpaceId = v.ID
			return &targetSpaceId, nil
		}
	}

	return nil, nil
}

func (client *AnytypeClient) ExportTargetObjects() error {
	var targetSpaceId string
	pTargetSpaceId, err := client.getTargetSpaceId()
	if err != nil {
		return err
	}
	if pTargetSpaceId == nil {
		msg := fmt.Sprintf("Target space %s does not exist", client.target)
		err = errors.New(msg)
		client.logger.Error(msg, err)
		return err
	} else {
		targetSpaceId = *pTargetSpaceId
	}

	// WARN: pagination surfing has not been implemented yet
	params := NewQueryParams().WithOffset(0).WithLimit(10)
	objectsResponse, err := client.RequestObjects(targetSpaceId, params)
	if err != nil {
		return err
	}

	for _, v := range objectsResponse.Data {
		objectResponse, err := client.RequestObject(targetSpaceId, v.ID)
		if err != nil {
			return err
		}

		err = client.exporter.ExportObject(objectResponse.Object)
		if err != nil {
			return err
		}
	}

	return nil
}
