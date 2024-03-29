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
	CustomerID   string  `json:"-"`
	Amount       float64 `json:"amount"`
	Type         string  `json:"type"`
	Notes        *string `json:"notes"`
	UserID       *string `json:"user_id"`
	SaleID       *string `json:"sale_id,omitempty"`
	ClientID     *string `json:"client_id,omitempty"`
	CreatedAt    *string `json:"created_at,omitempty"`
}

type StoreCreditCsv struct {
	CustomerID   *string
	CustomerCode *string
	Amount       *float64
}

// StoreCredits gets all Store Credit data from a store.
func (c *Client) StoreCredits() ([]StoreCredit, error) {

	// this endpoint does not support pagination.
	// #TODO This limit should be adjusted if/ once that endpoint supports pagination
	const storeCreditLimit = 1000000
	storecredits := []StoreCredit{}

	url := fmt.Sprintf("https://%v.vendhq.com/api/2.0/store_credits?page_size=%d", c.DomainPrefix, storeCreditLimit)
	data, err := c.MakeRequest("GET", url, nil)
	if err != nil {
		return []StoreCredit{}, fmt.Errorf("failed to retrieve a page of data: %w", err)
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

// CreditMap maps customer ids to store credit balances
func CreditMap(storeCredits []StoreCredit) map[string]float64 {

	CreditMap := make(map[string]float64)
	for _, storeCredit := range storeCredits {
		CreditMap[*storeCredit.CustomerID] = *storeCredit.Balance
	}

	return CreditMap
}
