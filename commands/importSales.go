package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
)

// import sales command
var (
	filePath            string
	overwrite           string
	mode                string
	timeZoneImportSales string

	importSalesCmd = &cobra.Command{
		Use:   "import-sales",
		Short: "Import Sales",
		Long: fmt.Sprintf(`
	Example:
	%s`, color.GreenString("vendcli import-sales -d DOMAINPREFIX -t TOKEN -f FILENAME.csv -m MODE -o OVERWRITE")),
		Run: func(cmd *cobra.Command, args []string) {
			importSales()
		},
	}
)

func init() {
	// Flag
	importSalesCmd.Flags().StringVarP(&filePath, "filename", "f", "", "The name of your file: filename.csv")
	importSalesCmd.MarkFlagRequired("filename")

	importSalesCmd.Flags().StringVarP(&overwrite, "overwrite", "o", "", "overwrite sales: true or false")

	importSalesCmd.Flags().StringVarP(&mode, "mode", "m", "parse", "modes: parse, post")
	importSalesCmd.Flags().StringVarP(&timeZoneImportSales, "Timezone", "z", "", "Timezone of the store in zoneinfo format.")

	rootCmd.AddCommand(importSalesCmd)

}

func importSales() {

	overwriteBool := parseOverwriteFlag(overwrite)
	isPostMode := validateModeFlag(mode)

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// fetch the jsons from file and store them into interfaces
	erroredSales, err := readJSONFile(filePath)
	if err != nil {
		fmt.Printf("Error reading json file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	if isPostMode {
		fmt.Printf("\nRunning command in post mode with overwrite set to %t\n", overwriteBool)
		postSales(erroredSales, overwriteBool)
	} else {
		fmt.Printf("\nRunning command in parse mode\n")
		// 1970-01-01T00:00:00Z is just a dummy date to validate the timezone
		if len(timeZoneImportSales) == 0 {
			fmt.Println("Timezone is required for parse mode")
			panic(vend.Exit{1})
		} else {
			validateTimeZone("1970-01-01T00:00:00Z", timeZone)
			parseSales(erroredSales)
		}
	}
}

// read JSONFile fetches the errored sales from file and stores them into an array of vend.Sale9 structs
func readJSONFile(jsonFilePath string) ([]vend.Sale9, error) {
	// read file
	// Open our jsonFile
	jsonFile, err := os.Open(jsonFilePath)
	if err != nil {
		fmt.Printf("Error opening json file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	// defer closing jsonFile so we can parse it
	defer jsonFile.Close()

	// read our opened jsonFile as a byte array
	fileContent, err := io.ReadAll(jsonFile)
	if err != nil {
		fmt.Printf("Error reading json file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	// unmarshal json
	var sales []vend.Sale9
	err = json.Unmarshal(fileContent, &sales)
	if err != nil {
		return nil, err
	}

	return sales, nil
}

func parseSales(sales9 []vend.Sale9) {
	sales, err := vend.ConvertSale9ToSale(sales9)
	if err != nil {
		fmt.Printf("Error converting sales:\n%s\n", err)
		panic(vend.Exit{1})
	}
	sortBySaleDate(sales)

	// Get other data
	registers, users, customers, customerGroupMap, products := GetVendDataForSalesReport(*vendClient)

	// Create report
	file, err := createErredSalesReport()
	if err != nil {
		fmt.Printf("Error creating CSV file:\n%s\n", err)
		panic(vend.Exit{1})
	}

	defer file.Close()

	file = addSalesReportHeader(file)

	fmt.Println("Writing Sales to CSV file")
	file = writeSalesReport(file, registers, users, customers, customerGroupMap, products, sales, DomainPrefix, timeZoneImportSales)

	fmt.Printf("Sales report created: %s\n", file.Name())
}

func postSales(sales []vend.Sale9, overwrite bool) {
	fmt.Println("Posting Sales!")
}

func parseOverwriteFlag(o string) bool {
	overwriteBool, err := strconv.ParseBool(o)
	if err != nil {
		// default to false
		overwriteBool = false
	}

	return overwriteBool
}

func createErredSalesReport() (*os.File, error) {

	fileName := fmt.Sprintf("erred_sales%s_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		fmt.Printf("Error creating CSV file: %s", err)
		panic(vend.Exit{1})
	}

	writer := csv.NewWriter(file)

	var warningMessage []string
	warningMessage = append(warningMessage, "NOTE:")
	warningMessage = append(warningMessage, "This file contains sales that received an error when posting.")
	warningMessage = append(warningMessage, "These sales have not been posted")

	writer.Write(warningMessage)
	writer.Flush()

	return file, err
}

func validateModeFlag(m string) bool {
	m = strings.ToLower(m)
	if m == "parse" {
		return false
	} else if m == "post" {
		return true
	} else {
		fmt.Printf("'%s' is not a valid option for -m mode. Mode should be 'parse' or 'post'\n", m)
		panic(vend.Exit{1})
	}
}
