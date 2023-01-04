package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// Command config
var (
	submitMode string

	updateStorecreditCmd = &cobra.Command{
		Use:   "update-storecredits",
		Short: "Update Store Credits",
		Long: fmt.Sprintf(`
This tool requires the Store Credit CSV template, you can download it here: http://bit.ly/vendclitemplates

Example:
%s`, color.GreenString("vendcli update-storecredits -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),

		Run: func(cmd *cobra.Command, args []string) {
			updateStoreCredit()
		},
	}
)

func init() {
	// Flag
	updateStorecreditCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	updateStorecreditCmd.MarkFlagRequired("Filename")

	updateStorecreditCmd.Flags().StringVarP(&submitMode, "mode", "m", "replace", "Submission Mode: Options ('replace', 'adjust')")

	rootCmd.AddCommand(updateStorecreditCmd)
}

func updateStoreCredit() {

	// Test mode option is set correctly
	submitMode = strings.ToLower(submitMode)
	if !(submitMode == "adjust" || submitMode == "replace") {
		log.Fatalf("'%s' is not a valid option for -m mode. Mode should be 'adjust' or replace'", submitMode)
	}

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Read rows from CSV file and store them into structs
	// Check if we've got customer codes in the csv
	fmt.Println("\nReading Store Credits CSV...")
	csvRows, usesCustomerCodes, err := readStoreCreditCSV(FilePath, submitMode)
	if err != nil {
		log.Fatalf(color.RedString("Couldnt read Store Credits CSV file,  %s", err))
	}

	// if there are submitted customer_codes, convert them to customer_id
	if usesCustomerCodes {
		fmt.Println("CSV contains customer codes, setting customer ids..")
		csvRows, err = getCustomerIDs(vc, csvRows)
		if err != nil {
			log.Fatalf("not able to map customer ids to codes", err)
		}
	}

	// if mode is replace the transaction amount should be newBalance - currentBalance
	if submitMode == "replace" {
		updateAmounts(vc, csvRows)
	}

	//Get primary admin
	primaryAdmin := getPrimaryAdmin(vc)

	// make the bodys for our requests from the csvRows
	transactions := makeTransactions(csvRows, primaryAdmin)

	// Post Store Credits to Vend
	var numPosted int
	numTransactions := len(transactions)
	fmt.Printf("%v Store Credits to post\n\n", numTransactions)
	for _, transaction := range transactions {
		if postStoreCredit(transaction) {
			numPosted++
		}
	}

	fmt.Println(color.GreenString("\nFinished! Succesfully Posted %s of %s Store Credits 🎉\n",
		strconv.Itoa(numPosted), strconv.Itoa(numTransactions)))
}

// Read passed CSV, returns a slice of Store Credits
func readStoreCreditCSV(filePath string, submitMode string) ([]vend.StoreCreditCsv, bool, error) {

	// Open our provided CSV file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Could not read from CSV file\n", err)
	}
	// Make sure to close at end
	defer file.Close()

	// Create CSV reader on our file
	reader := csv.NewReader(file)

	// Read and store our header line.
	headerRow, err := reader.Read()
	if err != nil {
		log.Fatal("Error reading header row")
	}

	// Check each header in the row is same as our template. Fail if not
	checkHeaders(submitMode, headerRow)

	// Read the rest of the data from the CSV
	rawData, err := reader.ReadAll()
	if err != nil {
		log.Fatal("Error reading data from csv", err)
	}

	var csvStructs []vend.StoreCreditCsv
	usesCustomerCodes := false

	// Loop through rows and assign them to a StoreCreditCsv struct.
	for idx, row := range rawData {

		amount, err := strconv.ParseFloat(row[2], 10)
		if err != nil {
			fmt.Printf("%s, \nError:%s - invalid amount \nSkipping...", color.RedString("Error in row %s", strconv.Itoa(idx+2)), err)
			continue
		}

		rowStruct := vend.StoreCreditCsv{
			CustomerID:   &row[0],
			CustomerCode: &row[1],
			Amount:       &amount,
		}

		// make sure there is at least a customer id or customer code present
		if *rowStruct.CustomerCode != "" {
			usesCustomerCodes = true
		} else if *rowStruct.CustomerID == "" {
			fmt.Printf("%s: You must have at least customer_id or customer_code \nSkipping row...\n\n", color.RedString("Error in row %s", strconv.Itoa(idx+2)))
			continue
		}

		// Append struct into our array of structs
		csvStructs = append(csvStructs, rowStruct)
	}

	return csvStructs, usesCustomerCodes, err
}

// Post each Store Credits to Vend
func postStoreCredit(storeCredit vend.StoreCreditTransaction) bool {

	// Posting Store Credits to Vend
	fmt.Printf("Posting: %s / %v \n", storeCredit.CustomerID, storeCredit.Amount)
	if postTransaction(storeCredit) {
		return true
	} else {
		return false
	}
}

func postTransaction(trans vend.StoreCreditTransaction) bool {

	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/store_credits/%s/transactions", DomainPrefix, trans.CustomerID)

	// Make the Request
	resBody, err := vendClient.MakeRequest("POST", url, trans)
	if err != nil {
		fmt.Println(color.RedString("\nFailed to post store credit transaction for customer: %s", trans.CustomerID), "\nServer Response:\n", string(resBody), "\n")
		return false
	}

	return true
}

func getPrimaryAdmin(vc vend.Client) string {

	// Get Users.
	users, err := vc.Users()
	if err != nil {
		log.Fatalf("Failed retrieving Users from Vend %v", err)
	}

	var primaryAdmin string
	for _, user := range users {
		if *user.IsPrimaryUser {
			primaryAdmin = *user.ID
			break
		}

	}
	return primaryAdmin
}

// checkHeaders makes sure the submitted CSV headers matches our desired format
func checkHeaders(submitMode string, headerRow []string) {
	headers := [3]string{"customer_id", "customer_code", ""}
	if submitMode == "replace" {
		headers[2] = "new_balance"
	} else if submitMode == "adjust" {
		headers[2] = "amount"
	}

	for i := range headerRow {
		if headerRow[i] != headers[i] {
			fmt.Println(color.RedString("Found error in header rows."))
			log.Fatalf("\n\n 🛑 Looks like we have a mismatch in headers, this mode (%s) needs three headers: customer_id, customer_code, %s \n No header match for: %s instead got: %s \n\n",
				submitMode, string(headers[2]), string(headers[i]), string(headerRow[i]))
		}
	}

}

// getCustomerIDs finds the associated customer_ids for a given customer_code
func getCustomerIDs(vc vend.Client, csvRows []vend.StoreCreditCsv) ([]vend.StoreCreditCsv, error) {

	// Get Customers
	customers, err := vc.Customers()
	if err != nil {
		log.Fatalf("Failed retrieving customers from Vend %v", err)
	}

	// Build Customer Map
	customerMap := vend.CustomerMap(customers)

	// Loop through our csv row structs and attach a customer_id if one doesn't exist,
	// if succesful include that in an updated struct array
	var updatedCSVRows []vend.StoreCreditCsv
	for idx, rowStruct := range csvRows {
		// skip if we've already got a customer_id
		if len(*rowStruct.CustomerID) > 0 {
			updatedCSVRows = append(updatedCSVRows, rowStruct)
			continue
		} else {
			// check the customer_code has a valid map to a customer_id
			// if so, set the customer_id
			if customerID, ok := customerMap[*rowStruct.CustomerCode]; ok {
				*rowStruct.CustomerID = customerID
				updatedCSVRows = append(updatedCSVRows, rowStruct)
			} else {
				fmt.Println(color.RedString("The customer code '%s' in row %s does not seem to be valid, skipping..",
					*rowStruct.CustomerCode, strconv.Itoa(idx+2)))
			}
		}
	}

	return updatedCSVRows, err
}

func updateAmounts(vc vend.Client, csvRows []vend.StoreCreditCsv) error {

	// get current balances
	storeCredits, err := vc.StoreCredits()
	if err != nil {
		log.Fatalf("Failed while retrieving store credits: %v", err)
	}

	// make customer id -> customer balance map
	creditMap := vend.CreditMap(storeCredits)

	// loop through structs and update the amount
	for _, rowStruct := range csvRows {

		if currentBalance, ok := creditMap[*rowStruct.CustomerID]; ok {
			*rowStruct.Amount = *rowStruct.Amount - currentBalance

		}

	}

	return err
}

// makeTransactions takes the info from csvStructs and makes bodys for our POST requests
func makeTransactions(csvRows []vend.StoreCreditCsv, primaryAdmin string) []vend.StoreCreditTransaction {

	var transactions []vend.StoreCreditTransaction

	for _, rowStruct := range csvRows {

		var transType string
		if *rowStruct.Amount < 0 {
			transType = "REDEMPTION"
		} else {
			transType = "ISSUE"
		}

		clientID := uuid.New().String()

		transaction := vend.StoreCreditTransaction{
			CustomerID: *rowStruct.CustomerID,
			Amount:     *rowStruct.Amount,
			Type:       transType,
			ClientID:   &clientID,
			UserID:     &primaryAdmin,
		}
		transactions = append(transactions, transaction)
	}

	return transactions
}
