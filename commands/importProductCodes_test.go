package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractingProductCodes(t *testing.T) {
	expectedCodes := []ProductCodeAdd{
		toCodeAddAction("0242ac12-0002-11e9-e8c4-659494e33102", "EAN", "45345"),
		toCodeAddAction("0242ac12-0002-11e9-e8c4-659494e33102", "ITF", "555"),
		toCodeAddAction("0242ac12-0002-11e9-e8c4-659494e33103", "ISBN", "3423"),
		toCodeAddAction("0242ac12-0002-11e9-e8c4-659494e33103", "UPC", "5354353"),
		toCodeAddAction("0242ac12-0002-11e9-e8c4-659494e33103", "CUSTOM", "1000022"),
		toCodeAddAction("0242ac12-0002-11e9-e8c4-659494e33104", "JAN", "3442"),
	}

	productCodes, err := readProductCodesCSV("./testdata/import_product_codes.csv")
	assert.NoError(t, err)
	assert.Equal(t, expectedCodes, productCodes)
}

func TestProductCodesUniqueness(t *testing.T) {
	_, err := readProductCodesCSV("./testdata/import_duplicate_product_codes.csv")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "duplicate code: 3442")
}

func toCodeAddAction(productID, codeType, code string) ProductCodeAdd {
	return ProductCodeAdd{
		Action:    AddCodeAction,
		ProductID: productID,
		Data: ProductCode{
			Type: codeType,
			Code: code,
		},
	}
}
