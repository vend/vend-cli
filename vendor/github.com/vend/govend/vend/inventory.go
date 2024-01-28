package vend

import (
	"encoding/json"
	"fmt"
)

// Inventory struct houses Inventory data
type InventoryRecord struct {
	ID             *string      `json:"id"`
	OutletID       *string      `json:"outlet_id"`
	ProductID      *string      `json:"product_id"`
	InventoryLevel *float64     `json:"inventory_level"`
	CurrentAmount  *float64     `json:"current_amount"`
	Version        *interface{} `json:"version"`
	DeletedAt      *interface{} `json:"deleted_at"`
	AverageCost    *float64     `json:"average_cost"`
	ReorderPoint   *float64     `json:"reorder_point"`
	ReorderAmount  *float64     `json:"reorder_amount"`
}

// Inventory() grabs inventory data and stores it into individual records
func (c *Client) Inventory() ([]InventoryRecord, error) {
	inventoryRecords := []InventoryRecord{}
	page := []InventoryRecord{}

	// v is a version that is used to get registers by page.
	data, v, err := c.ResourcePage(0, "GET", "inventory")
	if err != nil {
		return inventoryRecords, err
	}
	err = json.Unmarshal(data, &page)
	if err != nil {
		err = fmt.Errorf("error while unmarshalling: %s", err)
		return inventoryRecords, err
	}

	inventoryRecords = append(inventoryRecords, page...)

	// Use version to paginate through all pages
	for len(page) > 0 {
		page = []InventoryRecord{}
		data, v, err = c.ResourcePage(v, "GET", "inventory")
		if err != nil {
			return inventoryRecords, err
		}
		err = json.Unmarshal(data, &page)
		if err != nil {
			err = fmt.Errorf("error while unmarshalling: %s", err)
			return inventoryRecords, err
		}
		inventoryRecords = append(inventoryRecords, page...)
	}

	return inventoryRecords, err

}
