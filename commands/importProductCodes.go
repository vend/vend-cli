package cmd

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
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

// ProductCodeAddErrors represents product codes that were not processed successfully with reasons.
type ProductCodeAddErrors struct {
	ProductCodes []ProductCodeAdd
	Reason       string
	Message      string
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
	fmt.Println("\nReading product codes CSV...")
	productCodes, err := readProductCodesCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("couldnt read Product Code CSV file, %s", err)
		messenger.ExitWithError(err)
	}

	// Post Product Codes to Vend
	fmt.Println("\nPosting product codes to Vend...")
	err = postProductCodes(productCodes)
	if err != nil {
		err = fmt.Errorf("failed to post product codes, %s", err)
		messenger.ExitWithError(err)
	}
}

// Read passed CSV, returns a slice of product codes add instructions.
func readProductCodesCSV(filePath string) ([]ProductCodeAdd, error) {
	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	header, records, err := loadRecordsFromCSV(filePath)
	if err != nil {
		bar.AbortBar()
		return nil, err
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Ensure valid header fields have been provided
	err = validateHeader(header)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		fmt.Println("Header validation failed")
		return nil, err
	}

	// Ensure there are no duplicate product codes
	err = validateProductCodeUniqueness(records)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("uniqueness validation failed! All product codes must be unique across the product catalogue. %w", err)
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

	bar.SetIndeterminateBarComplete()
	p.Wait()

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
	for idx, row := range records {
		// Start at column 1, column 0 is always product_id
		for c := 1; c < len(row); c++ {
			pCode := row[c]
			if pCode == "" {
				continue
			}
			if _, ok := codes[pCode]; ok {
				return fmt.Errorf("duplicate code: %s on row %v", pCode, idx+1) // csv is 1-based index
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
	totalProducts := len(productCodes)

	failedProductCodes := map[int]ProductCodeAddErrors{}
	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/products/actions/bulk", DomainPrefix)

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(totalProducts, "Writing CSV")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	for i := 0; i < totalProducts; i += BatchSize {
		bar.IncBy(BatchSize)
		batchNum := i + 1
		j := i + BatchSize
		if j > len(productCodes) {
			j = len(productCodes)
		}
		// Make the request to Vend
		statusCode, response, err := makeRequest("POST", url, productCodes[i:j])
		if err != nil {
			return fmt.Errorf("something went wrong trying to post product code: %s, %s", err, response)
		}

		switch {
		case statusCode == http.StatusUnprocessableEntity:
			failedProductCodes[batchNum] = ProductCodeAddErrors{
				productCodes[i:j],
				"Validation",
				response,
			}
		default:
			fmt.Println("Unknown error: ", response)
			failedProductCodes[batchNum] = ProductCodeAddErrors{
				productCodes[i:j],
				"Unknown",
				response,
			}
		}
	}
	p.Wait()

	// If any codes failed, export them
	if len(failedProductCodes) > 0 {
		filename, err := writeOutput(failedProductCodes)
		if err != nil {
			fmt.Printf("\nUnsuccesssful! Failed to write ouput for %d Product Codes", len(failedProductCodes))
			return err
		}
		fmt.Println(color.GreenString("\nFinished! 🎉"))
		fmt.Println(color.RedString("Partially successful, %d batches failed. Please check %s file for the failed batches.", len(failedProductCodes), filename))
	} else {
		fmt.Println(color.GreenString("\nFinished! 🎉 Succesfully created %d Product Codes", len(productCodes)))
	}

	return err
}

// writeOutput writes outcome of product code creation to csv
func writeOutput(failedCodes map[int]ProductCodeAddErrors) (string, error) {
	headers := []string{"product_id", "type", "code", "batch_number", "reason", "message"}
	var rows [][]string

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(failedCodes), "Writing CSV")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	for batchNum, failures := range failedCodes {
		bar.Increment()
		for _, c := range failures.ProductCodes {
			row := []string{c.ProductID, c.Data.Type, c.Data.Code, strconv.Itoa(batchNum), failures.Reason, failures.Message}
			rows = append(rows, row)
		}
	}
	p.Wait()

	fileName := "product_code_add_" + time.Now().Local().Format("20060102150405") + ".csv"
	return fileName, writeCSV(fileName, headers, rows)
}
