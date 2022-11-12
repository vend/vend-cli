// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"fmt"
)

// GiftCardPayload hold Gift Card data
type GiftCardPayload struct {
	Data []GiftCard `json:"data"`
}

// GiftCard contains Gift card data
type GiftCard struct {
	ID                   *string               `json:"id"`
	Number               *string               `json:"number"`
	SaleID               *string               `json:"sale_id"`
	CreatedAt            *string               `json:"created_at"`
	ExpiresAt            *string               `json:"expires_at"`
	Status               *string               `json:"status"`
	Balance              *float64              `json:"balance"`
	TotalSold            *float64              `json:"total_sold"`
	TotalRedeemed        *float64              `json:"total_redeemed"`
	GiftCardTransactions []GiftCardTransaction `json:"gift_card_transactions"`
}

// GiftCardTransaction is a Gift Card object.
type GiftCardTransaction struct {
	ID        *string  `json:"id"`
	Amount    *float64 `json:"amount"`
	Type      *string  `json:"type"`
	UserID    *string  `json:"user_id"`
	CreatedAt *string  `json:"created_at"`
}

// GiftCards gets all gift card data from a store.
func (c *Client) GiftCards() ([]GiftCard, error) {

	giftcards := []GiftCard{}
	payload := GiftCardPayload{}

	// Here we get the first page.
	data, lastID, err := c.ResourcePageFlake("", "GET", "balances/gift_cards")
	if err != nil {
		return []GiftCard{}, fmt.Errorf("Failed to retrieve a page of data %v", err)
	}

	// Unmarshal payload into Gift Card object.
	err = json.Unmarshal(data, &payload)
	if err != nil {
		return []GiftCard{}, err
	}

	// Append page to list.
	giftcards = append(giftcards, payload.Data...)

	// NOTE: Turns out empty response is 2bytes.
	for len(payload.Data) > 1 {
		payload = GiftCardPayload{}

		// Continue grabbing pages until we receive an empty one.
		data, lastID, err = c.ResourcePageFlake(lastID, "GET", "balances/gift_cards")
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(data, &payload)
		if err != nil {
			return []GiftCard{}, err
		}

		// Last page will always return a gift card from the previous payload, removes the last gift card.
		if len(payload.Data) > 1 {
			giftcards = append(giftcards, payload.Data...)
		}
	}

	return giftcards, err
}
