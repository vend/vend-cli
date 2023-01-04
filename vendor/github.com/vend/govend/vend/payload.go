// Package vend handles interactions with the Vend API.
package vend

import "encoding/json"

// Payload contains resource data and versioning info.
// This is the default format returned by 2.0 endpoints.
type Payload struct {
	Data    json.RawMessage  `json:"data,omitempty"`
	Version map[string]int64 `json:"version,omitempty"`
}

type Pagination struct {
	Results  int64 `json:"results"`
	Page     int64 `json:"page"`
	PageSize int64 `json:"page_size"`
	Pages    int64 `json:"pages"`
}

type Errors struct {
	Error struct {
		Global []string `json:"global"`
	} `json:"errors"`
	Reference string `json:"reference"`
}
