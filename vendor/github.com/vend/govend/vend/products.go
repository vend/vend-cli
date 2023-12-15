// Package vend handles interactions with the Vend API.
package vend

import (
	"encoding/json"
	"fmt"
	"log"
)

// Product is a basic product object
// TODO: There are a number of unused fields left over from the 0.9 API, need to remove them and test
type Product struct {
	ID                  *string            `json:"id"`
	SourceID            *string            `json:"source_id"`
	VariantSourceID     *string            `json:"source_variant_id"`
	Handle              *string            `json:"handle"`
	HasVariants         bool               `json:"has_variants"`
	VariantParentID     *string            `json:"variant_parent_id"`
	IsComposite         bool               `json:"is_composite"`
	VariantOptions      []VariantOptions   `json:"variant_options,omitempty"`
	VariantName         *string            `json:"variant_name,omitempty"`
	Active              bool               `json:"active"`
	Name                *string            `json:"name"`
	Description         *string            `json:"description"`
	Image               *string            `json:"image"`
	ImageURL            *string            `json:"image_url"`
	ImageLarge          *string            `json:"image_large"`
	Images              []Image            `json:"images"`
	SKU                 *string            `json:"sku"`
	SKUCodes            []SKUCodes         `json:"product_codes"`
	Tags                *string            `json:"tags"`
	Brand               Brand              `json:"brand"`
	ProductSuppliers    []ProductSuppliers `json:"product_suppliers"`
	SupplyPrice         *float64           `json:"supply_price"`
	LoyaltyAmount       *float64           `json:"loyalty_amount"`
	AccountCodePurchase *string            `json:"account_code_purchase"`
	AccountCodeSales    *string            `json:"account_code_sales"`
	Source              *string            `json:"source"`
	TrackInventory      bool               `json:"track_inventory"`
	PriceBookEntries    []PriceBookEntry   `json:"price_book_entries"`
	PriceExcludingTax   *float64           `json:"price_excluding_tax"`
	Type                Type               `json:"type"`
	Tax                 *float64           `json:"tax"`
	TaxID               *string            `json:"tax_id"`
	TaxRate             *float64           `json:"tax_rate"`
	TaxName             *string            `json:"tax_name"`
	Taxes               []Tax              `json:"taxes"`
	TagIDs              []*string          `json:"tag_ids"`
	Weight              *float64           `json:"weight"`
	WeightUnit          *string            `json:"weight_unit"`
	SizeUnit            *string            `json:"dimensions_unit"`
	Height              *float64           `json:"height"`
	Width               *float64           `json:"width"`
	Length              *float64           `json:"length"`
	CreatedAt           *string            `json:"created_at"`
	UpdatedAt           *string            `json:"updated_at"`
	DeletedAt           *string            `json:"deleted_at"`
	Version             *int64             `json:"version"`
}

type ProductPayload struct {
	Data Product `json:"data"`
}

// Variant Options houses options related to variants
type VariantOptions struct {
	ID    *string `json:"id"`
	Name  *string `json:"name"`
	Value *string `json:"value"`
}

// Image is the info contained in each Vend image object.
type Image struct {
	ID      *string `json:"id,omitempty"`
	URL     *string `json:"url,omitempty"`
	Version *int64  `json:"version"`
}

type ImageDetailsPayload struct {
	Data ImageDetails `json:"data"`
}
type ImageDetails struct {
	ID        *string `json:"id"`
	Version   *int64  `json:"version"`
	ProductID *string `json:"product_id"`
	Position  *int64  `json:"position"`
	Status    *string `json:"status"`
}

// SKUCodes houses list of skus for a given product
type SKUCodes struct {
	ID   *string `json:"id"`
	Type *string `json:"type"`
	Code *string `json:"code"`
}

// Brand houses options related to brands
type Brand struct {
	ID          *string     `json:"id"`
	Name        *string     `json:"name"`
	Description interface{} `json:"description"`
	DeletedAt   interface{} `json:"deleted_at"`
	Version     *int64      `json:"version"`
}

// ProductSupplier houses options for supplier
type ProductSuppliers struct {
	ID           *string  `json:"id"`
	ProductID    *string  `json:"product_id"`
	SupplierID   *string  `json:"supplier_id"`
	SupplierName *string  `json:"supplier_name"`
	Code         *string  `json:"code"`
	Price        *float64 `json:"price"`
}

// PriceBookEntry houses product pricing object
type PriceBookEntry struct {
	ID                             string  `json:"id"`
	ProductID                      string  `json:"product_id"`
	PriceBookID                    string  `json:"price_book_id"`
	PriceBookName                  string  `json:"price_book_name"`
	Type                           string  `json:"type"`
	OutletName                     string  `json:"outlet_name"`
	OutletID                       string  `json:"outlet_id"`
	CustomerGroupName              string  `json:"customer_group_name"`
	CustomerGroupID                string  `json:"customer_group_id"`
	Price                          float64 `json:"price"`
	LoyaltyAmount                  int64   `json:"loyalty_amount"`
	Tax                            float64 `json:"tax"`
	TaxID                          string  `json:"tax_id"`
	TaxRate                        float64 `json:"tax_rate"`
	TaxName                        string  `json:"tax_name"`
	DisplayRetailPriceTaxInclusive int64   `json:"display_retail_price_tax_inclusive"`
	MinUnits                       string  `json:"min_units"`
	MaxUnits                       string  `json:"max_units"`
	ValidFrom                      string  `json:"valid_from"`
	ValidTo                        string  `json:"valid_to"`
}

type Type struct {
	ID        *string      `json:"id"`
	Name      *string      `json:"name"`
	DeletedAt *interface{} `json:"deleted_at"`
	Version   *int64       `json:"version"`
}

// Tax houses product tax object
type Tax struct {
	OutletID string `json:"outlet_id"`
	TaxID    string `json:"tax_id"`
}

// ImageUpload holds data for Images
type ImageUpload struct {
	Data Data `json:"data,omitempty"`
}

// Data is the information for each image contained in the response.
type Data struct {
	ID        *string `json:"id,omitempty"`
	ProductID *string `json:"product_id,omitempty"`
	Position  *int64  `json:"position,omitempty"`
	Status    *string `json:"status,omitempty"`
	Version   *int64  `json:"version,omitempty"`
}

// ProductUpload contains the fields needed to post an image to a product in Vend.
type ProductUpload struct {
	ID       string `json:"id,omitempty"`
	Handle   string `json:"handle,omitempty"`
	SKU      string `json:"sku,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// houses summary data about a store's inventory and products
type CatalogStats struct {
	TotalInventory     int64
	CountStandard      int64
	CountParentVariant int64
	CountChildVariant  int64
	CountComposite     int64
	CountActive        int64
	CountInactive      int64
	MaxCountSuppliers  int64
}

// houses info for tags
type Tags struct {
	ID   *string `json:"id"`
	Name *string `json:"name"`
}

// Products grabs and collates all products in pages of 10,000.
func (c *Client) Products() ([]Product, map[string]Product, error) {

	productMap := make(map[string]Product)
	products := []Product{}
	page := []Product{}
	data := []byte{}
	var v int64

	// v is a version that is used to get products by page.
	data, v, err := c.ResourcePage(0, "GET", "products")
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Printf("error while unmarshalling: %s", err)
	}

	products = append(products, page...)

	// Use version to paginate through all pages
	for len(page) > 0 {
		page = []Product{}
		data, v, err = c.ResourcePage(v, "GET", "products")
		err = json.Unmarshal(data, &page)
		products = append(products, page...)
	}

	productMap = buildProductMap(products)

	return products, productMap, err
}

func buildProductMap(products []Product) map[string]Product {
	productMap := make(map[string]Product)

	for _, product := range products {
		productMap[*product.ID] = product
	}

	return productMap
}

func (c *Client) Tags() (map[string]string, error) {

	tags := []Tags{}
	page := []Tags{}

	data, v, err := c.ResourcePage(0, "GET", "tags")
	err = json.Unmarshal(data, &page)
	if err != nil {
		log.Printf("error while unmarshalling: %s", err)
	}

	tags = append(tags, page...)

	for len(page) > 0 {
		page = []Tags{}
		data, v, err = c.ResourcePage(v, "GET", "tags")
		err = json.Unmarshal(data, &page)
		tags = append(tags, page...)

	}

	tagsMap := make(map[string]string)
	for _, tag := range tags {
		tagsMap[*tag.ID] = *tag.Name
	}

	return tagsMap, err

}

func (c *Client) ProductImagesDetails(id string) (ImageDetails, error) {

	var payload ImageDetailsPayload

	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/product_images/%s", c.DomainPrefix, id)
	body, err := c.MakeRequest("GET", url, nil)
	if err != nil {
		return payload.Data, err
	}

	err = json.Unmarshal(body, &payload)
	if err != nil {
		return payload.Data, err
	}

	return payload.Data, nil

}
