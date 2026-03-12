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

Anytype list endpoints commonly return a result set plus pagination metadata.

For the purposes of this exporter, these wrappers exist only to support
iterating through paginated API responses while collecting content-bearing
objects for downstream processing.

The pagination fields are preserved because they are required for complete
traversal of the dataset, even though they are not themselves meaningful
document content.
*/

type SpacesResponse struct {
	Objects    []Space    `json:"data"`
	Pagination Pagination `json:"pagination"`
}

type ObjectsResponse struct {
	Objects    []Object   `json:"data"`
	Pagination Pagination `json:"pagination"`
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
  - gateway_url → gateway base used by Anytype for serving related media/files

This exporter does not attempt to reconstruct full space presentation or UI
state. Space records are used primarily as contextual metadata for the
documents scraped from them.
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

The fields retained here are the ones most useful for document extraction
and LLM ingestion:

  - id        → stable object identifier
  - name      → object title, when present
  - markdown  → primary body content
  - snippet   → short preview text, especially useful for objects that may
                have little or no explicit title
  - space_id  → associates the object with its parent space
  - archived  → allows archived content to be filtered if desired
  - layout    → preserved only as lightweight context about the object's
                presentation category

This exporter intentionally focuses on textual and contextual data rather
than full object fidelity. As such, schema metadata, icon metadata, and
other presentation-oriented fields are not currently exported.
*/

type Object struct {
	Archived bool   `json:"archived"`
	ID       string `json:"id"`
	Layout   string `json:"layout"`
	Markdown string `json:"markdown"`
	Name     string `json:"name"`
	Snippet  string `json:"snippet"`
	SpaceID  string `json:"space_id"`

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
 TODO: Type handling

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
