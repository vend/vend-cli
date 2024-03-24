package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

type FailedLoyaltyAdjustment struct {
	CustomerID string
	Amount     string
	Reason     string
}

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

var failedLoyaltyAdjustments []FailedLoyaltyAdjustment

func loyaltyAdjustment() {

	// Read Loyalty Adjustemtns from CSV file
	fmt.Println("\nReading Loyalty Adjustment CSV...")
	loyaltyAdjustments, err := readLoyaltyAdjustmentCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("couldnt read Loyalty Adjustment CSV file: %s", err)
		messenger.ExitWithError(err)
	}

	// Posting Adjustments to Vend
	fmt.Println("\nPosting Loyalty Adjustments to Vend...")
	count := postLloyaltyAdjustments(loyaltyAdjustments)

	if len(failedLoyaltyAdjustments) > 0 {
		fmt.Println(color.RedString("\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_loyalty_adjustment_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedLoyaltyAdjustments)
		if err != nil {
			err = fmt.Errorf("couldnt write failed Loyalty Adjustments to CSV file: %s", err)
			messenger.ExitWithError(err)
		}
	}

	fmt.Println(color.GreenString("\n\nFinished! 🎉\nSuccesfully adjusted %d of %d Customer Loyalty Balances", count, len(loyaltyAdjustments)))

}

func postLloyaltyAdjustments(loyaltyAdjustments []vend.Customer) int {

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(loyaltyAdjustments), "Posting Adjustments")
	if err != nil {
		fmt.Println("Error creating progress bar:", err)
	}

	// Posting Adjustments to Vend
	var count int = 0
	for _, loyaltyAdjustment := range loyaltyAdjustments {
		bar.Increment()
		// Create the Vend URL
		url := fmt.Sprintf("https://%s.vendhq.com/api/customers", DomainPrefix)

		// Make the request to Vend
		vendClient := vend.NewClient(Token, DomainPrefix, "")
		_, err := vendClient.MakeRequest("POST", url, loyaltyAdjustment)
		if err != nil {
			err = fmt.Errorf("something went wrong trying to post loyalty: %s", err)
			failedLoyaltyAdjustments = append(failedLoyaltyAdjustments, FailedLoyaltyAdjustment{
				CustomerID: *loyaltyAdjustment.ID,
				Amount:     *loyaltyAdjustment.LoyaltyAdjustment,
				Reason:     err.Error(),
			})
			continue
		}
		count += 1
	}
	p.Wait()
	return count
}

// Read passed CSV, returns a slice of Loyalty Adjustments
func readLoyaltyAdjustmentCSV(filePath string) ([]vend.Customer, error) {

	headers := []string{"customer_id", "amount"}

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		err = fmt.Errorf("error creating progress bar:%s", err)
		return nil, err
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Open our provided CSV file
	file, err := os.Open(filePath)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		return nil, err
	}
	// Make sure to close at end
	defer file.Close()

	// Create CSV reader on our file
	reader := csv.NewReader(file)

	var loyaltyAdjustments []vend.Customer

	// Read and store our header line.
	headerRow, err := reader.Read()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return nil, err
	}

	// Check each header in the row is same as our template.
	for i := range headerRow {
		if headerRow[i] != headers[i] {
			bar.AbortBar()
			p.Wait()
			err = fmt.Errorf("found error in hearder rows. No header match for: %s Instead got: %s",
				string(headers[i]), string(headerRow[i]))
			messenger.ExitWithError(err)

		}
	}

	// Read the rest of the data from the CSV
	rawData, err := reader.ReadAll()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return nil, err
	}

	// Loop through rows and assign them to the Loyalty Adjustment type.
	for _, row := range rawData {
		// skip if its missing data
		if row[0] == "" || row[1] == "" {
			reason := "missing fields"
			failedLoyaltyAdjustments = append(failedLoyaltyAdjustments, FailedLoyaltyAdjustment{
				CustomerID: row[0],
				Amount:     row[1],
				Reason:     reason,
			})
			continue
		}

		loyaltyAdjustment := vend.Customer{
			ID:                &row[0],
			LoyaltyAdjustment: &row[1],
		}
		loyaltyAdjustments = append(loyaltyAdjustments, loyaltyAdjustment)
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

	return loyaltyAdjustments, err
}
