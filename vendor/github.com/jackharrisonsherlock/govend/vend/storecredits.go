// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"fmt"
)

// StoreCreditPayload hold Gift Card data
type StoreCreditPayload struct {
	Data []StoreCredit `json:"data"`
}

// StoreCredit contains Store Credit data
type StoreCredit struct {
	ID                      *string                  `json:"id"`
	CustomerID              *string                  `json:"customer_id"`
	CustomerCode            *string                  `json:"customer_code"`
	CreatedAt               *string                  `json:"created_at"`
	Balance                 *float64                 `json:"balance"`
	TotalIssued             *float64                 `json:"total_credit_issued"`
	TotalRedeemed           *float64                 `json:"total_credit_redeemed"`
	StoreCreditTransactions []StoreCreditTransaction `json:"store_credit_transactions"`
}

// StoreCreditTransaction is a Store Credit object.
type StoreCreditTransaction struct {
	ID           *string `json:"id,omitempty"`
	CustomerCode string  `json:"-"`
	CustomerID   *string `json:"-"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
	Notes        *string `json:"notes"`
	UserID       *string `json:"user_id"`
	SaleID       *string `json:"sale_id,omitempty"`
	ClientID     *string `json:"client_id,omitempty"`
	CreatedAt    *string `json:"created_at,omitempty"`
}

// StoreCredits gets all Store Credit data from a store.
func (c *Client) StoreCredits() ([]StoreCredit, error) {

	storecredits := []StoreCredit{}

	url := fmt.Sprintf("https://%v.vendhq.com/api/2.0/store_credits?page_size=1000", c.DomainPrefix)
	data, err := c.MakeRequest("GET", url, nil)
	if err != nil {
		return []StoreCredit{}, fmt.Errorf("Failed to retrieve a page of data %v", err)
	}

	payload := StoreCreditPayload{}

	// Unmarshal payload into Store Credit object.
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return []StoreCredit{}, err
	}

	// Append page to list.
	storecredits = append(storecredits, payload.Data...)

	return storecredits, err
}
