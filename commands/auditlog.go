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

// auditlogCmd represents the auditlog command
var auditlogCmd = &cobra.Command{
	Use:   "audit-log",
	Short: "Audit Log",
	Long: fmt.Sprintf(`
Example:
%s`, color.GreenString("vendcli audit-log -d DOMAINPREFIX -t TOKEN -F 2018-03-01T16:30:30 -T 2018-04-01T18:30:00")),

	Run: func(cmd *cobra.Command, args []string) {
		getAuditLog()
	},
}

func init() {
	// Flags
	auditlogCmd.Flags().StringVarP(&dateFrom, "DateFrom", "F", "", "Date from (YYYY-MM-DD)")
	auditlogCmd.Flags().StringVarP(&dateTo, "DateTo", "T", "", "Date to (YYYY-MM-DD)")
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
		fmt.Printf("incorrect date from: %v, %v", dateFrom, err)
		os.Exit(1)
	}

	_, err = time.Parse(layout, dateTo)
	if err != nil {
		fmt.Printf("incorrect date to: %v, %v", dateTo, err)
		os.Exit(1)
	}

	// Get log
	fmt.Println("Retrieving Audit Log from Vend...")
	audit, err := vc.AuditLog(dateFrom, dateTo)
	if err != nil {
		log.Fatalf("failed retrieving audit log from Vend %v", err)

	}

	// Write log to CSV
	fmt.Println("Writing log to CSV file...")
	err = aWriteFile(audit)
	if err != nil {
		log.Fatalf("failed writing audit log to CSV %v", err)
	}
}

func aWriteFile(audit []vend.AuditLog) error {

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

	for _, audits := range audit {

		var id, userID, kind, action, entityID, IPAddress, userAgent, occurredAt, createdAt string

		if audits.ID != nil {
			id = *audits.ID
		}
		if audits.UserID != nil {
			userID = *audits.UserID
		}
		if audits.Kind != nil {
			kind = *audits.Kind
		}
		if audits.Action != nil {
			action = *audits.Action
		}
		if audits.EntityID != nil {
			entityID = *audits.EntityID
		}
		if audits.IPAddress != nil {
			IPAddress = *audits.IPAddress
		}
		if audits.UserAgent != nil {
			userAgent = *audits.UserAgent
		}
		if audits.OccurredAt != nil {
			occurredAt = *audits.OccurredAt
		}
		if audits.CreatedAt != nil {
			createdAt = *audits.CreatedAt
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
