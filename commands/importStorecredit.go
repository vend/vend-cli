package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// Command config
var importStorecreditCmd = &cobra.Command{
	Use:   "import-storecredits",
	Short: "Import Store Credits",
	Long: fmt.Sprintf(`
This tool requires the Store Credit CSV template, you can download it here: http://bit.ly/vendclitemplates

Example:
%s`, color.GreenString("vendcli import-storecredits -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),

	Run: func(cmd *cobra.Command, args []string) {
		importStoreCredit()
	},
}

func init() {
	// Flag
	importStorecreditCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	importStorecreditCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(importStorecreditCmd)
}

func importStoreCredit() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Read Store Credits from CSV file
	fmt.Println("\nReading Store Credits CSV...")
	storeCredits, err := readStoreCreditCSV(FilePath)
	if err != nil {
		log.Fatalf(color.RedString("Couldnt read Store Credits CSV file,  %s", err))
	}

	// Post Store Credits to Vend
	fmt.Printf("%v Store Credits to post.\n\n", len(storeCredits))
	for _, sc := range storeCredits {
		err = postStoreCredit(sc)
	}

	fmt.Println(color.GreenString("\nFinished! 🎉\n"))
}

// Read passed CSV, returns a slice of Store Credits
func readStoreCreditCSV(filePath string) ([]vend.StoreCreditTransaction, error) {

	headers := []string{"customer_id", "amount", "note"}

	// Open our provided CSV file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Could not read from CSV file")
		return nil, err
	}
	// Make sure to close at end
	defer file.Close()

	// Create CSV reader on our file
	reader := csv.NewReader(file)

	var storeCredits []vend.StoreCreditTransaction

	// Read and store our header line.
	headerRow, err := reader.Read()

	// Check each header in the row is same as our template.
	for i := range headerRow {
		if headerRow[i] != headers[i] {
			fmt.Println(color.RedString("Found error in header rows."))
			log.Fatalf("\n\n 🛑 Looks like we have a mismatch in headers, this command needs three headers: customer_id, amount, note \n No header match for: %s instead got: %s \n\n",
				string(headers[i]), string(headerRow[i]))
		}
	}

	// Read the rest of the data from the CSV
	rawData, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Loop through rows and assign them to the Store Credit type.
	for _, row := range rawData {
		amount, err := strconv.ParseFloat(row[1], 10)
		if err != nil {
			return nil, err
		}
		storeCredit := vend.StoreCreditTransaction{
			CustomerID: row[0],
			Amount:     amount,
			Type:       "ISSUE",
			Notes:      &row[2],
		}

		// Append Store Credit info
		storeCredits = append(storeCredits, storeCredit)
	}

	return storeCredits, err
}

// Post each Store Credits to Vend
func postStoreCredit(storeCredit vend.StoreCreditTransaction) error {

	// // Find Customer ID from Customer Code
	// customerID, err := getCustomerID(storeCredit.CustomerCode)
	// if err != nil {
	// 	return fmt.Errorf("failed to get customer ID: %v", err)
	// }
	// storeCredit.CustomerID = customerID

	// Posting Store Credits to Vend
	fmt.Printf("Posting: %s / %v \n", storeCredit.CustomerID, storeCredit.Amount)
	err := postTransaction(storeCredit)
	if err != nil {
		return fmt.Errorf(color.RedString("Failed to post store credit: %v", err))
	}

	return nil
}

func postTransaction(trans vend.StoreCreditTransaction) error {

	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/store_credits/%s/transactions", DomainPrefix, trans.CustomerID)

	// Make the Request
	_, err := vendClient.MakeRequest("POST", url, trans)
	if err != nil {
		return fmt.Errorf(color.RedString("Failed to post store credit transaction: %s", err))
	}

	return nil
}
