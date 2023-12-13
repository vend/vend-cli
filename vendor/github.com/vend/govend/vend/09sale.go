package vend

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

type RegisterSale9 struct {
	RegisterSale9 []Sale9 `json:"register_sales"`
}

type Sale9 struct {
	ID                    *string               `json:"id"`
	Source                *string               `json:"source"`
	SourceID              *string               `json:"source_id"`
	RegisterID            *string               `json:"register_id"`
	MarketID              *string               `json:"market_id"`
	CustomerName          *string               `json:"customer_name"`
	UserID                *string               `json:"user_id"`
	UserName              *string               `json:"user_name"`
	SaleDate              *string               `json:"sale_date"`
	CreatedAt             *string               `json:"created_at"`
	UpdatedAt             *string               `json:"updated_at"`
	TotalPrice            *float64              `json:"total_price"`
	TotalCost             *float64              `json:"total_cost"`
	TotalTax              *float64              `json:"total_tax"`
	Note                  *string               `json:"note"`
	Status                *string               `json:"status"`
	ShortCode             *string               `json:"short_code"`
	InvoiceNumber         *string               `json:"invoice_number"`
	AccountsTransactionID *string               `json:"accounts_transaction_id"`
	ReturnFor             *string               `json:"return_for"`
	RegisterSaleProducts  []RegisterSaleProduct `json:"register_sale_products"`
	RegisterSalePayments  []RegisterSalePayment `json:"register_sale_payments"`
	CustomerID            *string               `json:"customer_id"`
	TaxName               *string               `json:"tax_name"`
	Taxes                 []Sale9Tax            `json:"taxes"`
	Totals                Totals                `json:"totals"`
}

type RegisterSaleProduct struct {
	ID                             *string       `json:"id"`
	ProductID                      *string       `json:"product_id"`
	RegisterID                     *string       `json:"register_id"`
	Sequence                       EnsureInt64   `json:"sequence"`
	Handle                         *string       `json:"handle"`
	Sku                            *string       `json:"sku"`
	Name                           *string       `json:"name"`
	Quantity                       *float64      `json:"quantity"`
	Price                          *float64      `json:"price"`
	Cost                           *float64      `json:"cost"`
	PriceSet                       *int64        `json:"price_set"`
	Discount                       *float64      `json:"discount"`
	LoyaltyValue                   *float64      `json:"loyalty_value"`
	Tax                            *float64      `json:"tax"`
	TaxID                          *string       `json:"tax_id"`
	TaxName                        *string       `json:"tax_name"`
	TaxRate                        *float64      `json:"tax_rate"`
	TaxTotal                       *float64      `json:"tax_total"`
	PriceTotal                     *float64      `json:"price_total"`
	DisplayRetailPriceTaxInclusive EnsureFloat64 `json:"display_retail_price_tax_inclusive"`
	Status                         *string       `json:"status"`
	Attributes                     []Attribute   `json:"attributes"`
}

type RegisterSalePayment struct {
	Amount                *float64    `json:"amount"`
	ID                    *string     `json:"id"`
	RegisterID            *string     `json:"register_id"`
	RetailerPaymentTypeID *string     `json:"retailer_payment_type_id"`
	PaymentDate           *string     `json:"payment_date"`
	Currency              *string     `json:"currency"`
	PaymentTypeID         EnsureInt64 `json:"payment_type_id"`
	Name                  *string     `json:"name"`
}

type Attribute struct {
	Name  *string `json:"name"`
	Value *string `json:"value"`
}

type Sale9Tax struct {
	Tax  *float64 `json:"tax"`
	Name *string  `json:"name"`
	Rate *float64 `json:"rate"`
	ID   *string  `json:"id"`
}

type Totals struct {
	TotalTax     *float64 `json:"total_tax"`
	TotalPrice   *float64 `json:"total_price"`
	TotalPayment *float64 `json:"total_payment"`
	TotalToPay   *float64 `json:"total_to_pay"`
}

type SaleUserUpload struct {
	SaleID string
	UserID string
}

type EnsureInt64 struct {
	Value *int64
}

// checks the interface and if a string converts it to an int64
// for some reason the vend api returns this as a string eventhough docs say it must be int
// https://x-series-api.lightspeedhq.com/reference/listregistersales
// but not always, erred sales are returned as int64 as expected so we need to check and handle both
func (e *EnsureInt64) UnmarshalJSON(b []byte) error {

	var raw interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case int64:
		e.Value = &v
	case float64:
		num := int64(v)
		e.Value = &num
	case string:
		num, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return err
		}
		e.Value = &num
	default:
		return fmt.Errorf("unexpected type %T", v)
	}
	return nil
}

type EnsureFloat64 struct {
	Value *float64
}

// checks the interface and if a string converts it to an float64
// for some reason the vend api returns this as a string eventhough docs say it must be int
// https://x-series-api.lightspeedhq.com/reference/listregistersales
// but not always, erred sales are returned as int64 as expected so we need to check and handle both
func (e *EnsureFloat64) UnmarshalJSON(b []byte) error {

	var raw interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	switch v := raw.(type) {
	case float64:
		e.Value = &v
	case string:
		num, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return err
		}
		e.Value = &num
	default:
		return fmt.Errorf("unexpected type %T", v)
	}
	return nil
}

// Error9 houses error responses for the 0.9 Vend API
type Error9 struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Details string `json:"details"`
}

// converts a slice of Sale9 to a slice of Sale
// NOTE: This is a very naive conversion,
// and should not be used for POST without further review
func ConvertSale9ToSale(sale9Array []Sale9) ([]Sale, error) {
	var saleArray []Sale

	for _, sale9 := range sale9Array {
		sale := Sale{
			ID:              sale9.ID,
			OutletID:        nil, // No equivalent field in Sale9
			RegisterID:      sale9.RegisterID,
			UserID:          sale9.UserID,
			CustomerID:      sale9.CustomerID,
			InvoiceNumber:   sale9.InvoiceNumber,
			ReceiptNumber:   nil, // No equivalent field in Sale9
			InvoiceSequence: nil, // No equivalent field in Sale9
			ReceiptSequence: nil, // No equivalent field in Sale9
			Status:          sale9.Status,
			Note:            sale9.Note,
			ShortCode:       sale9.ShortCode,
			ReturnFor:       sale9.ReturnFor,
			CreatedAt:       sale9.CreatedAt,
			UpdatedAt:       sale9.UpdatedAt,
			SaleDate:        convertSaleDate(sale9.SaleDate),
			DeletedAt:       nil, // No equivalent field in Sale9
			TotalPrice:      sale9.TotalPrice,
			TotalLoyalty:    nil, // No equivalent field in Sale9
			TotalTax:        sale9.Totals.TotalTax,
			LineItems:       &[]LineItem{}, // will set this below
			Payments:        &[]Payment{},  // will set this below
			Taxes:           nil,           // No equivalent field in Sale9
			VersionNumber:   nil,           // No equivalent field in Sale9
		}

		// Map line items from Sale9 to LineItems in Sale
		for _, product := range sale9.RegisterSaleProducts {
			lineItem := LineItem{
				ID:                product.ID,
				ProductID:         product.ProductID,
				Quantity:          product.Quantity,
				Price:             product.Price,
				UnitPrice:         nil, // No equivalent field in Sale9
				PriceTotal:        product.PriceTotal,
				TotalPrice:        product.PriceTotal, // Assuming TotalPrice is same as PriceTotal
				Discount:          product.Discount,
				UnitDiscount:      nil, // No equivalent field in Sale9
				DiscountTotal:     nil, // No equivalent field in Sale9
				TotalDiscount:     nil, // No equivalent field in Sale9
				LoyaltyValue:      product.LoyaltyValue,
				UnitLoyaltyValue:  nil, // No equivalent field in Sale9
				TotalLoyaltyValue: nil, // No equivalent field in Sale9
				Cost:              product.Cost,
				UnitCost:          nil, // No equivalent field in Sale9
				CostTotal:         nil, // No equivalent field in Sale9
				TotalCost:         nil, // No equivalent field in Sale9
				Tax:               product.Tax,
				UnitTax:           nil, // No equivalent field in Sale9
				TaxTotal:          product.TaxTotal,
				TotalTax:          product.TaxTotal, // Assuming TotalTax is same as TaxTotal
				TaxID:             product.TaxID,
				PriceSet:          convertPriceSet(product.PriceSet),
				Sequence:          product.Sequence.Value,
				Status:            product.Status,
				IsReturn:          nil, // No equivalent field in Sale9
				TaxComponents:     nil, // No equivalent field in Sale9
			}

			*sale.LineItems = append(*sale.LineItems, lineItem)
		}

		// Map payments from Sale9 to Payments in Sale
		for _, payment := range sale9.RegisterSalePayments {
			salePayment := Payment{
				ID:                    payment.ID,
				RegisterID:            payment.RegisterID,
				RetailerPaymentTypeID: payment.RetailerPaymentTypeID,
				PaymentTypeID:         convertPaymentTypeId(payment.PaymentTypeID.Value),
				Name:                  payment.Name,
				PaymentDate:           parseTime(payment.PaymentDate),
				Amount:                payment.Amount,
			}
			*sale.Payments = append(*sale.Payments, salePayment)
		}
		saleArray = append(saleArray, sale)
	}

	return saleArray, nil
}

func convertPriceSet(i *int64) *bool {
	result := false
	if i != nil && *i != 0 {
		result = true
	}
	return &result
}

func convertPaymentTypeId(i *int64) *string {
	result := ""
	if i != nil {
		result = strconv.FormatInt(*i, 10)
	}
	return &result
}

func convertSaleDate(sale9Date *string) *string {
	if sale9Date == nil {
		return nil
	}

	// Parse the Sale9 date format
	parsedTime, err := time.Parse("2006-01-02 15:04:05", *sale9Date)
	if err != nil {
		return nil
	}

	// Format the time to the desired Sale date format
	formattedDate := parsedTime.Format("2006-01-02T15:04:05Z")

	return &formattedDate
}

func parseTime(sale9Date *string) *time.Time {
	if sale9Date == nil {
		return nil
	}

	// Parse the Sale9 date format
	parsedTime, err := time.Parse("2006-01-02 15:04:05", *sale9Date)
	if err != nil {
		return nil
	}

	return &parsedTime
}
