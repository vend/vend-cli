// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"fmt"
)

// Vend API Docs: https://docs.vendhq.com/v0.9/reference#customers-2

// CustomerSearchResponse is a data object to hold Customer
type CustomerSearchResponse struct {
	Data []*Customer `json:"data,omitempty"`
}

// Customer is a customer object.
type Customer struct {
	ID                *string  `json:"id,omitempty"`
	Code              *string  `json:"customer_code,omitempty"`
	FirstName         *string  `json:"first_name,omitempty"`
	LastName          *string  `json:"last_name,omitempty"`
	Email             *string  `json:"email,omitempty"`
	YearToDate        *float64 `json:"year_to_date,omitempty"`
	Balance           *float64 `json:"balance,omitempty"`
	LoyaltyBalance    *float64 `json:"loyalty_balance,omitempty"`
	Note              *string  `json:"note,omitempty"`
	Gender            *string  `json:"gender,omitempty"`
	DateOfBirth       *string  `json:"date_of_birth,omitempty"`
	CompanyName       *string  `json:"company_name,omitempty"`
	GroupId           *string  `json:"customer_group_id,omitempty"`
	DoNotEmail        *bool    `json:"do_not_email,omitempty"`
	Phone             *string  `json:"phone,omitempty"`
	Mobile            *string  `json:"mobile,omitempty"`
	Fax               *string  `json:"fax,omitempty"`
	Twitter           *string  `json:"twitter,omitempty"`
	Website           *string  `json:"website,omitempty"`
	PhysicalSuburb    *string  `json:"physical_suburb,omitempty"`
	PhysicalCity      *string  `json:"physical_city,omitempty"`
	PhysicalPostcode  *string  `json:"physical_postcode,omitempty"`
	PhysicalState     *string  `json:"physical_state,omitempty"`
	PostalSuburb      *string  `json:"postal_suburb,omitempty"`
	PostalCity        *string  `json:"postal_city,omitempty"`
	PostalState       *string  `json:"postal_state,omitempty"`
	CreatedAt         *string  `json:"created_at,omitempty"`
	PostalPostcode    *string  `json:"postal_postcode,omitempty"`
	PhysicalAddress1  *string  `json:"physical_address_1,omitempty"`
	PhysicalAddress2  *string  `json:"physical_address_2,omitempty"`
	PostalAddress1    *string  `json:"postal_address_1,omitempty"`
	PostalAddress2    *string  `json:"postal_address_2,omitempty"`
	PostalCountryID   *string  `json:"postal_country_id,omitempty"`
	CustomField1      *string  `json:"custom_field_1,omitempty"`
	CustomField2      *string  `json:"custom_field_2,omitempty"`
	CustomField3      *string  `json:"custom_field_3,omitempty"`
	CustomField4      *string  `json:"custom_field_4,omitempty"`
	DeletedAt         *string  `json:"deleted_at"`
	LoyaltyAdjustment *string  `json:"loyalty_adjustment"`
}

type CustomerGroups struct {
	ID   string `json:"id,omitempty`
	Name string `json:"name,omitempty"`
}

// Customers grabs and collates all customers in pages of 10,000.
func (c *Client) Customers() ([]Customer, error) {

	customers := []Customer{}
	page := []Customer{}

	// v is a version that is used to get customers by page.
	data, v, err := c.ResourcePage(0, "GET", "customers")
	if err != nil {
		return customers, err
	}
	err = json.Unmarshal(data, &page)
	if err != nil {
		err = fmt.Errorf("error while unmarshalling: %s", err)
		return customers, err
	}

	customers = append(customers, page...)

	// Use version to paginate through all pages
	for len(page) > 0 {
		page = []Customer{}
		data, v, err = c.ResourcePage(v, "GET", "customers")
		if err != nil {
			return customers, err
		}
		err = json.Unmarshal(data, &page)
		if err != nil {
			err = fmt.Errorf("error while unmarshalling: %s", err)
			return customers, err
		}
		customers = append(customers, page...)
	}

	return customers, err
}

func (c *Client) CustomerGroups() (map[string]string, error) {
	groups := []CustomerGroups{}
	page := []CustomerGroups{}

	data, v, err := c.ResourcePage(0, "GET", "customer_groups")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &page)
	if err != nil {
		err = fmt.Errorf("error while unmarshalling: %s", err)
		return nil, err
	}

	groups = append(groups, page...)

	for len(page) > 0 {
		page = []CustomerGroups{}
		data, v, err = c.ResourcePage(v, "GET", "customer_groups")
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &page)
		if err != nil {
			err = fmt.Errorf("error while unmarshalling: %s", err)
			return nil, err
		}
		groups = append(groups, page...)

	}

	CustomerGroupMap := make(map[string]string)
	for _, group := range groups {
		CustomerGroupMap[group.ID] = group.Name
	}

	return CustomerGroupMap, err

}

// CustomerMap maps customer codes to customer ids
func CustomerMap(customers []Customer) map[string]string {

	CustomerMap := make(map[string]string)
	for _, customer := range customers {
		CustomerMap[*customer.Code] = *customer.ID
	}

	return CustomerMap
}
