package cmd

import (
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

const AddCodeAction = "product.code.add"

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
	header, records, err := loadRecordsFromFile(filePath)
	if err != nil {
		return nil, err
	}

	// Ensure valid header fields have been provided
	err = validateHeader(header)
	if err != nil {
		fmt.Println("Header validation failed")
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

func loadRecordsFromFile(path string) ([]string, [][]string, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Could not read from CSV file")
		return nil, nil, err
	}
	return readRecords(raw)
}

func readRecords(csvBytes []byte) ([]string, [][]string, error) {
	reader := csv.NewReader(bytes.NewReader(csvBytes))
	reader.FieldsPerRecord = -1
	header, err := reader.Read()
	if err != nil {
		return nil, nil, err
	}
	records, err := reader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	return header, records, nil
}

// Post product codes to Vend
func postProductCodes(productCodes []ProductCodeAdd) error {
	var err error

	// Posting product codes to Vend
	fmt.Printf("%d Product codes to post.\n \n", len(productCodes))
	for _, code := range productCodes {
		fmt.Printf("Posting: %v \n", code)
		// Create the Vend URL
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/products/actions/bulk", DomainPrefix)

		// Make the request to Vend
		res, err := vendClient.MakeRequest("POST", url, code)
		if err != nil {
			return fmt.Errorf("something went wrong trying to post product code: %s, %s", err, string(res))
		}

	}
	fmt.Printf("\nFinished! Succesfully created %d Product Codes", len(productCodes))

	return err
}
