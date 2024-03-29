// fetches and houses tax data from the taxes and outlettaxes endpoints
package vend

import (
	"encoding/json"
	"fmt"
)

// outletTaxes houses data from the /api/2.0/outlet_taxes endpoint
type OutletTaxes struct {
	OutletID  *string `json:"outlet_id"`
	ProductID *string `json:"product_id"`
	TaxID     *string `json:"tax_id"`
	DeletedAt *string `json:"deleted_at"`
	Version   *int64  `json:"version"`
}

type Taxes struct {
	ID          *string    `json:"id"`
	Name        *string    `json:"name"`
	Version     *int64     `json:"version"`
	TaxRates    []TaxRates `json:"rates"`
	IsDefault   *bool      `json:"is_default"`
	DisplayName *string    `json:"display_name"`
}

type TaxRates struct {
	ID          *string  `json:"id"`
	Name        *string  `json:"name"`
	Rate        *float64 `json:"rate"`
	DisplayName *string  `json:"display_name"`
}

func (c *Client) Taxes() ([]Taxes, map[string]Taxes, error) {
	taxes := []Taxes{}
	page := []Taxes{}

	data, v, err := c.ResourcePage(0, "GET", "taxes")
	if err != nil {
		return taxes, nil, err
	}
	err = json.Unmarshal(data, &page)
	if err != nil {
		err = fmt.Errorf("error while unmarshalling: %s", err)
		return taxes, nil, err
	}

	taxes = append(taxes, page...)

	for len(page) > 0 {
		page = []Taxes{}
		data, v, err = c.ResourcePage(v, "GET", "taxes")
		if err != nil {
			return taxes, nil, err
		}
		err = json.Unmarshal(data, &page)
		if err != nil {
			err = fmt.Errorf("error while unmarshalling: %s", err)
			return taxes, nil, err
		}
		taxes = append(taxes, page...)

	}

	taxesMap := make(map[string]Taxes)
	for _, tax := range taxes {
		taxesMap[*tax.ID] = tax
	}

	return taxes, taxesMap, err

}

func (c *Client) OutletTaxes() ([]OutletTaxes, error) {
	outletTaxes := []OutletTaxes{}
	page := []OutletTaxes{}

	data, v, err := c.ResourcePage(0, "GET", "outlet_taxes")
	if err != nil {
		return outletTaxes, err
	}
	err = json.Unmarshal(data, &page)
	if err != nil {
		err = fmt.Errorf("error while unmarshalling: %s", err)
		return outletTaxes, err
	}

	outletTaxes = append(outletTaxes, page...)

	for len(page) > 0 {
		page = []OutletTaxes{}
		data, v, err = c.ResourcePage(v, "GET", "outlet_taxes")
		if err != nil {
			return outletTaxes, err
		}
		err = json.Unmarshal(data, &page)
		if err != nil {
			err = fmt.Errorf("error while unmarshalling: %s", err)
			return outletTaxes, err
		}
		outletTaxes = append(outletTaxes, page...)

	}

	return outletTaxes, err

}
