package client

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/allofthenamesaretaken/anytype-exporter/utils"
)

/*
 NOTE: QueryParams

QueryParams provides a small builder-style helper for constructing optional
query parameters used by paginated Anytype list endpoints.

Currently supported parameters:

  - offset → the number of items to skip before returning results
  - limit  → the maximum number of items to return

These are stored as pointers so that the exporter can distinguish between:

  - parameters intentionally set to zero
  - parameters omitted entirely

Only parameters that are explicitly provided are serialized into the request
URL.

This avoids sending unnecessary query parameters and keeps requests aligned
with the Anytype API specification.
*/

type QueryParams struct {
	offset *uint
	limit  *uint
}

/*
 NOTE: AnytypeRequest client

AnytypeRequest acts as a lightweight request builder for interacting with the
Anytype HTTP API.

The struct encapsulates:

  - base URL of the Anytype API
  - API authentication token
  - API version header
  - a logger used for structured error reporting

Configuration is sourced from environment variables:

  ANYTYPE_BASE_URL   → base API endpoint
  ANYTYPE_API_KEY    → bearer token used for authentication
  ANYTYPE_VERSION    → API version header required by Anytype

This client intentionally focuses only on request construction. It does not
execute HTTP requests or perform response decoding. Those responsibilities are
handled in client/client.go.

The design keeps this layer simple and focused on building correctly formatted
requests for the endpoints used by the exporter.
*/

type AnytypeRequest struct {
	logger         *utils.Logger
	anytypeBaseUrl string
	anytypeApiKey  string
	anytypeVersion string
}

/*
 NOTE: QueryParams constructor

Creates an empty QueryParams instance used to build optional query parameters
for paginated endpoints.

The idiomatic way to assign values to the respective fields is with the use of
the WithOffset and WithLimit methods. The purpose of this is to improve the
ergonomics of setting *uint type values without having to pass pointers.
*/

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

/*
 NOTE: QueryParams builder methods

WithOffset and WithLimit implement a small fluent-style builder pattern.

Example usage:

  params := NewQueryParams().
      WithOffset(0).
      WithLimit(100)

This pattern improves readability when constructing optional query parameters
and keeps pagination logic concise when iterating through API responses.
*/

func (queryParams *QueryParams) WithOffset(offset uint) *QueryParams {
	queryParams.offset = &offset
	return queryParams
}

func (queryParams *QueryParams) WithLimit(limit uint) *QueryParams {
	queryParams.limit = &limit
	return queryParams
}

/*
 NOTE: HTTP request construction

The internal get() helper builds a fully configured HTTP request for the
Anytype API.

Responsibilities:

  - construct the endpoint URL
  - append pagination query parameters if provided
  - create the HTTP request object
  - attach required API headers

Headers applied:

  Authorization     → Bearer API key
  Accept            → application/json
  Anytype-Version   → requested API version

Centralizing request construction ensures that all endpoints use the same
authentication and header configuration while keeping the endpoint-specific
methods small and readable.
*/

func (clientRequest *AnytypeRequest) get(path string, queryParams *QueryParams) (*http.Request, error) {
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

	request, err := http.NewRequest("GET", requestUrl.String(), nil)
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

/*
 NOTE: List spaces endpoint

GetSpaces constructs a request for the Anytype endpoint:

  GET /v1/spaces

This endpoint returns the list of spaces accessible to the authenticated user.
Spaces act as the top-level containers for content objects.

The response is decoded into SpacesResponse and then used to discover which
spaces should be traversed when collecting objects for export.
*/

func (clientRequest *AnytypeRequest) GetSpaces(queryParams *QueryParams) (*http.Request, error) {
	return clientRequest.get("/v1/spaces", queryParams)
}

func (clientRequest *AnytypeRequest) GetSpace(spaceId string) (*http.Request, error) {
	path := fmt.Sprintf("/v1/spaces/%s", spaceId)
	return clientRequest.get(path, nil)
}

/*
 NOTE: List objects endpoint

GetObjects constructs a request for the Anytype endpoint:

  GET /v1/spaces/{space_id}/objects

This endpoint returns the objects contained within a specific space.

Objects represent the primary content units in Anytype and may include notes,
pages, bookmarks, and other document-like entities.

For this exporter, these objects are the main source of text content used for
LLM ingestion. Their markdown bodies, snippets, and names are extracted and
processed downstream.
*/

func (clientRequest *AnytypeRequest) GetObjects(spaceId string, queryParams *QueryParams) (*http.Request, error) {
	path := fmt.Sprintf("/v1/spaces/%s/objects", spaceId)
	return clientRequest.get(path, queryParams)
}

func (clientRequest *AnytypeRequest) GetObject(spaceId string, objectId string) (*http.Request, error) {
	path := fmt.Sprintf("/v1/spaces/%s/objects/%s", spaceId, objectId)
	return clientRequest.get(path, nil)
}
