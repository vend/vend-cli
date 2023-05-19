package cmd

import (
	"fmt"
	"log"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

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
	ids, err := readCSV(FilePath)
	if err != nil {
		log.Printf(color.RedString("Failed to get IDs from the file: %s", FilePath))
		panic(vend.Exit{1})
	}

	// Make the requests
	for _, id := range ids {
		fmt.Printf("\nDeleting %v", id)
		url := fmt.Sprintf("https://%s.vendhq.com/api/2.0/customers/%s", DomainPrefix, id)
		_, err = vendClient.MakeRequest("DELETE", url, nil)
		if err != nil {
			fmt.Printf(color.RedString("Failed to delete customer: %v", err))
		}
	}
	fmt.Println(color.GreenString("\n\nFinished! 🎉\n"))
}
