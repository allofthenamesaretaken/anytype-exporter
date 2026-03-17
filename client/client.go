package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/allofthenamesaretaken/anytype-exporter/utils"
)

type AnytypeClient struct {
	self    *http.Client
	request *AnytypeRequest
	logger  *utils.Logger
}

func NewAnytypeClient(logger *utils.Logger) *AnytypeClient {
	return &AnytypeClient{
		self: &http.Client{
			Timeout: time.Second * 10,
		},
		request: NewAnytypeRequest(logger),
		logger:  logger,
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
