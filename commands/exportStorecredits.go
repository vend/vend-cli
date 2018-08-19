package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

// Command config
var exportStorecreditsCmd = &cobra.Command{
	Use:   "export-storecredits",
	Short: "Export Store Credits",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-storecredits -d DOMAINPREFIX -t TOKEN")),

	Run: func(cmd *cobra.Command, args []string) {
		getStoreCredits()
	},
}

func init() {
	rootCmd.AddCommand(exportStorecreditsCmd)
}

// Run executes the process of exporting Store Credits then writing them to CSV.
func getStoreCredits() {

	// Create new Vend Client
	vc := vend.NewClient(Token, DomainPrefix, "")

	// Get Store Credits
	fmt.Println("\nRetrieving Store Credits from Vend...")
	storeCredits, err := vc.StoreCredits()
	if err != nil {
		log.Fatalf("Failed while retrieving store credits: %v", err)
	}

	// Write Store Credits to CSV
	fmt.Println("Writing Store Credits to CSV file...")
	err = scWriterFile(storeCredits)
	if err != nil {
		log.Fatalf("Failed while writing Store Credits to CSV: %v", err)
	}

	fmt.Println(color.GreenString("\nExported %v Store Credits\n", len(storeCredits)))
}

// WriteFile writes Store Credits to CSV
func scWriterFile(sc []vend.StoreCredit) error {

	// Find Customer ID from Customer Code
	//	customerCode, err := getCustomerCode(sc.CustomerID)
	//	if err != nil {
	//		return fmt.Errorf("failed to get customer ID: %v", err)
	//	}
	//	sc.CustomerCode = &customerCode

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_storecredit_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		log.Fatalf("Failed to create CSV: %v", err)
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	// Write the header line.
	var header []string
	header = append(header, "id")
	header = append(header, "customer_id")
	header = append(header, "customer_code")
	header = append(header, "created_at")
	header = append(header, "balance")
	header = append(header, "total_issued")
	header = append(header, "total_redeemed")

	// Commit the header.
	writer.Write(header)

	// Now loop through each gift card object and populate the CSV.
	for _, storeCredits := range sc {

		var ID, customerID, customerCode, createdAt, balance, totalIssued, totalRedeemed string

		if storeCredits.ID != nil {
			ID = *storeCredits.ID
		}
		if storeCredits.CustomerID != nil {
			customerID = *storeCredits.CustomerID
		}
		if storeCredits.CustomerCode != nil {
			customerCode = *storeCredits.CustomerCode
		}
		if storeCredits.CreatedAt != nil {
			createdAt = *storeCredits.CreatedAt
		}
		if storeCredits.Balance != nil {
			balance = fmt.Sprintf("%v", *storeCredits.Balance)
		}
		if storeCredits.TotalIssued != nil {
			totalIssued = fmt.Sprintf("%v", *storeCredits.TotalIssued)
		}
		if storeCredits.TotalRedeemed != nil {
			totalRedeemed = fmt.Sprintf("%v", *storeCredits.TotalRedeemed)
		}

		var record []string
		record = append(record, ID)
		record = append(record, customerID)
		record = append(record, customerCode)
		record = append(record, createdAt)
		record = append(record, balance)
		record = append(record, totalIssued)
		record = append(record, totalRedeemed)
		writer.Write(record)
	}

	writer.Flush()
	return err
}

// func getCustomerCode() (string, error) {

// 	url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/customers/%v", DomainPrefix, &s.CustomerID)

// 	res, err := vendClient.MakeRequest("GET", url, nil)
// 	if err != nil {
// 		return "", err
// 	}

// 	c := &vend.CustomerSearchResponse{}

// 	err = json.Unmarshal(res, &c)
// 	if err != nil {
// 		fmt.Printf("Failed to Unmarshal JSON from Vend. Error: %v", err)
// 	}

// 	if len(c.Data) == 0 {
// 		return "", fmt.Errorf("no customers found for the supplied customer code")
// 	}

// 	return *c.Data[0].ID, nil
// }
