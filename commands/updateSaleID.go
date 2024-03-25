package cmd

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/vend/vend-cli/pkg/csvparser"
	"github.com/vend/vend-cli/pkg/messenger"
	pbar "github.com/vend/vend-cli/pkg/progressbar"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/vend/govend/vend"
	"github.com/wallclockbuilder/stringutil"
)

type FailedUpdateSaleIDRequests struct {
	SaleID string
	UserID string
	Reason string
}

// Command config
var updateSaleIDcmd = &cobra.Command{
	Use:   "update-sale-user-id",
	Short: "Update the user id for a list of sales",
	Long: fmt.Sprintf(`
Updates the user for a given list of sales.

%s This command is very dangerous and can result in duplicate sales, missing data or other unintended behavior. 
If you are unsure of its use, %s and seek help from a Product Specialist


csv should be in the following format
+-------------+-------------+
|   sale_id   |   user_id   |
+-------------+-------------+
| <sale UUID> | <user UUID> |
+-------------+-------------+

** Also prints a "post in case of emergency" log file 
** filename: DOMAINNAME_update_saleid_original_sales_before_update_TIMESTAMP.txt
** Keep this file, in case an issue occurs we can use this to recover data

Example: %s
`,

		color.RedString("WARNING:"),
		color.RedString("do not use this"),
		color.GreenString("vendcli update-sale-user-id -t TOKEN -d DOMAINPREFIX -f PATH/TO/FILE")),

	Run: func(cmd *cobra.Command, args []string) {
		updateSaleID()
	},
}

var failedUpdateSaleIDRequests []FailedUpdateSaleIDRequests

func init() {
	// Flag
	updateSaleIDcmd.Flags().StringVarP(&FilePath, "Filename", "f", "", "The name of your file: filename.csv")
	updateSaleIDcmd.MarkFlagRequired("Filename")

	rootCmd.AddCommand(updateSaleIDcmd)

}

func updateSaleID() {

	// Create Log File
	logFileName := fmt.Sprintf("%s_update_saleid_original_sales_before_update_%v.txt", DomainPrefix, time.Now().Unix())
	logFile, err := os.OpenFile(fmt.Sprintf("./%s", logFileName), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		err = fmt.Errorf("error:%s", err)
		messenger.ExitWithError(err)
	}
	fmt.Println("\nSaving original sales to: ", color.YellowString(logFileName))
	fmt.Println("-- Keep this file, in case an issue occurs we can use this to recover data --")
	log.SetOutput(logFile)
	defer logFile.Close()

	fmt.Println("\n\nStarting Command Update Sale User ID..")
	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")
	vendClient = &vc

	// Get passed entities from CSV
	fmt.Printf("\nReading CSV...\n")
	saleList, err := ReadSaleUserCSV(FilePath)
	if err != nil {
		err = fmt.Errorf("failed to get ids from the file: %s, error: %w", FilePath, err)
		messenger.ExitWithError(err)
	}

	// loop through entities, fetch the data from vend, swap sale_id, and post to vend
	fmt.Println("\nUpdating Sales...")
	succesfulPosts := PostUpdateSaleID(saleList)

	if len(failedUpdateSaleIDRequests) > 0 {
		fmt.Println(color.RedString("\nThere were some errors. Writing failures to csv.."))
		fileName := fmt.Sprintf("%s_failed_update_saleid_requests_%v.csv", DomainPrefix, time.Now().Unix())
		err := csvparser.WriteErrorCSV(fileName, failedUpdateSaleIDRequests)
		if err != nil {
			messenger.ExitWithError(err)
			return
		}

	}
	fmt.Println(color.GreenString("\n\nFinished! 🎉\nSuccesfully adjusted %d of %d Customer Loyalty Balances", succesfulPosts, len(saleList)))
}

func PostUpdateSaleID(saleList []vend.SaleUserUpload) int {
	p := pbar.CreateSingleBar()
	bar, err := p.AddProgressBar(len(saleList), "Updating")
	if err != nil {
		fmt.Printf("Error creating progress bar:%s\n", err)
	}
	var count int = 0
	for _, saleRequest := range saleList {
		bar.Increment()

		// Get Sale
		sale, err := getSale9(saleRequest.SaleID)
		if err != nil {
			err = fmt.Errorf("error getting sale info: %s", err)
			failedUpdateSaleIDRequests = append(failedUpdateSaleIDRequests,
				FailedUpdateSaleIDRequests{
					SaleID: saleRequest.SaleID,
					UserID: saleRequest.UserID,
					Reason: err.Error(),
				})
			continue
		}

		// Change User on the Sale
		sale = changeUser(sale, saleRequest.UserID)

		// Make the request
		url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales", DomainPrefix)
		resp, err := vendClient.MakeRequest("POST", url, sale)
		if err != nil {
			err = fmt.Errorf("error updating sale info: %s, response: %s", err, string(resp))
			failedUpdateSaleIDRequests = append(failedUpdateSaleIDRequests,
				FailedUpdateSaleIDRequests{
					SaleID: saleRequest.SaleID,
					UserID: saleRequest.UserID,
					Reason: err.Error(),
				})
			continue
		}
		count += 1
	}
	p.Wait()
	return count
}

// getSale9 pulls sales information from the 0.9 Vend API
func getSale9(id string) (vend.Sale9, error) {

	sale := vend.Sale9{}
	saleResponse := vend.RegisterSale9{}

	// Create the Vend URL
	url := fmt.Sprintf("https://%s.vendhq.com/api/register_sales/%s", DomainPrefix, id)
	res, err := vendClient.MakeRequest("GET", url, nil)
	if err != nil {
		err = fmt.Errorf("error getting sale info: %s", err)
		return sale, err
	}
	// log the sale info for later in case we need to recover it
	log.Println(id, string(res))

	// Unmarshal JSON Response
	err = json.Unmarshal(res, &saleResponse)
	if err != nil {
		err = fmt.Errorf("error unmarshalling sale info: %s", err)
		return sale, err
	}

	// return sale info if results, otherwise return err
	if len(saleResponse.RegisterSale9) > 0 {
		return saleResponse.RegisterSale9[0], nil
	} else {
		err = errors.New("failed to GET sale info from vend")
		return sale, err
	}
}

// Swap exisiting user for new desired user
func changeUser(sale vend.Sale9, userID string) vend.Sale9 {

	*sale.UserID = userID
	// also set any salesperson ids to new desired user
	if len(sale.RegisterSaleProducts) > 0 {
		for _, product := range sale.RegisterSaleProducts {
			for _, attribute := range product.Attributes {
				if attribute.Name != nil && *attribute.Name == "salesperson_id" && attribute.Value != nil {
					*attribute.Value = userID
				}
			}
		}
	}
	return sale
}

// ReadSaleUserCSV reads the provided CSV file and stores the input as sale structs.
func ReadSaleUserCSV(FilePath string) ([]vend.SaleUserUpload, error) {

	p := pbar.CreateSingleBar()
	bar, err := p.AddIndeterminateProgressBar("Reading CSV")
	if err != nil {
		err = fmt.Errorf("error creating progress bar:%s", err)
		return nil, err
	}
	done := make(chan struct{})
	go bar.AnimateIndeterminateBar(done)

	file, err := os.Open(FilePath)
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf(`%s - please check you've specified the right file path.%sTip: make sure you're in the same folder as your file. Use "cd ~/Downloads" to navigate to your Downloads folder`, err, "\n")
		return []vend.SaleUserUpload{}, err
	}
	defer file.Close()
	reader := csv.NewReader(file)

	header := []string{"sale_id", "user_id"}
	headerRow, err := reader.Read()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("failed to read headerow. %w", err)
		return []vend.SaleUserUpload{}, err
	}

	if len(headerRow) > 2 {
		bar.AbortBar()
		p.Wait()
		err = errors.New("header row longer than expected")
		return []vend.SaleUserUpload{}, err
	}

	// Check each string in the header row is same as our template.
	for i, row := range headerRow {
		if stringutil.Strip(strings.ToLower(row)) != header[i] {
			bar.AbortBar()
			p.Wait()
			err = fmt.Errorf("mismatched CSV headers, expecting {sale_id, user_id}")
			return []vend.SaleUserUpload{}, fmt.Errorf("mistmatched headers %v", err)
		}
	}

	// Read the rest of the data from the CSV.
	rawData, err := reader.ReadAll()
	if err != nil {
		bar.AbortBar()
		p.Wait()
		return []vend.SaleUserUpload{}, err
	}

	// Loop through rows and assign them to saleUserUpload struct.
	var sale vend.SaleUserUpload
	var saleList []vend.SaleUserUpload
	for idx, row := range rawData {
		sale, err = readSaleUserRow(row)
		if err != nil {
			err = fmt.Errorf("error reading csv row %d: %s", idx+1, err) // +1 to make it 1-indexed
			failedUpdateSaleIDRequests = append(failedUpdateSaleIDRequests,
				FailedUpdateSaleIDRequests{
					SaleID: row[0],
					UserID: row[1],
					Reason: err.Error(),
				})
			continue
		}
		saleList = append(saleList, sale)
	}
	// Check how many rows we successfully read and stored.
	if len(saleList) > 0 {
	} else {
		bar.AbortBar()
		p.Wait()
		err = fmt.Errorf("no valid sales found")
		return saleList, err
	}
	bar.SetIndeterminateBarComplete()
	p.Wait()

	return saleList, err
}

// Read a single row of a CSV file and check for errors.
func readSaleUserRow(row []string) (vend.SaleUserUpload, error) {
	var sale vend.SaleUserUpload

	sale.SaleID = row[0]
	sale.UserID = row[1]

	for i := range row {
		if len(row[i]) < 1 {
			err := errors.New("missing field")
			return sale, err
		}
	}
	return sale, nil
}
