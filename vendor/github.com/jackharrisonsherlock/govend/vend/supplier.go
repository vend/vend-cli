// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"fmt"
)

// Vend API Docs: https://docs.vendhq.com/v0.9/reference#suppliers-1

type SupplierCollectionResponse struct {
	Suppliers  []SupplierBase `json:"suppliers"`
	Pagination Pagination     `json:"pagination"`
}

// Supplier contains supplier data.
type SupplierBase struct {
	ID          *string  `json:"id,omitempty"`
	Name        *string  `json:"name,omitempty"`
	Description *string  `json:"description,omitempty"`
	Source      *string  `json:"source,omitempty"`
	Contact     *Contact `json:"contact,omitempty"`
}

// Contact is a supplier object
type Contact struct {
	FirstName         *string `json:"first_name,omitempty"`
	LastName          *string `json:"last_name,omitempty"`
	CompanyName       *string `json:"company_name,omitempty"`
	Phone             *string `json:"phone,omitempty"`
	Mobile            *string `json:"mobile,omitempty"`
	Fax               *string `json:"fax,omitempty"`
	Email             *string `json:"email,omitempty"`
	Twitter           *string `json:"twitter,omitempty"`
	Website           *string `json:"website,omitempty"`
	PhysicalAddress1  *string `json:"physical_address1,omitempty"`
	PhysicalAddress2  *string `json:"physical_address2,omitempty"`
	PhysicalSuburb    *string `json:"physical_suburb,omitempty"`
	PhysicalCity      *string `json:"physical_city,omitempty"`
	PhysicalPostcode  *string `json:"physical_postcode,omitempty"`
	PhysicalState     *string `json:"physical_state,omitempty"`
	PhysicalCountryID *string `json:"physical_country_id,omitempty"`
	PostalAddress1    *string `json:"postal_address1,omitempty"`
	PostalAddress2    *string `json:"postal_address2,omitempty"`
	PostalSuburb      *string `json:"postal_suburb,omitempty"`
	PostalCity        *string `json:"postal_city,omitempty"`
	PostalPostcode    *string `json:"postal_postcode,omitempty"`
	PostalState       *string `json:"postal_state,omitempty"`
	PostalCountryID   *string `json:"postal_country_id,omitempty"`
}

// Suppliers gets all Suppliers from a store.
func (c *Client) Suppliers() ([]SupplierBase, error) {

	suppliers := []SupplierBase{}

	data, more, p, err := c.Pages("api/supplier", 0)
	suppliers = append(suppliers, data...)

	for more {
		// Continue grabbing pages until we receive an empty one.
		data, more, p, err = c.Pages("api/supplier", p)
		if err != nil {
			return nil, err
		}
		// Append supplier page to list of suppliers.
		suppliers = append(suppliers, data...)
	}

	return suppliers, err
}

func (c Client) Pages(resource string, page int64) ([]SupplierBase, bool, int64, error) {

	url := ""

	if page > 0 {
		url = fmt.Sprintf("https://%s.vendhq.com/%s?page_size=200&page=%v", c.DomainPrefix, resource, page)
	} else {
		url = fmt.Sprintf("https://%s.vendhq.com/%s?page_size=200", c.DomainPrefix, resource)
	}

	body, err := c.MakeRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error getting resource: %s", err)
	}

	// Decode the raw JSON.
	response := SupplierCollectionResponse{}
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("\nError unmarshalling payload: %s", err)
		return nil, false, 0, err
	}

	// Data is the resource body.
	data := response.Suppliers

	// Page contains the maximum version number of the resources.
	pages := response.Pagination.Pages
	pg := response.Pagination.Page
	more := pg != pages
	pg++

	return data, more, pg, err
}
