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

// Command config
var updateSaleIDcmd = &cobra.Command{
	Use:   "update-sale-user-id",
	Short: "Update the user id for a list of sales",
	Long: fmt.Sprintf(`
Updates the user for a given list of sales.

%s This command is very dangerous and can result in duplicate sales, missing data or other unintended behavior. 
If you are unsure of its use, %s and seek help from a Product Specialist


csv should be in the following format
+-------------+-------------+
|   sale_id   |   user_id   |
+-------------+-------------+
| <sale UUID> | <user UUID> |
+-------------+-------------+

** Also prints a "post in case of emergency" log file 
** Keep this file, in case an issue occurs we can use this to recover data

Example: %s
`,

		color.RedString("WARNING:"),
		color.RedString("do not use this"),
		color.GreenString("vendcli update-sale-user-id -t TOKEN -d DOMAINPREFIX -f PATH/TO/FILE")),

	Run: func(cmd *cobra.Command, args []string) {
		updateSaleID()
	},
}

func init() {
	// Flag
	updateSaleIDcmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	updateSaleIDcmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(updateSaleIDcmd)

}

func updateSaleID() {

	// Create Log File
	logFileName := fmt.Sprintf("%s_post_in_case_of_emergency_%v.txt", DomainPrefix, time.Now().Unix())
	logFile, err := os.OpenFile(fmt.Sprintf("./%s", logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		err = fmt.Errorf("error:%s", err)
		messenger.ExitWithError(err)
	}
	log.SetOutput(logFile)

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Printf("\nReading CSV...\n")
	saleList, err := ReadSaleUserCSV(FilePath)
	if err != nil {
		fmt.Printf(color.RedString("Failed to get ids from the file: %s", FilePath))
	}

	// loop through entities, fetch the data from vend, swap sale_id, and post to vend
	for _, saleRequest := range saleList {
		fmt.Printf("Updating %s:", saleRequest.SaleID)

		// Get Sale
		sale, err := getSale9(saleRequest.SaleID)
		// If there's an error getting sale, skip to the next one
		if err != nil {
			fmt.Printf("%s", color.RedString(" ERROR %s\n", err))
			continue
		}

		// Change User on the Sale
		sale = changeUser(sale, saleRequest.UserID)

		// Make the request
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
			fmt.Printf("Error maing request: %v\n", err)
		} else if len(response.Error) > 0 {
			fmt.Printf("%s", color.RedString(" %s\n", response.Error))
			if len(response.Details) > 0 {
				fmt.Printf("%s", color.RedString(" %s\n", response.Details))
			}
		} else if strings.HasPrefix(responseStr, "<!DOCTYPE html>") {
			fmt.Printf("%s - Check if user_id is correct\n", color.RedString(" ERROR"))
		} else {
			fmt.Printf("%s", color.GreenString(" UPDATED\n"))
		}
	}
}

// getSale9 pulls sales information from the 0.9 Vend API
func getSale9(id string) (vend.Sale9, error) {

	sale := vend.Sale9{}
	saleResponse := vend.RegisterSale9{}

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

	// return sale info if results, otherwise return err
	if len(saleResponse.RegisterSale9) > 0 {
		return saleResponse.RegisterSale9[0], nil
	} else {
		err = errors.New("failed to GET sale info from vend")
		return sale, err
	}
}

// Swap exisiting user for new desired user
func changeUser(sale vend.Sale9, userID string) vend.Sale9 {

	*sale.UserID = userID

	// also set any salesperson ids to new desired user
	if len(sale.RegisterSaleProducts) > 0 {
		for _, product := range sale.RegisterSaleProducts {
			for _, attribute := range product.Attributes {
				if attribute.Name != nil && *attribute.Name == "salesperson_id" && attribute.Value != nil {
					*attribute.Value = userID
				}
			}
		}
	}

	return sale
}

// ReadSaleUserCSV reads the provided CSV file and stores the input as sale structs.
func ReadSaleUserCSV(FilePath string) ([]vend.SaleUserUpload, error) {
	header := []string{"sale_id", "user_id"}

	// Open our provided CSV file.
	file, err := os.Open(FilePath)
	if err != nil {
		errorMsg := `error opening csv file - please check you've specified the right file

Tip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`
		fmt.Println(errorMsg, "\n")
		return []vend.SaleUserUpload{}, err
	}
	// Make sure to close at end.
	defer file.Close()

	// Create CSV reader on our file.
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, err := reader.Read()
	if err != nil {
		fmt.Printf("Failed to read headerow.")
		return []vend.SaleUserUpload{}, err
	}

	if len(headerRow) > 2 {
		fmt.Printf("Header row longer than expected")
	}

	// Check each string in the header row is same as our template.
	for i, row := range headerRow {
		if stringutil.Strip(strings.ToLower(row)) != header[i] {
			fmt.Println("Mismatched CSV headers, expecting {sale_id, user_id}")
			return []vend.SaleUserUpload{}, fmt.Errorf("Mistmatched Headers %v", err)
		}
	}

	// Read the rest of the data from the CSV.
	rawData, err := reader.ReadAll()
	if err != nil {
		return []vend.SaleUserUpload{}, err
	}

	var sale vend.SaleUserUpload
	var saleList []vend.SaleUserUpload
	var rowNumber int

	// Loop through rows and assign them to saleUserUpload struct.
	for _, row := range rawData {
		rowNumber++
		sale, err = readSaleUserRow(row)
		if err != nil {
			fmt.Println("Error reading row from CSV")
			continue
		}

		// Append each saleUserUpload struct to our list.
		saleList = append(saleList, sale)
	}

	// Check how many rows we successfully read and stored.
	if len(saleList) > 0 {
	} else {
		fmt.Println("No valid sales found")
	}

	return saleList, err
}

// Read a single row of a CSV file and check for errors.
func readSaleUserRow(row []string) (vend.SaleUserUpload, error) {
	var sale vend.SaleUserUpload

	sale.SaleID = row[0]
	sale.UserID = row[1]

	for i := range row {
		if len(row[i]) < 1 {
			err := errors.New("Missing field")
			return sale, err
		}
	}
	return sale, nil
}
