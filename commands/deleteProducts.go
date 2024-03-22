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

type FailedDeleteProductRequest struct {
	ProductID string
	Reason    string
}

// deleteProductsCmd represents the deleteProducts command
var deleteProductsCmd = &cobra.Command{
	Use:   "delete-products",
	Short: "Delete Products",
	Long: fmt.Sprintf(`
This tool requires a CSV of Product IDs, no headers.

Example:
%s`, color.GreenString("vendcli delete-products -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),
	Run: func(cmd *cobra.Command, args []string) {
		deleteProducts()
	},
}

func init() {
	// Flag
	deleteProductsCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	deleteProductsCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(deleteProductsCmd)
}

func deleteProducts() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Println("\nReading CSV...")
	ids, err := csvparser.ReadIdCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("Failed to get ids from the file: %s\nError:%s", FilePath, err)
		messenger.ExitWithError(err)
	}

	failedRequests := []FailedDeleteProductRequest{}

	// Make the requests
	fmt.Println("\nDeleting products...")
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(ids), "Deleting Products")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	count := 0
	for _, id := range ids {
		bar.Increment()
		url := fmt.Sprintf("https://%s.vendhq.com/api/products/%s", DomainPrefix, id)
		_, err = vendClient.MakeRequest("DELETE", url, nil)
		if err != nil {
			failedRequests = append(failedRequests, FailedDeleteProductRequest{ProductID: id, Reason: err.Error()})
			continue
		}
		count += 1
	}
	p.Wait()

	if len(failedRequests) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_delete_product_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedRequests)
		if err != nil {
			messenger.ExitWithError(err)
			return
		}
	}

	fmt.Printf(color.GreenString("\n\nFinished! 🎉\nDeleted %d out of %d consignments\n"), count, len(ids))
}
