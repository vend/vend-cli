package cmd

import (
	"fmt"
	"time"

	"github.com/vend/vend-cli/pkg/messenger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
	csvparser "github.com/vend/vend-cli/pkg/csvparser"
	pbar "github.com/vend/vend-cli/pkg/progressbar"
)

type FailedVoidRequest struct {
	SaleID string
	Reason string
}

// voidSaleCmd represents the voidSale command
var voidSaleCmd = &cobra.Command{
	Use:   "void-sales",
	Short: "void-sales",
	Long: fmt.Sprintf(`
This tool requires a CSV of Sale IDs, no headers.

Example:
%s`, color.GreenString("vendcli void-sales -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),
	Run: func(cmd *cobra.Command, args []string) {
		voidSales()
	},
}

func init() {
	// Flag
	voidSaleCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	voidSaleCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(voidSaleCmd)
}

func voidSales() {

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

	failedRequests := []FailedVoidRequest{}

	// Make the requests
	fmt.Println("\nVoiding sales...")
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(ids), "Voiding Sales")
	if err != nil {
		fmt.Println("Error creating progress bar:", err)
	}
	for _, id := range ids {

		sale, err := getSaleRaw(id)
		bar.Increment()
		if err != nil {
			failedRequests = append(failedRequests, FailedVoidRequest{SaleID: id, Reason: err.Error()})
			continue
		}
		if _, ok := sale["status"]; ok {
			sale["status"] = "VOIDED"
		} else {
			failedRequests = append(failedRequests, FailedVoidRequest{SaleID: id, Reason: "Sale is malformed and does not have a status field."})
			continue
		}

		//Make the request
		url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales", DomainPrefix)
		_, err = vendClient.MakeRequest("POST", url, sale)
		if err != nil {
			failedRequests = append(failedRequests, FailedVoidRequest{SaleID: id, Reason: err.Error()})
			continue
		}
	}
	p.Wait()

	if len(failedRequests) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		saveFailedVoidRequestsToCSV(failedRequests)
	}

	fmt.Println(color.GreenString("\n\nFinished! 🎉\n"))

}

func saveFailedVoidRequestsToCSV(failedRequests []FailedVoidRequest) {

	fileName := fmt.Sprintf("%s_failed_void_requests__%v.csv", DomainPrefix, time.Now().Unix())
	err := csvparser.WriteErrorCSV(fileName, failedRequests)
	if err != nil {
		messenger.ExitWithError(err)
		return
	}
}
