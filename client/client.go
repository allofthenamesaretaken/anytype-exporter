package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	SpacesResponse | SpaceResponse | ObjectsResponse | ObjectResponse | PropertiesResponse | PropertyResponse | TagsResponse | TagResponse
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

func (client *AnytypeClient) RequestProperties(spaceId string, queryParams *QueryParams) (*PropertiesResponse, error) {
	request, err := client.request.GetProperties(spaceId, queryParams)
	return JSONRequest[PropertiesResponse](client, request, err)
}

func (client *AnytypeClient) RequestProperty(spaceId string, propertyId string) (*PropertyResponse, error) {
	request, err := client.request.GetProperty(spaceId, propertyId)
	return JSONRequest[PropertyResponse](client, request, err)
}

func (client *AnytypeClient) RequestTags(spaceId string, propertyId string) (*TagsResponse, error) {
	request, err := client.request.GetTags(spaceId, propertyId)
	return JSONRequest[TagsResponse](client, request, err)
}

func (client *AnytypeClient) RequestTag(spaceId string, propertyId string, tagId string) (*TagResponse, error) {
	request, err := client.request.GetTag(spaceId, propertyId, tagId)
	return JSONRequest[TagResponse](client, request, err)
}

func (client *AnytypeClient) objectIsPrivate(object Object) (bool, error) {
	var tagPropertyId string
	for _, property := range object.Properties {
		if property.Key == "tag" {
			tagPropertyId = property.ID
		}
	}

	if tagPropertyId == "" {
		return false, nil
	}

	tagsResponse, err := client.RequestTags(object.SpaceID, tagPropertyId)
	if err != nil {
		return false, err
	}

	for _, tag := range tagsResponse.Data {
		if strings.ToLower(tag.Name) == "private" {
			return true, nil
		}
	}

	return false, nil
}

func (client *AnytypeClient) objectIsEmpty(object Object) bool {
	client.logger.Warn("objectIsEmpty is an experimental utility that only checks for empty name fields")
	trimmedName := strings.Trim(object.Name, " ")
	return trimmedName == ""
}

func (client *AnytypeClient) filterPrivateObjects(objects []Object) []Object {
	var filteredObjects []Object
	for _, object := range objects {
		isPrivate, err := client.objectIsPrivate(object)
		if err != nil {
			client.logger.Error("Non fatal: failed to check if object is private", err)
			continue
		}

		if !isPrivate {
			filteredObjects = append(filteredObjects, object)
		}
	}

	return filteredObjects
}

func (client *AnytypeClient) filterEmptyObjects(objects []Object) []Object {
	var filteredObjects []Object
	for _, object := range objects {
		isEmpty := client.objectIsEmpty(object)

		if !isEmpty {
			filteredObjects = append(filteredObjects, object)
		}
	}

	return filteredObjects
}

func (client *AnytypeClient) getTargetSpaceId() (string, error) {
	params := NewQueryParams().WithOffset(0).WithLimit(10)
	spacesResponse, err := client.RequestSpaces(params)
	if err != nil {
		return "", err
	}

	var targetSpaceId string
	for _, v := range spacesResponse.Data {
		if v.Name == client.target {
			targetSpaceId = v.ID
			return targetSpaceId, nil
		}
	}

	msg := fmt.Sprintf("Target space %s does not exist", client.target)
	err = errors.New(msg)
	client.logger.Error(msg, err)
	return "", err
}

func (client *AnytypeClient) getTargetSpaceObjects() ([]Object, error) {
	targetSpaceId, err := client.getTargetSpaceId()
	if err != nil {
		return nil, err
	}

	client.logger.Warn("RequestObjects has been limited to requesting 10 objects")
	client.logger.Warn("Pagination surfing has not been implemented yet")
	params := NewQueryParams().WithOffset(0).WithLimit(10)
	objectsResponse, err := client.RequestObjects(targetSpaceId, params)
	if err != nil {
		return nil, err
	}

	if objectsResponse.Pagination.HasMore {
		client.logger.Error("Non Fatal: The number of pages has exceeded single pagination limits", nil)
	}

	var objects []Object
	for _, object := range objectsResponse.Data {
		objectResponse, err := client.RequestObject(targetSpaceId, object.ID)
		if err != nil {
			return nil, err
		}

		objects = append(objects, objectResponse.Object)
	}

	return objects, nil
}

func (client *AnytypeClient) ExportTargetObjects() error {
	objects, err := client.getTargetSpaceObjects()
	if err != nil {
		return err
	}

	filteredObjects := client.filterEmptyObjects(objects)
	filteredObjects = client.filterPrivateObjects(filteredObjects)
	err = client.exporter.ExportObjects(filteredObjects)
	if err != nil {
		return err
	}
	return nil
}
