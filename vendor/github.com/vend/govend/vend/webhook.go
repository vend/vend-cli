// Package vend handles interactions with the Vend API.
package vend

// Vend API Docs: https://docs.vendhq.com/v0.9/reference#webhooks-1

// Webhook contains Webhooks data
type Webhook struct {
	ID         *string `json:"id"`
	RetailerID *string `json:"retailer_id"`
	UserID     *string `json:"user_id"`
	URL        *string `json:"url"`
	Active     *bool   `json:"active"`
	Type       *string `json:"type"`
}
