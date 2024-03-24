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

type FailedImageDeleteRequest struct {
	ImageID string
	Reason  string
}

// deleteImagesCmd represents the deleteImages command
var deleteImagesCmd = &cobra.Command{
	Use:   "delete-images",
	Short: "Delete Images",
	Long: fmt.Sprintf(`
This tool requires a CSV of Image IDs, no headers.

Example:
%s`, color.GreenString("vendcli delete-images -d DOMAINPREFIX -t TOKEN -f FILENAME.csv")),
	Run: func(cmd *cobra.Command, args []string) {
		deleteImages()
	},
}

func init() {
	// Flag
	deleteImagesCmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	deleteImagesCmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(deleteImagesCmd)
}

func deleteImages() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Println("\nReading CSV...")
	ids, err := csvparser.ReadIdCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("failed to get IDs from the file: %s Error:%s", FilePath, err)
		messenger.ExitWithError(err)
	}

	failedRequests := []FailedImageDeleteRequest{}

	// Make the requests
	fmt.Println("\nDeleting images...")
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(ids), "Deleting")
	if err != nil {
		fmt.Println("Error creating progress bar:", err)
	}

	count := 0
	for _, id := range ids {
		bar.Increment()
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/product_images/%s", DomainPrefix, id)
		_, err = vendClient.MakeRequest("DELETE", url, nil)
		if err != nil {
			failedRequests = append(failedRequests, FailedImageDeleteRequest{ImageID: id, Reason: err.Error()})
			continue
		}
		count += 1
		fmt.Println(count)

	}
	p.Wait()

	if len(failedRequests) > 0 {
		fmt.Println(color.RedString("\n\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_delete_image_requests__%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedRequests)
		if err != nil {
			messenger.ExitWithError(err)
			return
		}
	}

	fmt.Printf(color.GreenString("\n\nFinished! 🎉\nDeleted %d out of %d images"), count, len(ids))
}
