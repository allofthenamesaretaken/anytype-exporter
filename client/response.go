package client

/*
 INFO: Exporter scope

This package models only the subset of the Anytype API needed for content
extraction and LLM ingestion.

Primary goal:
Extract human-meaningful document data such as object names, markdown bodies,
snippets, and basic space context.

Non-goals:
Reconstruct full Anytype objects with complete visual, schema, or UI fidelity.

As a result, some fields and types defined by the API are intentionally
omitted or commented out in this file. Those fields are either:

 - presentation metadata (for example icons),
 - schema metadata (for example object type definitions and property models),
 - or additional structured data not currently required for document scraping.

If the exporter is later expanded to preserve richer object fidelity, the
commented types can be restored and completed as needed.
*/

/*
 NOTE: Response wrapper

Anytype endpoints commonly return either:

  - a paginated result set under a top-level "data" field, or
  - a single resource wrapped in a named top-level field such as "space" or
    "object"

For the purposes of this exporter, these wrappers exist only to match the API
response shape and support decoding.

List wrappers are used to iterate through paginated responses while collecting
content-bearing objects for downstream processing.

Single-item wrappers are used to retrieve richer representations of individual
resources. In particular, full object retrieval may expose more complete
content and metadata than list responses, even though this exporter only keeps
the subset it needs for text extraction.

The wrapper fields themselves are operational rather than semantic and should
not be treated as exported document content.
*/

/*
 NOTE: Space list response

Some Anytype endpoints return a paginated list of spaces under a top-level
"data" field together with pagination metadata.

This wrapper is used to enumerate available spaces and support full traversal
of all spaces accessible to the API key.

List responses provide sufficient information for identifying and labeling
spaces, but may omit additional metadata available in single space retrieval.

The wrapper exists only to match the API response shape. The contained Space
objects are used as contextual metadata for associating and organizing
exported content.
*/
type SpacesResponse struct {
	Data       []Space    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

/*
 NOTE: Single space response

Some Anytype endpoints return a single space wrapped in a top-level "space"
field rather than inside a paginated "data" array.

The full space response may include additional workspace metadata beyond the
subset modeled by the Space struct, such as internal workspace IDs used by
Anytype for home, archive, profile, and related views.

This exporter intentionally retains only the space fields useful for content
association and basic context. The wrapper exists only to match the API
response shape.
*/
type SpaceResponse struct {
	Space Space `json:"space"`
}

/*
 NOTE: Object list response

Some Anytype endpoints return a paginated list of objects under a top-level
"data" field together with pagination metadata.

This wrapper is used for space-scoped object discovery and bulk traversal.
List responses are useful for enumerating candidate objects to export, but
they may not include the full richness of a dedicated single-object retrieval.

In particular, the exporter should treat list responses as suitable for
indexing, selection, and lightweight metadata capture, while full object
retrieval remains the preferred source for complete textual extraction when
available.

The wrapper itself exists only to match the API response shape.
*/
type ObjectsResponse struct {
	Data       []Object   `json:"data"`
	Pagination Pagination `json:"pagination"`
}

/*
 NOTE: Single object response

Some Anytype endpoints return a single object wrapped in a top-level "object"
field rather than inside a paginated "data" array.

This wrapper is especially important for full object retrieval, which may
return a richer and more complete representation than list responses.

For this exporter, the single-object response should be treated as the
authoritative shape for extracting document content when full text is needed.
The wrapper itself exists only to match the API response shape.
*/
type ObjectResponse struct {
	Object Object `json:"object"`
}

/*
 NOTE: Pagination

Pagination describes how much data was returned by a list endpoint and
whether additional requests are required to retrieve the remaining items.

This metadata is operational rather than semantic. It is used only for
API traversal and should not be treated as part of the exported document
content.
*/

type Pagination struct {
	Total   int  `json:"total"`
	Offset  int  `json:"offset"`
	Limit   int  `json:"limit"`
	HasMore bool `json:"has_more"`
}

/*
 NOTE: ObjectType

The Anytype API uses the "object" field as a lightweight discriminator for
the kind of record returned by an endpoint.

Examples include:

  - "space"  → a space/container
  - "chat"   → a chat object
  - "object" → a normal content object

For this exporter, ObjectType is used only to distinguish broad API record
categories during decoding. It is not used to preserve Anytype's full schema
model and should not be interpreted as a replacement for the richer "type"
metadata available on objects.
*/

type ObjectType string

const (
	ObjectTypeSpace  ObjectType = "space"
	ObjectTypeChat   ObjectType = "chat"
	ObjectTypeObject ObjectType = "object"
)

/*
 NOTE: Space handling

A Space represents a high-level container or workspace in Anytype.

For this exporter, space data is preserved only insofar as it helps associate
content objects with their originating container. The meaningful fields are:

  - id          → stable identifier for the space
  - name        → human-readable label
  - description → optional descriptive context
  - network_id  → Anytype network identifier
  - gateway_url → base gateway URL used by Anytype to serve media and files
  - object      → broad API discriminator for the returned record kind

The Anytype API may expose additional space metadata on full retrieval,
including workspace-specific IDs and presentation details. This exporter does
not attempt to reconstruct full space configuration or UI state. Space records
are used primarily as contextual metadata for the documents scraped from them.
*/
type Space struct {
	Description string     `json:"description"`
	GatewayUrl  string     `json:"gateway_url"`
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	NetworkID   string     `json:"network_id"`
	Object      ObjectType `json:"object"`

	// Icon     *Icon      `json:"icon"`
}

/*
 NOTE: Object handling

This exporter is centered on content-bearing Anytype objects.

The Anytype API exposes richer object data than this struct models, especially
on full object retrieval. Depending on the endpoint, object responses may also
include icons, type information, blocks, details, and structured properties.

This struct intentionally retains only the subset most useful for document
extraction and LLM ingestion:

  - id        → stable object identifier
  - name      → object title, when present
  - markdown  → textual body content when returned by the API
  - snippet   → short preview text, useful for discovery and fallback indexing
  - space_id  → associates the object with its parent space
  - archived  → allows archived content to be filtered if desired
  - layout    → lightweight context about the object's presentation category
  - object    → broad API discriminator indicating the returned record kind

Important:
List-object responses and single-object responses should not be assumed to
have identical field completeness. In particular, full object retrieval is the
more appropriate source when the exporter needs the most complete available
textual representation.

This exporter intentionally focuses on textual and contextual data rather
than full object fidelity. Schema metadata, icon metadata, blocks/details
beyond the extracted text, and other presentation-oriented fields are not
currently exported.
*/
type Object struct {
	Archived bool       `json:"archived"`
	ID       string     `json:"id"`
	Layout   string     `json:"layout"`
	Markdown string     `json:"markdown"`
	Name     string     `json:"name"`
	Object   ObjectType `json:"object"`
	Snippet  string     `json:"snippet"`
	SpaceID  string     `json:"space_id"`

	// Type       *Type      `json:"type"`
	// Properties []Property `json:"properties"`
	// Icon       *Icon      `json:"icon"`
}

/*
 NOTE: Icon handling

The Anytype API represents object icons using three formats:

  - "emoji" → a literal Unicode emoji
  - "icon"  → a built-in named icon with an optional color
  - "file"  → a file identifier referencing an icon asset

This exporter intentionally does NOT support exporting icons.

Reason:
The API does not provide enough information to reconstruct icon files
for two of the three formats:

  1. emoji
     The API provides only the Unicode emoji character. While this is
     sufficient to render an emoji using a system font or emoji library,
     it does not guarantee reproduction of the exact visual asset used
     by Anytype.


  2. icon
     These are built-in icons referenced only by name and color. The API
     does not expose the underlying SVG or any download URL for the
     built-in icon set. Therefore the actual icon file cannot be exported.

  3. file
     While this format references an icon asset via a file identifier,
     it is not required for reconstructing exported objects and may rely
     on gateway/media serving behavior outside the scope of this exporter.

Because reliable reconstruction of icon assets is not possible using the
data returned by the API alone, icon data is parsed only for schema
compatibility and is otherwise ignored during export.
*/

// type Icon struct {
// 	Emoji  string     `json:"emoji,omitempty"`
// 	Format IconFormat `json:"format"`
// }
//
// type IconFormat string
//
// const (
// 	IconFormatEmoji IconFormat = "emoji"
// 	IconFormatFile  IconFormat = "file"
// 	IconFormatIcon  IconFormat = "icon"
// )

/*
 NOTE: Type handling

Anytype objects may include a nested "type" object describing the schema or
template the object belongs to, such as Page, Note, Bookmark, or Task.

A type record is metadata about classification and UI behavior. It can define:

  - the type's stable id/key,
  - its display name,
  - its layout,
  - and the set of properties typically associated with it.

 WARN: This exporter intentionally does NOT currently support exporting type data.

Reason:
The primary purpose of this program is to extract document content for LLM
consumption. Type metadata is useful for reproducing Anytype's schema model,
but it is not required to capture the main textual information contained in
an object.

In other words:
  - markdown and snippet carry the content,
  - type mainly describes how Anytype categorizes and presents that content.

If richer semantic export is needed later, type support can be restored to
attach schema/category information to exported documents.
*/

// type Type struct {
// 	Archived   bool       `json:"archived"`
// 	ID         string     `json:"id"`
// 	Key        string     `json:"key"`
// 	Layout     string     `json:"layout"`
// 	Name       string     `json:"name"`
// 	Object     string     `json:"object"`
// 	PluralName string     `json:"plural_name"`
// 	Properties []Property `json:"properties"`
//
// 	// Icon *Icon `json:"icon"`
// }

/*
 NOTE: Property handling

Anytype properties are structured metadata fields attached to an object.

Examples include values such as:
  - text,
  - number,
  - checkbox,
  - url,
  - email,
  - phone,
  - dates,
  - files,
  - object references,
  - and select/multi-select values.

Properties are useful for preserving structured semantics, filtering, and
database-like organization inside Anytype. However, they are not the primary
target of this exporter.

This exporter intentionally does NOT currently support exporting properties.

Reason:
The current goal is document scraping for LLM ingestion, where the highest-
value fields are the object's human-readable textual content, especially
"name", "markdown", and "snippet".

Supporting properties fully would require handling many format-specific
shapes and value encodings. That additional complexity is not necessary for
basic content extraction and would provide limited benefit relative to the
main exporter objective.

If future use cases require richer structured metadata, property support can
be added and mapped into frontmatter, JSON metadata, or other downstream
representations.
*/

// type Property struct {
// 	Format string `json:"format"`
// 	ID     string `json:"id"`
// 	Key    string `json:"key"`
// 	Name   string `json:"name"`
// 	Object string `json:"object"`
//
// 	Text     string  `json:"text,omitempty"`
// 	Number   float64 `json:"number,omitempty"`
// 	Checkbox bool    `json:"checkbox,omitempty"`
// 	URL      string  `json:"url,omitempty"`
// 	Email    string  `json:"email,omitempty"`
// 	Phone    string  `json:"phone,omitempty"`
// }
