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

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
	"github.com/wallclockbuilder/stringutil"
)

type UpdateSaleInvoiceRequest struct {
	SaleID           string
	NewInvoiceNumber string
}

type FailedUpdateSaleInvoiceRequests struct {
	SaleID           string
	NewInvoiceNumber string
	Reason           string
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
template: https://bit.ly/vendcli-csv-templates
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

var failedUpdateSaleInvoiceRequests []FailedUpdateSaleInvoiceRequests

func init() {
	// Flag
	updateSaleInvoiceCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	updateSaleInvoiceCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(updateSaleInvoiceCmd)

}

func updateSaleInvoice() {

	// Create Log File
	logFileName := fmt.Sprintf("%s_update_sale_invoice_number_original_sales_before_update_%v.txt", DomainPrefix, time.Now().Unix())
	logFile, err := os.OpenFile(fmt.Sprintf("./%s", logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		messenger.ExitWithError(err)
	}
	fmt.Println("\nSaving original sales to: ", color.YellowString(logFileName))
	fmt.Println("-- Keep this file, in case an issue occurs we can use this to recover data --")
	log.SetOutput(logFile)
	defer logFile.Close()

	fmt.Println("\n\nStarting Command Update Invoice Number..")
	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Printf("\nReading CSV...\n")
	saleList, err := ReadSaleInvoiceCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("failed to get ids from the file: %s, error: %w", FilePath, err)
		messenger.ExitWithError(err)
	}

	fmt.Println("\nUpdating Invoice Numbers...")
	succesfulPosts := fetchSaleAndUpdateInvoiceNumber(saleList)

	if len(failedUpdateSaleInvoiceRequests) > 0 {
		fmt.Println(color.RedString("\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_update_invoice_number_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedUpdateSaleInvoiceRequests)
		if err != nil {
			messenger.ExitWithError(err)
		}
	}
	fmt.Println(color.GreenString("\n\nFinished! 🎉\nSuccesfully adjusted %d of %d sales", succesfulPosts, len(saleList)))
}

func fetchSaleAndUpdateInvoiceNumber(saleRequests []UpdateSaleInvoiceRequest) int {
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(saleRequests), "Updating")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}
	var count int = 0
	for _, saleRequest := range saleRequests {
		bar.Increment()

		// get the sale
		sale, err := getSaleRaw(saleRequest.SaleID)
		if err != nil {
			err = fmt.Errorf("error getting sale info: %s", err)
			failedUpdateSaleInvoiceRequests = append(failedUpdateSaleInvoiceRequests,
				FailedUpdateSaleInvoiceRequests{
					SaleID:           saleRequest.SaleID,
					NewInvoiceNumber: saleRequest.NewInvoiceNumber,
					Reason:           err.Error(),
				})
			continue
		}

		// change the invoice number
		sale["invoice_number"] = saleRequest.NewInvoiceNumber

		//Make the request
		url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales", DomainPrefix)
		resp, err := vendClient.MakeRequest("POST", url, sale)
		if err != nil {
			err = fmt.Errorf("error making request to vend: %s response: %s", err, string(resp))
			failedUpdateSaleInvoiceRequests = append(failedUpdateSaleInvoiceRequests,
				FailedUpdateSaleInvoiceRequests{
					SaleID:           saleRequest.SaleID,
					NewInvoiceNumber: saleRequest.NewInvoiceNumber,
					Reason:           err.Error(),
				})
			continue
		}
		count += 1
	}
	p.Wait()
	return count
}

func getSaleRaw(id string) (map[string]interface{}, error) {

	var saleResponse map[string][]json.RawMessage
	var sale map[string]interface{}

	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales/%s", DomainPrefix, id)
	res, err := vendClient.MakeRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("error getting sale info: %s", err)
		return sale, err
	}
	// log the sale info for later in case we need to recover it
	log.Println(id, string(res))

	// Unmarshal JSON Response
	err = json.Unmarshal(res, &saleResponse)
	if err != nil {
		err = fmt.Errorf("error unmarshalling sale info: %s", err)
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

// ReadSaleUserCSV reads the provided CSV file and stores the input as UpdateSaleInvoiceRequest structs.
func ReadSaleInvoiceCSV(FilePath string) ([]UpdateSaleInvoiceRequest, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		err = fmt.Errorf("error creating progress bar:%s", err)
		return nil, err
	}
	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Open our provided CSV file.
	file, err := os.Open(FilePath)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		return []UpdateSaleInvoiceRequest{}, err
	}
	defer file.Close()
	reader := csv.NewReader(file)

	// Read and store our header line.
	header := []string{"sale_id", "invoice_number"}
	headerRow, err := reader.Read()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("failed to read header row: %v", err)
		return []UpdateSaleInvoiceRequest{}, err
	}

	if len(headerRow) > 2 {
		bar.AbortBar()
		p.Wait()
		err = errors.New("header row longer than expected")
		return []UpdateSaleInvoiceRequest{}, err
	}

	// Check each string in the header row is same as our template.
	for i, row := range headerRow {
		if stringutil.Strip(strings.ToLower(row)) != header[i] {
			bar.AbortBar()
			p.Wait()
			err = fmt.Errorf("mismatched CSV headers, expecting {sale_id, invoice_number}")
			return []UpdateSaleInvoiceRequest{}, err
		}
	}

	// Read the rest of the data from the CSV.
	rawData, err := reader.ReadAll()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return []UpdateSaleInvoiceRequest{}, err
	}

	var request UpdateSaleInvoiceRequest
	var requests []UpdateSaleInvoiceRequest
	for idx, row := range rawData {
		request, err = readSaleInvoiceRow(row)
		if err != nil {
			err = fmt.Errorf("error reading csv row %d: %s", idx+1, err) // +1 to make it 1-indexed
			failedUpdateSaleInvoiceRequests = append(failedUpdateSaleInvoiceRequests,
				FailedUpdateSaleInvoiceRequests{
					SaleID:           row[0],
					NewInvoiceNumber: row[1],
					Reason:           err.Error(),
				})
			continue
		}
		requests = append(requests, request)
	}
	// Check how many rows we successfully read and stored.
	if len(requests) > 0 {
	} else {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("no valid rows in csv")
		return []UpdateSaleInvoiceRequest{}, err
	}
	bar.SetIndeterminateBarComplete()
	p.Wait()

	return requests, err
}

// Read a single row of a CSV file and check for errors.
func readSaleInvoiceRow(row []string) (UpdateSaleInvoiceRequest, error) {
	var request UpdateSaleInvoiceRequest

	request.SaleID = row[0]
	request.NewInvoiceNumber = row[1]

	for i := range row {
		if len(row[i]) < 1 {
			err := errors.New("missing field")
			return request, err
		}
	}
	return request, nil
}
