package client

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/allofthenamesaretaken/anytype-exporter/utils"
)

type QueryParams struct {
	offset *uint
	limit  *uint
}

type AnytypeRequest struct {
	logger         *utils.Logger
	anytypeBaseUrl string
	anytypeApiKey  string
	anytypeVersion string
}

func NewQueryParams() *QueryParams {
	return &QueryParams{}
}

func NewAnytypeRequest(logger *utils.Logger) *AnytypeRequest {
	return &AnytypeRequest{
		logger:         logger,
		anytypeBaseUrl: os.Getenv("ANYTYPE_BASE_URL"),
		anytypeApiKey:  os.Getenv("ANYTYPE_API_KEY"),
		anytypeVersion: os.Getenv("ANYTYPE_VERSION"),
	}
}

func (queryParams *QueryParams) WithOffset(offset uint) *QueryParams {
	queryParams.offset = &offset
	return queryParams
}

func (queryParams *QueryParams) WithLimit(limit uint) *QueryParams {
	queryParams.limit = &limit
	return queryParams
}

func (clientRequest *AnytypeRequest) get(method string, path string, queryParams *QueryParams) (*http.Request, error) {
	rawUrl := clientRequest.anytypeBaseUrl + path

	requestUrl, err := url.Parse(rawUrl)
	if err != nil {
		clientRequest.logger.Error("Failed to parse base url and path into url struct", err)
		return nil, err
	}

	if queryParams != nil {
		queries := url.Values{}
		if queryParams.offset != nil {
			queries.Set("offset", fmt.Sprintf("%d", *queryParams.offset))
		}
		if queryParams.limit != nil {
			queries.Set("limit", fmt.Sprintf("%d", *queryParams.limit))
		}
		requestUrl.RawQuery = queries.Encode()
	}

	request, err := http.NewRequest(method, requestUrl.String(), nil)
	if err != nil {
		clientRequest.logger.Error("Failed to create new request", err)
		return nil, err
	}

	auth := "Bearer " + clientRequest.anytypeApiKey

	request.Header.Set("Accept", "application/json")
	request.Header.Set("Authorization", auth)
	request.Header.Set("Anytype-Version", clientRequest.anytypeVersion)

	return request, nil
}

func (clientRequest *AnytypeRequest) GetSpaces(queryParams *QueryParams) (*http.Request, error) {
	return clientRequest.get("GET", "/v1/spaces", queryParams)
}

func (clientRequest *AnytypeRequest) GetObjects(spaceId string, queryParams *QueryParams) (*http.Request, error) {
	path := fmt.Sprintf("/v1/spaces/%s/objects", spaceId)
	return clientRequest.get("GET", path, queryParams)
}
