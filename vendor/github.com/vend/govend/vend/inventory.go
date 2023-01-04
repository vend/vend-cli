package vend

import (
	"encoding/json"
	"log"
)

// Inventory struct houses Inventory data
type InventoryRecord struct {
	ID             *string      `json:"id"`
	OutletID       *string      `json:"outlet_id"`
	ProductID      *string      `json:"product_id"`
	InventoryLevel *int64       `json:"inventory_level"`
	CurrentAmount  *int64       `json:"current_amount"`
	Version        *interface{} `json:"version"`
	DeletedAt      *interface{} `json:"deleted_at"`
	AverageCost    *float64     `json:"average_cost"`
	ReorderPoint   *int64       `json:"reorder_point"`
	ReorderAmount  *int64       `json:"reorder_amount"`
}

// Inventory() grabs inventory data and stores it into individual records
func (c *Client) Inventory() ([]InventoryRecord, error) {
	inventoryRecords := []InventoryRecord{}
	page := []InventoryRecord{}

	// v is a version that is used to get registers by page.
	data, v, err := c.ResourcePage(0, "GET", "inventory")
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Printf("error while unmarshalling: %s", err)
	}

	inventoryRecords = append(inventoryRecords, page...)

	// Use version to paginate through all pages
	for len(page) > 0 {
		page = []InventoryRecord{}
		data, v, err = c.ResourcePage(v, "GET", "inventory")
		err = json.Unmarshal(data, &page)
		if err != nil {
			log.Printf("error while unmarshalling: %s", err)
		}
		inventoryRecords = append(inventoryRecords, page...)
	}

	return inventoryRecords, err

}
