// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"log"
	"time"
)

// Vend API Docs: https://docs.vendhq.com/v0.9/reference#consignments-2

// ConsignmentPayload contains register data and versioning info.
type ConsignmentPayload struct {
	Data    []Consignment    `json:"data,omitempty"`
	Version map[string]int64 `json:"version,omitempty"`
}

// Consignment is a ConsignmentPayload object.
type Consignment struct {
	ID              *string    `json:"id,omitempty"`
	OutletID        *string    `json:"outlet_id,omitempty"`
	Name            *string    `json:"name,omitempty"`
	Type            *string    `json:"type,omitempty"`
	Status          *string    `json:"status,omitempty"`
	ConsignmentDate *string    `json:"consignment_date,omitempty"` // NOTE: Using string for ParseVendDT.
	DeletedAt       *time.Time `json:"deleted_at,omitempty"`
}

// Consignments gets all stock consignments and transfers from a store.
func (c *Client) Consignments() ([]Consignment, error) {

	var consignments, page []Consignment
	var v int64

	// v is a version that is used to objects by page.
	data, v, err := c.ResourcePage(0, "GET", "consignments")
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Printf("error while unmarshalling: %s", err)
	}

	consignments = append(consignments, page...)

	// Use version to paginate through all pages
	for len(data) > 2 {
		page = []Consignment{}
		data, v, err = c.ResourcePage(v, "GET", "consignments")
		err = json.Unmarshal(data, &page)
		consignments = append(consignments, page...)
	}

	return consignments, err
}
