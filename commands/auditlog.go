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

// auditlogCmd represents the auditlog command
var auditlogCmd = &cobra.Command{
	Use:   "export-auditlog",
	Short: "Export Audit Log",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli export-auditlog -d DOMAINPREFIX -t TOKEN -F 2018-03-15T16:30:30 -T 2018-04-01T18:30:00")),

	Run: func(cmd *cobra.Command, args []string) {
		getAuditLog()
	},
}

func init() {
	// Flags
	auditlogCmd.Flags().StringVarP(&dateFrom, "DateFrom", "F", "", "Date from (YYYY-MM-DDT00:00:00)")
	auditlogCmd.Flags().StringVarP(&dateTo, "DateTo", "T", "", "Date to (YYYY-MM-DDT00:00:00)")
	auditlogCmd.MarkFlagRequired("DateFrom")
	auditlogCmd.MarkFlagRequired("DateTo")

	rootCmd.AddCommand(auditlogCmd)
}

func getAuditLog() {

	// Create a new Vend Client
	vc := vend.NewClient(Token, DomainPrefix, "")

	// Parse date input for errors. Sample: 2017-11-20T15:04:05
	layout := "2006-01-02T15:04:05"
	_, err := time.Parse(layout, dateFrom)
	if err != nil {
		log.Printf("incorrect date from: %v, %v", dateFrom, err)
		panic(vend.Exit{1})
	}

	_, err = time.Parse(layout, dateTo)
	if err != nil {
		log.Printf("incorrect date to: %v, %v", dateTo, err)
		panic(vend.Exit{1})
	}

	// Get log
	fmt.Println("\nRetrieving Audit Log from Vend...")
	audit, err := vc.AuditLog(dateFrom, dateTo)
	if err != nil {
		log.Printf("failed retrieving audit log from Vend %v", err)
		panic(vend.Exit{1})
	}

	// Write log to CSV
	fmt.Println("Writing log to CSV file...")
	err = aWriteFile(audit)
	if err != nil {
		log.Printf("failed writing audit log to CSV %v", err)
		panic(vend.Exit{1})
	}

	fmt.Println(color.GreenString("\nFinished!\n"))
}

func aWriteFile(auditEvents []vend.AuditLog) error {

	// Create a blank CSV file
	filename := fmt.Sprintf("%s_audit_log_%v.csv", DomainPrefix, time.Now().Unix())
	file, err := os.Create(fmt.Sprintf("./%s", filename))
	if err != nil {
		return err
	}

	defer file.Close()
	writer := csv.NewWriter(file)

	var header []string
	header = append(header, "id")
	header = append(header, "user_id")
	header = append(header, "kind")
	header = append(header, "action")
	header = append(header, "entity_id")
	header = append(header, "ip_address")
	header = append(header, "user_agent")
	header = append(header, "occurred_at")
	header = append(header, "created_at")

	writer.Write(header)

	for _, auditEvent := range auditEvents {

		var id, userID, kind, action, entityID, IPAddress, userAgent, occurredAt, createdAt string

		if auditEvent.ID != nil {
			id = *auditEvent.ID
		}
		if auditEvent.UserID != nil {
			userID = *auditEvent.UserID
		}
		if auditEvent.Kind != nil {
			kind = *auditEvent.Kind
		}
		if auditEvent.Action != nil {
			action = *auditEvent.Action
		}
		if auditEvent.EntityID != nil {
			entityID = *auditEvent.EntityID
		}
		if auditEvent.IPAddress != nil {
			IPAddress = *auditEvent.IPAddress
		}
		if auditEvent.UserAgent != nil {
			userAgent = *auditEvent.UserAgent
		}
		if auditEvent.OccurredAt != nil {
			occurredAt = *auditEvent.OccurredAt
		}
		if auditEvent.CreatedAt != nil {
			createdAt = *auditEvent.CreatedAt
		}

		var record []string
		record = append(record, id)
		record = append(record, userID)
		record = append(record, kind)
		record = append(record, action)
		record = append(record, entityID)
		record = append(record, IPAddress)
		record = append(record, userAgent)
		record = append(record, occurredAt)
		record = append(record, createdAt)

		writer.Write(record)
	}

	writer.Flush()
	return err
}
