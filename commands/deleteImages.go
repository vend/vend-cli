package cmd

import (
	"fmt"
	"log"

	"github.com/vend/vend-cli/pkg/messenger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

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
	ids, err := readCSV(FilePath)
	if err != nil {
		log.Printf(color.RedString("Failed to get IDs from the file: %s", FilePath))
		messenger.ExitWithError(err)
	}

	// Make the requests
	count := 0
	for _, id := range ids {
		fmt.Printf("\nDeleting %v", id)
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/product_images/%s", DomainPrefix, id)
		_, err = vendClient.MakeRequest("DELETE", url, nil)
		if err != nil {
			fmt.Printf(color.RedString("Failed to delete image: %v", err))
			continue
		}
		count += 1
		fmt.Println(count)

	}
	fmt.Printf(color.GreenString("\n\nFinished! 🎉\nDeleted %d out of %d images"), count, len(ids))
}
