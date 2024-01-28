package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vend/vend-cli/pkg/messenger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
	"github.com/wallclockbuilder/stringutil"
)

type UpdateSaleInvoiceRequest struct {
	SaleID           string
	NewInvoiceNumber string
}

// Command config
var updateSaleInvoiceCmd = &cobra.Command{
	Use:   "update-sale-invoice-number",
	Short: "Update the invoice number for a list of sales",
	Long: fmt.Sprintf(`
Updates the invoice number for a given list of sales.

%s This command is very dangerous and can result in duplicate sales, missing data or other unintended behavior. 
If you are unsure of its use, %s and seek help from a Product Specialist


csv should be in the following format
+-------------+--------------------------+
|   sale_id   |   invoice_number         |
+-------------+--------------------------+
| <sale UUID> | <desired invoice number> |
+-------------+--------------------------+

** Also prints a "post in case of emergency" log file 
** Keep this file, in case an issue occurs we can use this to recover data

Example: %s
`,

		color.RedString("WARNING:"),
		color.RedString("do not use this"),
		color.GreenString("vendcli update-sale-invoice-number -t TOKEN -d DOMAINPREFIX -f PATH/TO/FILE")),

	Run: func(cmd *cobra.Command, args []string) {
		updateSaleInvoice()
	},
}

func init() {
	// Flag
	updateSaleInvoiceCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	updateSaleInvoiceCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(updateSaleInvoiceCmd)

}

func updateSaleInvoice() {

	// Create Log File
	logFileName := fmt.Sprintf("%s_post_in_case_of_emergency_%v.txt", DomainPrefix, time.Now().Unix())
	logFile, err := os.OpenFile(fmt.Sprintf("./%s", logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		messenger.ExitWithError(err)
	}
	log.SetOutput(logFile)

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Printf("\nReading CSV...\n")
	saleList, err := ReadSaleInvoiceCSV(FilePath)
	if err != nil {
		fmt.Printf(color.RedString("Failed to get ids from the file: %s", FilePath))
	}

	// loop through entities, fetch the data from vend, swap sale_id, and post to vend
	for _, saleRequest := range saleList {
		fmt.Printf("Updating %s\n", saleRequest.SaleID)

		sale, err := getSaleRaw(saleRequest.SaleID)
		if err != nil {
			fmt.Printf("Could not fetch %s.\n", saleRequest.SaleID)
			fmt.Printf("%sSkipping...\n\n", color.RedString("ERROR: %s\n", err))
			continue
		}

		fmt.Println("old invoice number: ", sale["invoice_number"])
		sale["invoice_number"] = saleRequest.NewInvoiceNumber
		fmt.Println("new invoice number: ", sale["invoice_number"])
		fmt.Println("posting...")

		// Get Sale
		// sale, err := getSale9(saleRequest.SaleID)
		// // If there's an error getting sale, skip to the next one
		// if err != nil {
		// 	fmt.Printf("%s", color.RedString(" ERROR %s\n", err))
		// 	continue
		// }

		// Change User on the Sale
		//sale = changeUser(sale, saleRequest.UserID)

		//Make the request
		var res []byte
		url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales", DomainPrefix)
		res, err = vendClient.MakeRequest("POST", url, sale)

		// Check if there is an error in the response; if so display it
		// TODO: We should be using response code for this, but that's not being returned by MakeRequest
		responseStr := string(res)
		response := vend.Error9{}
		_ = json.Unmarshal(res, &response)

		//uncomment below line for debugging
		//fmt.Println(responseStr)

		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
		} else if len(response.Error) > 0 {
			fmt.Printf("%s", color.RedString(" %s\n", response.Error))
			if len(response.Details) > 0 {
				fmt.Printf("%s", color.RedString(" %s\n", response.Details))
			}
		} else if strings.HasPrefix(responseStr, "<!DOCTYPE html>") {
			fmt.Printf("%s - got robot error; unknown reason \n", color.RedString(" ERROR"))
		} else {
			fmt.Printf("%s", color.GreenString("UPDATED\n"))
		}
		fmt.Println()
	}
}

func getSaleRaw(id string) (map[string]interface{}, error) {

	var saleResponse map[string][]json.RawMessage
	var sale map[string]interface{}

	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales/%s", DomainPrefix, id)

	// Make the request
	res, err := vendClient.MakeRequest("GET", url, nil)

	// log the sale info for later in case we need to recover it
	log.Println(id, string(res))

	// Unmarshal JSON Response
	err = json.Unmarshal(res, &saleResponse)
	if err != nil {
		return sale, err
	}

	// check the data is valid
	data, ok := saleResponse["register_sales"]
	if !ok {
		return sale, fmt.Errorf("unable to parse sale response")
	} else if len(data) < 1 {
		return sale, fmt.Errorf("sale not found. check that your sale_id is valid")
	}

	err = json.Unmarshal(saleResponse["register_sales"][0], &sale)
	if err != nil {
		return sale, err
	}

	return sale, nil
}

// getSale9 pulls sales information from the 0.9 Vend API
// func getSale9(id string) (vend.Sale9, error) {

// 	sale := vend.Sale9{}
// 	saleResponse := vend.RegisterSale9{}

// 	// Create the Vend URL
// 	url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales/%s", DomainPrefix, id)

// 	// Make the request
// 	res, err := vendClient.MakeRequest("GET", url, nil)

// 	// log the sale info for later in case we need to recover it
// 	log.Println(id, string(res))

// 	// Unmarshal JSON Response
// 	err = json.Unmarshal(res, &saleResponse)
// 	if err != nil {
// 		return sale, err
// 	}

// 	// return sale info if results, otherwise return err
// 	if len(saleResponse.RegisterSale9) > 0 {
// 		return saleResponse.RegisterSale9[0], nil
// 	} else {
// 		err = errors.New("failed to GET sale info from vend")
// 		return sale, err
// 	}
// }

// Change

// ReadSaleUserCSV reads the provided CSV file and stores the input as UpdateSaleInvoiceRequest structs.
func ReadSaleInvoiceCSV(FilePath string) ([]UpdateSaleInvoiceRequest, error) {
	header := []string{"sale_id", "invoice_number"}

	// Open our provided CSV file.
	file, err := os.Open(FilePath)
	if err != nil {
		errorMsg := `error opening csv file - please check you've specified the right file

Tip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`
		fmt.Println(errorMsg, "\n")
		return []UpdateSaleInvoiceRequest{}, err
	}
	// Make sure to close at end.
	defer file.Close()

	// Create CSV reader on our file.
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, err := reader.Read()
	if err != nil {
		fmt.Printf("Failed to read headerow.")
		return []UpdateSaleInvoiceRequest{}, err
	}

	if len(headerRow) > 2 {
		fmt.Printf("Header row longer than expected")
	}

	// Check each string in the header row is same as our template.
	for i, row := range headerRow {
		if stringutil.Strip(strings.ToLower(row)) != header[i] {
			fmt.Println("Mismatched CSV headers, expecting {sale_id, invoice_number}")
			return []UpdateSaleInvoiceRequest{}, fmt.Errorf("Mistmatched Headers %v", err)
		}
	}

	// Read the rest of the data from the CSV.
	rawData, err := reader.ReadAll()
	if err != nil {
		return []UpdateSaleInvoiceRequest{}, err
	}

	var request UpdateSaleInvoiceRequest
	var requests []UpdateSaleInvoiceRequest
	var rowNumber int

	// Loop through rows and assign them to UpdateSaleInvoiceRequest struct.
	for _, row := range rawData {
		rowNumber++
		request, err = readSaleInvoiceRow(row)
		if err != nil {
			fmt.Println("Error reading row from CSV")
			continue
		}

		// Append each UpdateSaleInvoiceRequest struct to our list.
		requests = append(requests, request)
	}

	// Check how many rows we successfully read and stored.
	if len(requests) > 0 {
	} else {
		return []UpdateSaleInvoiceRequest{}, fmt.Errorf("No valid rows in csv")
	}

	return requests, err
}

// Read a single row of a CSV file and check for errors.
func readSaleInvoiceRow(row []string) (UpdateSaleInvoiceRequest, error) {
	var request UpdateSaleInvoiceRequest

	request.SaleID = row[0]
	request.NewInvoiceNumber = row[1]

	for i := range row {
		if len(row[i]) < 1 {
			err := errors.New("Missing field")
			return request, err
		}
	}
	return request, nil
}
