package cmd

import (
	"fmt"
	"time"

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

type FailedCustomerDeleteRequest struct {
	CustomerID string
	Reason     string
}

// deleteCustomersCmd represents the deleteCustomers command
var deleteCustomersCmd = &cobra.Command{
	Use:   "delete-customers",
	Short: "Delete Customers",
	Long: fmt.Sprintf(`
This tool requires a CSV of Customer IDs, no headers.

Example:
%s`, color.GreenString("vendcli delete-customers -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),
	Run: func(cmd *cobra.Command, args []string) {
		deleteCustomers()
	},
}

func init() {
	// Flag
	deleteCustomersCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	deleteCustomersCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(deleteCustomersCmd)
}

func deleteCustomers() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Println("\nReading CSV...")
	ids, err := csvparser.ReadIdCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("Failed to get IDs from the file: %s\nError:%s", FilePath, err)
		messenger.ExitWithError(err)
	}

	failedRequests := []FailedCustomerDeleteRequest{}

	// Make the requests
	fmt.Println("\nDeleting customers...")
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(ids), "Deleting")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}
	for _, id := range ids {
		bar.Increment()
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/customers/%s", DomainPrefix, id)
		_, err = vendClient.MakeRequest("DELETE", url, nil)
		if err != nil {
			failedRequests = append(failedRequests, FailedCustomerDeleteRequest{CustomerID: id, Reason: err.Error()})
		}
	}
	p.Wait()

	if len(failedRequests) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		saveFailedCustomerDeleteRequestsToCSV(failedRequests)
	}

	fmt.Println(color.GreenString("\n\nFinished! 🎉\n"))

}

func saveFailedCustomerDeleteRequestsToCSV(failedRequests []FailedCustomerDeleteRequest) {

	fileName := fmt.Sprintf("%s_failed_delete_customer_requests__%v.csv", DomainPrefix, time.Now().Unix())
	err := csvparser.WriteErrorCSV(fileName, failedRequests)
	if err != nil {
		messenger.ExitWithError(err)
		return
	}
}
