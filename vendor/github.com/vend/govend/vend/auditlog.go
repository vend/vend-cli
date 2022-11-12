package vend

import (
	"encoding/json"
	"fmt"
)

type AuditResponse struct {
	Data []AuditLog `json:"data"`
}

// Auditlog is a basic auditlog  object.
type AuditLog struct {
	ID         *string `json:"id"`
	UserID     *string `json:"user_id"`
	Kind       *string `json:"type"`
	Action     *string `json:"action"`
	EntityID   *string `json:"entity_id"`
	IPAddress  *string `json:"ip_address"`
	UserAgent  *string `json:"user_agent"`
	OccurredAt *string `json:"occurred_at"`
	CreatedAt  *string `json:"created_at"`
}

// Auditlog grabs and collates all logs in pages of 1,000.
func (c *Client) AuditLog(dateFrom, dateTo string) ([]AuditLog, error) {

	currentOffset := 0
	audit := []AuditLog{}

	// Build the URL for the endpoint.
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/auditlog_events?from=%s&to=%s&offset=%v", c.DomainPrefix, dateFrom, dateTo, currentOffset)
	data, err := c.MakeRequest("GET", url, nil)
	response := &AuditResponse{}
	err = json.Unmarshal(data, response)
	if err != nil {
		fmt.Printf("\nError unmarshalling Vend register payload: %s", err)
		return nil, err
	}

	// Set lastcount based on the response
	lastCount := len(response.Data)
	if lastCount > 0 {
		audit = append(audit, response.Data...)
	}

	for lastCount > 0 {
		currentOffset += lastCount

		// Build the URL for the endpoint including the offset
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/auditlog_events?from=%s&to=%s&offset=%v", c.DomainPrefix, dateFrom, dateTo, currentOffset)
		data, err := c.MakeRequest("GET", url, nil)
		response := &AuditResponse{}
		err = json.Unmarshal(data, response)
		if err != nil {
			fmt.Printf("\nError unmarshalling Vend register payload: %s", err)
			return audit, err
		}

		lastCount = len(response.Data)
		if lastCount > 0 {
			audit = append(audit, response.Data...)
		}
	}
	return audit, nil
}
