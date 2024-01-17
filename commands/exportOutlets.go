package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// voidSalesCmd represents the voidSales command
var exportOutletsCmd = &cobra.Command{
	Use:   "export-outlets",
	Short: "Export Outlets",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-outlets -d DOMAINPREFIX -t TOKEN")),
	Run: func(cmd *cobra.Command, args []string) {
		exportOutlets()
	},
}

func init() {
	// Flag
	rootCmd.AddCommand(exportOutletsCmd)

}

func exportOutlets() {

	// Get Vend Client
	vc := vend.NewClient(Token, DomainPrefix, "")

	// Get Outlets
	outlets, _, err := vc.Outlets()
	if err != nil {
		log.Printf("Failed retrieving outlets from Vend %v", err)
		panic(vend.Exit{1})
	}

	file, err := createOutletReport(DomainPrefix)
	if err != nil {
		log.Printf("Failed creating CSV file %v", err)
		panic(vend.Exit{1})
	}

	file = addHeaderOutletReport(file)
	file = writeOutletReport(file, outlets)
	fmt.Printf("Exported %v outlets to %s\n", len(outlets), file.Name())

}

func createOutletReport(domainPrefix string) (*os.File, error) {

	fileName := fmt.Sprintf("%s_export_outlets_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		return nil, err
	}

	return file, err
}

func addHeaderOutletReport(file *os.File) *os.File {

	writer := csv.NewWriter(file)

	// Set header values.
	var headerLine []string
	headerLine = append(headerLine, "id")   //0
	headerLine = append(headerLine, "name") //1

	// Write headerline to file.
	writer.Write(headerLine)
	writer.Flush()

	return file
}

func writeOutletReport(file *os.File, outlets []vend.Outlet) *os.File {

	writer := csv.NewWriter(file)

	for _, outlet := range outlets {

		var line []string
		line = append(line, *outlet.ID)   //0
		line = append(line, *outlet.Name) //1

		// Write line to file.
		writer.Write(line)
	}
	writer.Flush()
	return file

}
