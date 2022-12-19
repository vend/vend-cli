package vend

type RegisterSale9 struct {
	RegisterSale9 []Sale9 `json:"register_sales"`
}

type Sale9 struct {
	ID                    *string  `json:"id"`
	Source                *string  `json:"source"`
	SourceID              *string  `json:"source_id"`
	RegisterID            *string  `json:"register_id"`
	MarketID              *string  `json:"market_id"`
	CustomerName          *string  `json:"customer_name"`
	UserID                *string  `json:"user_id"`
	UserName              *string  `json:"user_name"`
	SaleDate              *string  `json:"sale_date"`
	CreatedAt             *string  `json:"created_at"`
	UpdatedAt             *string  `json:"updated_at"`
	TotalPrice            *float64 `json:"total_price"`
	TotalCost             *float64 `json:"total_cost"`
	TotalTax              *float64 `json:"total_tax"`
	Note                  *string  `json:"note"`
	Status                *string  `json:"status"`
	ShortCode             *string  `json:"short_code"`
	InvoiceNumber         *string  `json:"invoice_number"`
	AccountsTransactionID *string  `json:"accounts_transaction_id"`
	ReturnFor             *string  `json:"return_for"`
	RegisterSaleProducts  []struct {
		ID                             *string  `json:"id"`
		ProductID                      *string  `json:"product_id"`
		RegisterID                     *string  `json:"register_id"`
		Sequence                       *string  `json:"sequence"`
		Handle                         *string  `json:"handle"`
		Sku                            *string  `json:"sku"`
		Name                           *string  `json:"name"`
		Quantity                       *float64 `json:"quantity"`
		Price                          *float64 `json:"price"`
		Cost                           *float64 `json:"cost"`
		PriceSet                       *int64   `json:"price_set"`
		Discount                       *float64 `json:"discount"`
		LoyaltyValue                   *float64 `json:"loyalty_value"`
		Tax                            *float64 `json:"tax"`
		TaxID                          *string  `json:"tax_id"`
		TaxName                        *string  `json:"tax_name"`
		TaxRate                        *float64 `json:"tax_rate"`
		TaxTotal                       *float64 `json:"tax_total"`
		PriceTotal                     *float64 `json:"price_total"`
		DisplayRetailPriceTaxInclusive *string  `json:"display_retail_price_tax_inclusive"`
		Status                         *string  `json:"status"`
		Attributes                     []struct {
			Name  *string `json:"name"`
			Value *string `json:"value"`
		} `json:"attributes"`
	} `json:"register_sale_products"`
	CustomerID *string `json:"customer_id"`
	TaxName    *string `json:"tax_name"`
	Taxes      []struct {
		Tax  *float64 `json:"tax"`
		Name *string  `json:"name"`
		Rate *float64 `json:"rate"`
		ID   *string  `json:"id"`
	} `json:"taxes"`
	Totals struct {
		TotalTax     *float64 `json:"total_tax"`
		TotalPrice   *float64 `json:"total_price"`
		TotalPayment *float64 `json:"total_payment"`
		TotalToPay   *float64 `json:"total_to_pay"`
	} `json:"totals"`
}

type SaleUserUpload struct {
	SaleID string
	UserID string
}

// Error9 houses error responses for the 0.9 Vend API
type Error9 struct {
	Status  string `json:"status"`
	Error   string `json:"error"`
	Details string `json:"details"`
}
