package cmd

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

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
	ids, err := readCSV(FilePath)
	if err != nil {
		log.Printf(color.RedString("Failed to get IDs from the file: %s", FilePath))
		panic(vend.Exit{1})
	}

	// Make the requests
	count := 0
	for _, id := range ids {
		fmt.Printf("\nDeleting %v", id)
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/consignments/%s", DomainPrefix, id)
		_, err = vendClient.MakeRequest("DELETE", url, nil)
		if err != nil {
			fmt.Printf(color.RedString("Failed to delete consignment: %v", err))
			continue
		}
		count += 1

	}
	fmt.Printf(color.GreenString("\n\nFinished! 🎉\nDeleted %d out of %d consignments"), count, len(ids))
}
