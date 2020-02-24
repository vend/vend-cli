// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"log"
	"time"
)

// Vend API Docs: https://docs.vendhq.com/v0.9/reference#users-2

// UserPayload contains sales data and versioning info.
type UserPayload struct {
	Data    []User           `json:"data,omitempty"`
	Version map[string]int64 `json:"version,omitempty"`
}

// User is a basic user object.
type User struct {
	ID          *string    `json:"id,omitempty"`
	Username    *string    `json:"username,omitempty"`
	DisplayName *string    `json:"display_name,omitempty"`
	Email       *string    `json:"email,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// Users gets all users from a store.
func (c *Client) Users() ([]User, error) {

	users := []User{}
	page := []User{}

	// v is a version that is used to get registers by page.
	data, v, err := c.ResourcePage(0, "GET", "users")
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Printf("error while unmarshalling: %s", err)
	}

	users = append(users, page...)

	// Use version to paginate through all pages
	for len(page) > 0 {
		page = []User{}
		data, v, err = c.ResourcePage(v, "GET", "users")
		err = json.Unmarshal(data, &page)
		users = append(users, page...)
	}

	return users, err
}
