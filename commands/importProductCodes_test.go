package cmd

import "testing"

func TestExtractingProductCodes(t *testing.T) {
	productCodes, err := readProductCodesCSV("./testdata/import_product_codes.csv")
	if err != nil {
		t.Error(err)
	}

	if len(productCodes) < 1 {
		t.Error("No codes")
	}
}
