package cmd

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

const (
	AddCodeAction = "product.code.add"
	BatchSize     = 99 // Bulk API limited to 100 actions per request
)

// ProductCodeAdd represents an intent to add a structured product code.
type ProductCodeAdd struct {
	Action    string      `json:"action"`
	ProductID string      `json:"product_id"`
	Data      ProductCode `json:"data"`
}
type ProductCode struct {
	Type string `json:"type"`
	Code string `json:"code"`
}

// Command config
var importProductCodesCmd = &cobra.Command{
	Use:   "import-product-codes",
	Short: "Import Product Codes",
	Long: fmt.Sprintf(`
This tool requires the Product Codes CSV template, you can download it here: http://bit.ly/vendclitemplates
Example:
%s`, color.GreenString("vendcli import-product-codes -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),

	Run: func(cmd *cobra.Command, args []string) {
		importProductCodes()
	},
}

func init() {
	// Flags
	importProductCodesCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	importProductCodesCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(importProductCodesCmd)
}

func importProductCodes() {
	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Read Product Codes from CSV file
	fmt.Println("\nReading Product Codes CSV...")
	productCodes, err := readProductCodesCSV(FilePath)
	if err != nil {
		log.Fatalf("Couldnt read Product Code CSV file, %s", err)
	}

	// Post Product Codes to Vend
	err = postProductCodes(productCodes)
	if err != nil {
		log.Fatalf("Failed to post product codes, %s", err)
	}

	fmt.Println(color.GreenString("\nFinished!\n"))
}

// Read passed CSV, returns a slice of product codes add instructions.
func readProductCodesCSV(filePath string) ([]ProductCodeAdd, error) {
	header, records, err := loadRecordsFromCSV(filePath)
	if err != nil {
		return nil, err
	}

	// Ensure valid header fields have been provided
	err = validateHeader(header)
	if err != nil {
		fmt.Println("Header validation failed")
		return nil, err
	}

	// Ensure there are no duplicate product codes
	err = validateProductCodeUniqueness(records)
	if err != nil {
		fmt.Println("Uniqueness validation failed: ", err.Error())
		return nil, err
	}

	var prodCodes []ProductCodeAdd

	for _, row := range records {
		productId := row[0]
		// Start at column 1, column 0 is always product_id
		for c := 1; c < len(row); c++ {
			pCode := row[c]
			// Only add codes where a value was provided.
			if pCode != "" {
				prodCodes = append(prodCodes, ProductCodeAdd{
					Action:    AddCodeAction,
					ProductID: productId,
					Data: ProductCode{
						Type: header[c],
						Code: pCode,
					},
				})
			}
		}
	}
	return prodCodes, err
}

func validateHeader(header []string) error {
	if len(header) < 2 {
		return errors.New("incomplete data, expecting at least one product code")
	}
	if header[0] != "product_id" {
		return errors.New("missing product_id column")
	}
	return nil
}

func validateProductCodeUniqueness(records [][]string) error {
	codes := make(map[string]interface{})
	for _, row := range records {
		// Start at column 1, column 0 is always product_id
		for c := 1; c < len(row); c++ {
			pCode := row[c]
			if pCode == "" {
				continue
			}
			if _, ok := codes[pCode]; ok {
				return errors.New("duplicate code: " + pCode)
			} else {
				codes[pCode] = nil
			}
		}
	}
	return nil
}

// Post product codes to Vend
func postProductCodes(productCodes []ProductCodeAdd) error {
	var err error
	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.works/api/2.0/products/actions/bulk", DomainPrefix)

	fmt.Println("Begin processing product codes.")

	for i := 0; i < len(productCodes); i += BatchSize {
		j := i + BatchSize
		if j > len(productCodes) {
			j = len(productCodes)
		}
		// Make the request to Vend
		fmt.Printf("Posting: %v \n", productCodes[i:j])
		statusCode, response, err := makeRequest("POST", url, productCodes[i:j])
		if err != nil {
			return fmt.Errorf("something went wrong trying to post product code: %s, %s", err, response)
		}

		switch statusCode {
		case http.StatusOK:
			fmt.Printf("\nBatch complete! Succesfully created %d Product Codes", len(productCodes[i:j]))
		case http.StatusUnprocessableEntity:
			fmt.Println("Validation error: ", response)
		default:
			fmt.Println("Unknown error: ", response)
		}
	}

	fmt.Printf("\nFinished! Succesfully created %d Product Codes", len(productCodes))

	return err
}
