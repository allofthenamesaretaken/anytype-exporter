package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
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

func (client *AnytypeClient) RequestSpaces(queryParams *QueryParams) (*SpacesResponse, error) {
	request, err := client.request.GetSpaces(queryParams)
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

	var spacesData SpacesResponse
	err = json.NewDecoder(response.Body).Decode(&spacesData)
	if err != nil {
		client.logger.Error("Anytype client failed to decode response into json", err)
		return nil, err
	}

	return &spacesData, nil
}

func (client *AnytypeClient) RequestObjects(spaceId string, queryParams *QueryParams) (any, error) {
	request, err := client.request.GetObjects(spaceId, queryParams)
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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		client.logger.Error("Unable to read response body io stream", err)
		return nil, err
	}

	return body, nil
}
