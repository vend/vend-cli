// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"log"
	"time"
)

// Vend API Docs: https://docs.vendhq.com/v0.9/reference#registers-2

// RegisterPayload contains register data and versioning info.
type RegisterPayload struct {
	Data    []Register       `json:"data,omitempty"`
	Version map[string]int64 `json:"version,omitempty"`
}

// Register is a register object.
type Register struct {
	ID        *string    `json:"id,omitempty"`
	Name      *string    `json:"name,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// Registers gets all registers from a store.
func (c *Client) Registers() ([]Register, error) {

	registers := []Register{}
	page := []Register{}

	// v is a version that is used to get registers by page.
	data, v, err := c.ResourcePage(0, "GET", "registers")
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Printf("error while unmarshalling: %s", err)
	}

	registers = append(registers, page...)

	// Use version to paginate through all pages
	for len(page) > 0 {
		page = []Register{}
		data, v, err = c.ResourcePage(v, "GET", "registers")
		err = json.Unmarshal(data, &page)
		registers = append(registers, page...)
	}

	return registers, err
}
