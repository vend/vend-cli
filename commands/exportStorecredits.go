package cmd

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
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

	// Get Store Credits
	fmt.Println("\nRetrieving Store Credits from Vend...")
	storeCredits := fetchDataForStoreCreditExport()

	// Write Store Credits to CSV
	fmt.Println("\nWriting Store Credits to CSV file...")
	err := scWriterFile(storeCredits)
	if err != nil {
		err = fmt.Errorf("failed while writing Store Credits to CSV: %v", err)
		messenger.ExitWithError(err)
	}

	fmt.Println(color.GreenString("\nExported %v Store Credits  🎉\n", len(storeCredits)))
}

func fetchDataForStoreCreditExport() []vend.StoreCredit {
	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("store credits")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	storeCredits, err := vc.StoreCredits()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("failed while retrieving store credits: %v", err)
		messenger.ExitWithError(err)
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

	return storeCredits
}

// WriteFile writes Store Credits to CSV
func scWriterFile(sc []vend.StoreCredit) error {

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(sc), "Writing CSV")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_storecredit_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("failed to create CSV: %v", err)
		messenger.ExitWithError(err)
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	// Write the header line.
	var header []string
	header = append(header, "id")
	header = append(header, "customer_id")
	// header = append(header, "customer_code")
	header = append(header, "created_at")
	header = append(header, "balance")
	header = append(header, "total_issued")
	header = append(header, "total_redeemed")

	// Commit the header.
	writer.Write(header)

	// Now loop through each gift card object and populate the CSV.
	for _, storeCredits := range sc {
		bar.Increment()

		var ID, customerID, createdAt, balance, totalIssued, totalRedeemed string

		if storeCredits.ID != nil {
			ID = *storeCredits.ID
		}
		if storeCredits.CustomerID != nil {
			customerID = *storeCredits.CustomerID
		}
		// if storeCredits.CustomerCode != nil {
		// 	customerCode = *storeCredits.CustomerCode
		// }
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
		// record = append(record, customerCode)
		record = append(record, createdAt)
		record = append(record, balance)
		record = append(record, totalIssued)
		record = append(record, totalRedeemed)
		writer.Write(record)
	}
	p.Wait()
	writer.Flush()
	return err
}
