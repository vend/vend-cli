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

type FailedDeleteRequest struct {
	ConsignmentID string
	Reason        string
}

// deleteConsignmentsCmd represents the deleteConsignments command
var deleteConsignmentsCmd = &cobra.Command{
	Use:   "delete-consignments",
	Short: "Delete Consignments",
	Long: fmt.Sprintf(`
This tool requires a CSV of Congsignment IDs, no headers.

Example:
%s`, color.GreenString("vendcli delete-consignments -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),
	Run: func(cmd *cobra.Command, args []string) {
		deleteConsignments()
	},
}

func init() {
	// Flag
	deleteConsignmentsCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	deleteConsignmentsCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(deleteConsignmentsCmd)
}

func deleteConsignments() {

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

	failedRequests := []FailedDeleteRequest{}

	// Make the requests
	fmt.Println("\nDeleting consignments...")
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(ids), "Deleting")
	if err != nil {
		fmt.Println("Error creating progress bar:", err)
	}

	count := 0
	for _, id := range ids {
		bar.Increment()
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/consignments/%s", DomainPrefix, id)
		_, err = vendClient.MakeRequest("DELETE", url, nil)
		if err != nil {
			failedRequests = append(failedRequests, FailedDeleteRequest{ConsignmentID: id, Reason: fmt.Sprintf("Failed to delete consignment: %v", err)})
			continue
		}
		count += 1

	}
	p.Wait()

	if len(failedRequests) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_delete_requests__%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedRequests)
		if err != nil {
			messenger.ExitWithError(err)
			return
		}

	}

	fmt.Printf(color.GreenString("\n\nFinished! 🎉\nDeleted %d out of %d consignments"), count, len(ids))
}
