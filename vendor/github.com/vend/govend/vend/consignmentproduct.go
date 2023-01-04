// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"fmt"
	"time"
)

// Vend API Docs: https://docs.vendhq.com/v0.9/reference#consignments-2

// ConsignmentProductPayload contains data and versioning info.
type ConsignmentProductPayload struct {
	Data    []ConsignmentProduct `json:"data,omitempty"`
	Version map[string]int64     `json:"version,omitempty"`
}

// ConsignmentProduct is a ConsignmentProductPayload object.
type ConsignmentProduct struct {
	ProductID *string `json:"product_id,omitempty"`
	SKU       *string `json:"product_sku,omitempty"`
	Count     *string `json:"count,omitempty"`
	Received  *string `json:"received,omitempty"`
	Cost      *string `json:"cost,omitempty"`
	// Name      *string    `json:"name,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// ConsignmentProducts gets all products inside Stock consignments and transfers from a store.
func (c *Client) ConsignmentProducts(consignments *[]Consignment) ([]ConsignmentProduct, map[string][]ConsignmentProduct, error) {

	// var err error
	// var data response.Data
	consignmentProducts := []ConsignmentProduct{}
	consignmentProductMap := make(map[string][]ConsignmentProduct)

	var URL string

	for _, consignment := range *consignments {

		response := ConsignmentProductPayload{}
		data := response.Data

		// Check and ignore cancelled consignments.
		if *consignment.Status == "CANCELLED" {
			continue
		}

		// Build the URL for the consignment product page.
		URL = c.urlFactory(0, *consignment.ID, "consignments")

		body, err := c.MakeRequest("GET", URL, nil)
		if err != nil {
			fmt.Printf("Error getting resource: %s", err)
		}

		// Decode the JSON into our defined consignment object.
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Printf("\nError unmarshalling Vend consignment payload: %s", err)
			return []ConsignmentProduct{}, nil, err
		}

		// Data is an array of consignment product objects.
		data = response.Data

		for _, product := range data {
			consignmentProductMap[*consignment.ID] = append(consignmentProductMap[*consignment.ID], product)
		}

		// Append each lot of consignment products to our list.
		consignmentProducts = append(consignmentProducts, data...)
	}

	return consignmentProducts, consignmentProductMap, nil
}
