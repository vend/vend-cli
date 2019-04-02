package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/fatih/color"
	"github.com/jackharrisonsherlock/govend/vend"
	"github.com/spf13/cobra"
)

// exportusersCmd represents the exportusers command
var exportusersCmd = &cobra.Command{
	Use:   "export-users",
	Short: "Export Users",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-users -d DOMAINPREFIX -t TOKEN")),

	Run: func(cmd *cobra.Command, args []string) {
		getAllUsers()
	},
}

func init() {
	rootCmd.AddCommand(exportusersCmd)
}

// Run executes the process of grabbing Users then writing them to CSV.
func getAllUsers() {

	// Create new Vend Client.
	vc := vend.NewClient(Token, DomainPrefix, "")

	// Get Users.
	fmt.Println("\nRetrieving Users from Vend...")
	users, err := vc.Users()
	if err != nil {
		log.Fatalf("Failed retrieving Users from Vend %v", err)
	}

	// Write Users to CSV
	fmt.Println("Writing Users to CSV file...")
	err = uWriteFile(users)
	if err != nil {
		log.Fatalf("Failed writing Users to CSV: %v", err)
	}

	fmt.Println(color.GreenString("\nExported %v Users\n", len(users)))
}

// WriteFile writes customer info to file.
func uWriteFile(users []vend.User) error {

	// Create a blank CSV file.
	fileName := fmt.Sprintf("%s_user_export_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", fileName))
	if err != nil {
		return err
	}

	// Ensure the file is closed at the end.
	defer file.Close()

	// Create CSV writer on the file.
	writer := csv.NewWriter(file)

	// Write the header line.
	var header []string
	header = append(header, "id")
	header = append(header, "username")
	header = append(header, "display_name")
	header = append(header, "account_type")
	header = append(header, "email")
	header = append(header, "restricted_outlet")
	header = append(header, "created_at")
	header = append(header, "deleted_at")

	// Commit the header.
	writer.Write(header)

	// Now loop through each Users object and populate the CSV.
	for _, user := range users {

		var id, username, displayName, accountType, email, restrictedOutlet, createdAt, deletedAt string

		if user.ID != nil {
			id = *user.ID
		}
		if user.Username != nil {
			username = *user.Username
		}
		if user.DisplayName != nil {
			displayName = *user.DisplayName
		}
		if user.AccountType != nil {
			accountType = *user.AccountType
		}
		if user.Email != nil {
			email = *user.Email
		}
		if user.RestrictedOutlet != nil {
			restrictedOutlet = *user.RestrictedOutlet
		}
		if user.CreatedAt != nil {
			createdAt = *user.CreatedAt
		}
		if user.DeletedAt != nil {
			deletedAt = *user.DeletedAt
		}

		var record []string
		record = append(record, id)
		record = append(record, username)
		record = append(record, displayName)
		record = append(record, accountType)
		record = append(record, email)
		record = append(record, restrictedOutlet)
		record = append(record, createdAt)
		record = append(record, deletedAt)

		writer.Write(record)
	}

	writer.Flush()
	return err
}
