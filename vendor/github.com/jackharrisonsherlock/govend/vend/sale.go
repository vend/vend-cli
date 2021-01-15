package vend

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

// Vend API Docs: https://docs.vendhq.com/v0.9/reference#sales

// RegisterSale holds the Sale object
type RegisterSales struct {
	RegisterSales []Sale `json:"register_sales"`
}

// Sale is a basic sale object.
type Sale struct {
	ID              *string     `json:"id,omitempty"`
	OutletID        *string     `json:"outlet_id,omitempty"`
	RegisterID      *string     `json:"register_id,omitempty"`
	UserID          *string     `json:"user_id,omitempty"`
	CustomerID      *string     `json:"customer_id,omitempty"`
	InvoiceNumber   *string     `json:"invoice_number,omitempty"`
	ReceiptNumber   *string     `json:"receipt_number,omitempty"`
	InvoiceSequence *int64      `json:"invoice_sequence,omitempty"`
	ReceiptSequence *int64      `json:"receipt_sequence,omitempty"`
	Status          *string     `json:"status,omitempty"`
	Note            *string     `json:"note,omitempty"`
	ShortCode       *string     `json:"short_code,omitempty"`
	ReturnFor       *string     `json:"return_for,omitempty"`
	CreatedAt       *string     `json:"created_at,omitempty"`
	UpdatedAt       *string     `json:"updated_at,omitempty"`
	SaleDate        *string     `json:"sale_date,omitempty"`
	DeletedAt       *string     `json:"deleted_at,omitempty"`
	TotalPrice      *float64    `json:"total_price,omitempty"`
	TotalLoyalty    *float64    `json:"total_loyalty,omitempty"`
	TotalTax        *float64    `json:"total_tax,omitempty"`
	LineItems       *[]LineItem `json:"line_items,omitempty"`
	Payments        *[]Payment  `json:"payments,omitempty"`
	Taxes           *[]SaleTax  `json:"taxes,omitempty"`
	VersionNumber   *int64      `json:"version,omitempty"`
}

// LineItem is a product on a sale.
type LineItem struct {
	ID                *string         `json:"id,omitempty"`
	ProductID         *string         `json:"product_id,omitempty"`
	Quantity          *float64        `json:"quantity,omitempty"`
	Price             *float64        `json:"price,omitempty"`
	UnitPrice         *float64        `json:"unit_price,omitempty"`
	PriceTotal        *float64        `json:"price_total,omitempty"`
	TotalPrice        *float64        `json:"total_price,omitempty"`
	Discount          *float64        `json:"discount,omitempty"`
	UnitDiscount      *float64        `json:"unit_discount,omitempty"`
	DiscountTotal     *float64        `json:"discount_total,omitempty"`
	TotalDiscount     *float64        `json:"total_discount,omitempty"`
	LoyaltyValue      *float64        `json:"loyalty_value,omitempty"`
	UnitLoyaltyValue  *float64        `json:"unit_loyalty_value,omitempty"`
	TotalLoyaltyValue *float64        `json:"total_loyalty_value,omitempty"`
	Cost              *float64        `json:"cost,omitempty"`
	UnitCost          *float64        `json:"unit_cost,omitempty"`
	CostTotal         *float64        `json:"cost_total,omitempty"`
	TotalCost         *float64        `json:"total_cost,omitempty"`
	Tax               *float64        `json:"tax,omitempty"`
	UnitTax           *float64        `json:"unit_tax,omitempty"`
	TaxTotal          *float64        `json:"tax_total,omitempty"`
	TotalTax          *float64        `json:"total_tax,omitempty"`
	TaxID             *string         `json:"tax_id,omitempty"`
	PriceSet          *bool           `json:"price_set,omitempty"`
	Sequence          *int64          `json:"sequence,omitempty"`
	Status            *string         `json:"status,omitempty"`
	IsReturn          *bool           `json:"is_return,omitempty"`
	TaxComponents     *[]TaxComponent `json:"tax_components,omitempty"`
}

// TaxComponent is a tax object on a sale.
type TaxComponent struct {
	RateID   string
	TotalTax int64
}

// Payment is a payment on a sale.
type Payment struct {
	ID                    *string    `json:"id,omitempty"`
	RegisterID            *string    `json:"register_id,omitempty"`
	RetailerPaymentTypeID *string    `json:"retailer_payment_type_id,omitempty"`
	PaymentTypeID         *string    `json:"payment_type_id,omitempty"`
	Name                  *string    `json:"name,omitempty"`
	PaymentDate           *time.Time `json:"payment_date,omitempty"`
	Amount                *float64   `json:"amount,omitempty"`
}

// SaleTax is tax on a sale.
type SaleTax struct {
	ID     *string  `json:"id,omitempty"`
	Amount *float64 `json:"amount,omitempty"`
}

type SalesResponse struct {
	Data    []Sale   `json:"data"`
	Version *Version `json:"version"`
}

type Version struct {
	Max int64 `json:"max"`
	Min int64 `json:"min"`
}

// Sales grabs all-time sales - version after 0
func (c *Client) Sales() ([]Sale, error) {
	return salesAfterVersion(0, c)
}

// SalesAfter grabs sales after the provided version
func (c *Client) SalesAfter(version int64) ([]Sale, error) {
	return salesAfterVersion(version, c)
}

// salesAfterVersion grabs sales after the specified version
func salesAfterVersion(version int64, c *Client) ([]Sale, error) {
	sales := []Sale{}
	page := []Sale{}

	// v is a version that is used to get customers by page.
	data, v, err := c.ResourcePage(version, "GET", "sales")
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Printf("error while unmarshalling: %s", err)
	}

	sales = append(sales, page...)

	// Use version to paginate through all pages
	for len(page) > 0 {
		page = []Sale{}
		data, v, err = c.ResourcePage(v, "GET", "sales")
		err = json.Unmarshal(data, &page)
		sales = append(sales, page...)
	}

	return sales, err
}

// GetStartVersion retrieves the version of the sale offset by a couple days
// of the specified dateFrom.  Time object and string both needed from the calling
// function
func (c *Client) GetStartVersion(dateFrom time.Time, dateStr string) (int64, error) {
	offSetTime := dateFrom.AddDate(0, 0, -7)

	queryDateFrom := offSetTime.Format("2006-01-02T15:04:05Z")

	//debug
	// fmt.Println(queryDateFrom)

	endpoint := fmt.Sprintf("https://%s.vendhq.com/api/2.0/search?type=sales&date_from=%s&page_size=1&order_direction=asc", c.DomainPrefix, queryDateFrom)

	body, err := c.MakeRequest("GET", endpoint, nil)
	if err != nil {
		return 0, err
	}

	response := Payload{}

	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Printf("Error unmarshalling payload: %s", err)
		return 0, err
	}

	data := response.Data

	// no sale
	if len(data) == 0 {
		return 0, nil
	}

	//this will be a single sale, page_size=1
	sales := []Sale{}
	err = json.Unmarshal(data, &sales)
	sales = append(sales, sales...)

	sale := sales[0]

	// fmt.Printf("In GetStartVersion: datefrom = %v, version = %d", dateFrom, *sale.VersionNumber)

	return *sale.VersionNumber, err
}

// // SaleSearch for Sales based on Outlet and date range
// func (c *Client) SalesSearch(dateFrom, dateTo, outlet string) ([]Sale, error) {

// 	currentOffset := 0
// 	AllSales := []Sale{}
// 	outletID := ""

// 	// Get outlet ID by name
// 	if outlet != "" {
// 		oID, err := c.getOutlet(outlet)
// 		if err != nil {
// 			fmt.Printf("\nError retrieving Outlets %s", err)
// 			return AllSales, err
// 		}
// 		outletID = oID
// 	}

// 	// Build the URL for the endpoint.
// 	url := buildSearchURL(c.DomainPrefix, dateFrom, dateTo, outletID, currentOffset)
// 	data, statusCode, err := c.MakeRequest("GET", url, nil)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if statusCode > 299 {
// 		return nil, fmt.Errorf("unexpected response status code for request to: %s", url)
// 	}

// 	// Decode the raw JSON.
// 	response := &SalesResponse{}
// 	err = json.Unmarshal(data, response)
// 	if err != nil {
// 		fmt.Printf("\nError unmarshalling Vend register payload: %s", err)
// 		return nil, err
// 	}

// 	// Set lastcount based on the response
// 	lastCount := len(response.Data)
// 	if lastCount > 0 {
// 		AllSales = append(AllSales, response.Data...)
// 	}

// 	for lastCount > 0 {
// 		currentOffset += lastCount
// 		// Build the URL for the endpoint including the offset
// 		url := buildSearchURL(c.DomainPrefix, dateFrom, dateTo, outletID, currentOffset)

// 		data, statusCode, err := c.MakeRequest("GET", url, nil)
// 		if err != nil {
// 			return AllSales, err
// 		}

// 		if statusCode > 299 {
// 			return nil, fmt.Errorf("Unexpected response status code")
// 		}

// 		// Decode the raw JSON.
// 		response := &SalesResponse{}

// 		err = json.Unmarshal(data, response)
// 		if err != nil {
// 			// fmt.Printf("\nError unmarshalling Vend register payload: %s", err)
// 			log.Printf("error decoding response: %v", err)
// 			if e, ok := err.(*json.SyntaxError); ok {
// 				log.Printf("syntax error at byte offset %d", e.Offset)
// 			}
// 			log.Printf("response: %q", data)
// 			return AllSales, err
// 		}

// 		lastCount = len(response.Data)
// 		if lastCount > 0 {
// 			AllSales = append(AllSales, response.Data...)
// 		}
// 	}
// 	return AllSales, nil
// }

// // Get Outlet name by ID
// func (c Client) getOutlet(outlet string) (string, error) {
// 	outlets, _, err := c.Outlets()
// 	if err != nil {
// 		return "", fmt.Errorf("No outlet with the given name: %v", err)
// 	}

// 	nameMap := map[string]string{}
// 	for _, o := range outlets {
// 		nameMap[*o.Name] = *o.ID
// 	}

// 	id, ok := nameMap[outlet]
// 	if !ok {
// 		return "", fmt.Errorf("No outlet with the given name")
// 	}

// 	return id, nil
// }

// // Build URL For Sales Legder Export
// func buildSearchURL(domainPrefix, dateFrom, dateTo, outletID string, currentOffset int) string {
// 	if outletID == "" {
// 		return fmt.Sprintf("https://%s.vendhq.com/api/2.0/search?type=sales&date_from=%s&date_to=%s&page_size=10000&offset=%v", domainPrefix, dateFrom, dateTo, currentOffset)
// 	}

// 	return fmt.Sprintf("https://%s.vendhq.com/api/2.0/search?type=sales&date_from=%s&date_to=%s&outlet_id=%s&page_size=10000&offset=%v", domainPrefix, dateFrom, dateTo, outletID, currentOffset)
// }
