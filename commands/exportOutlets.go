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

	// Get Outlets
	fmt.Println("\nRetrieving Outlets from Vend...")
	outlets := fetchDataForOutletExport()

	// Write Outlets to CSV
	fmt.Println("\nWriting Outlets to CSV file...")
	err := writeOutletExport(outlets)
	if err != nil {
		err = fmt.Errorf("failed creating CSV file %v", err)
		messenger.ExitWithError(err)
	}

	fmt.Println(color.GreenString("\nExported %v outlets 🎉\n", len(outlets)))

}

func fetchDataForOutletExport() []vend.Outlet {
	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("outlets")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	outlets, _, err := vc.Outlets()

	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("failed while retrieving outlets: %v", err)
		messenger.ExitWithError(err)
	}

	bar.SetIndeterminateBarComplete()
	p.Wait()

	return outlets
}

func writeOutletExport(outlets []vend.Outlet) error {
	file, err := createOutletReport()
	if err != nil {
		return err
	}

	file = addHeaderOutletReport(file)
	writeOutletReport(file, outlets)

	return nil
}

func createOutletReport() (*os.File, error) {

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

	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(outlets), "Writing CSV")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}

	writer := csv.NewWriter(file)

	for _, outlet := range outlets {
		bar.Increment()

		var line []string
		line = append(line, *outlet.ID)   //0
		line = append(line, *outlet.Name) //1

		// Write line to file.
		writer.Write(line)
	}
	p.Wait()
	writer.Flush()
	return file

}
