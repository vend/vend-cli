package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// loyaltyAdjustmentCmd represents the loyaltyAdjustment command
var loyaltyAdjustmentCmd = &cobra.Command{
	Use:   "loyalty-adjustment",
	Short: "Customer Loyalty Adjustment",
	Long: fmt.Sprintf(`
This tool requires the Customer Loyalty Adjustment CSV template, you can download it here: http://bit.ly/vendclitemplates

Example:
%s`, color.GreenString("vendcli loyalty-adjustment -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),

	Run: func(cmd *cobra.Command, args []string) {
		loyaltyAdjustment()
	},
}

func init() {
	// Flag
	loyaltyAdjustmentCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	loyaltyAdjustmentCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(loyaltyAdjustmentCmd)
}

func loyaltyAdjustment() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Read Loyalty Adjustemtns from CSV file
	fmt.Println("\nReading Loyalty Adjustment CSV...")
	loyaltyAdjustments, err := readLoyaltyAdjustmentCSV(FilePath)
	if err != nil {
		log.Fatalf("Couldnt read Loyalty Adjustment CSV file,  %s", err)
	}

	// Posting Adjustments to Vend
	fmt.Printf("%d Loyalty Adjustments to post.\n \n", len(loyaltyAdjustments))
	for _, loyaltyAdjustment := range loyaltyAdjustments {
		// Create the Vend URL
		url := fmt.Sprintf("https://%s.vendhq.com/api/customers", DomainPrefix)

		// Make the request to Vend
		_, err := vendClient.MakeRequest("POST", url, loyaltyAdjustment)
		if err != nil {
			fmt.Printf("Something went wrong trying to post supplier: %s", err)
		}

	}
	fmt.Printf("\nFinished! Succesfully adjusted %d Customer Loyalty Balances", len(loyaltyAdjustments))

}

// Read passed CSV, returns a slice of Loyalty Adjustments
func readLoyaltyAdjustmentCSV(filePath string) ([]vend.Customer, error) {

	headers := []string{"customer_id", "amount"}

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

	var loyaltyAdjustments []vend.Customer

	// Read and store our header line.
	headerRow, err := reader.Read()

	// Check each header in the row is same as our template.
	for i := range headerRow {
		if headerRow[i] != headers[i] {
			fmt.Println("Found error in header rows.")
			log.Fatalf("No header match for: %s Instead got: %s.",
				string(headers[i]), string(headerRow[i]))
		}
	}

	// Read the rest of the data from the CSV
	rawData, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Loop through rows and assign them to the Loyalty Adjustment type.
	for _, row := range rawData {
		loyaltyAdjustment := vend.Customer{
			ID:                &row[0],
			LoyaltyAdjustment: &row[1],
		}

		// Append Adjustments info
		loyaltyAdjustments = append(loyaltyAdjustments, loyaltyAdjustment)
	}

	return loyaltyAdjustments, err
}
