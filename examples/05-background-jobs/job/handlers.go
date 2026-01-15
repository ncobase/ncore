package job

import (
	"context"
	"fmt"
	"time"

	"github.com/ncobase/ncore/examples/05-background-jobs/job/structs"
)

// RegisterBuiltInHandlers registers built-in job handlers.
func RegisterBuiltInHandlers(mgr *Manager) {
	// Email job handler
	mgr.RegisterHandler("email", func(ctx context.Context, job *structs.Job, updateProgress func(int)) error {
		to, ok := job.Payload["to"].(string)
		if !ok {
			return fmt.Errorf("invalid 'to' parameter")
		}

		subject, _ := job.Payload["subject"].(string)
		_ = job.Payload["body"] // Not used in simulation, but would be sent in real implementation

		// Simulate email sending
		updateProgress(10)
		time.Sleep(500 * time.Millisecond)

		updateProgress(50)
		time.Sleep(500 * time.Millisecond)

		updateProgress(90)
		time.Sleep(500 * time.Millisecond)

		job.Result = map[string]any{
			"sent_to": to,
			"subject": subject,
			"message": "Email sent successfully",
		}

		return nil
	})

	// Data export job handler
	mgr.RegisterHandler("export", func(ctx context.Context, job *structs.Job, updateProgress func(int)) error {
		format, _ := job.Payload["format"].(string)
		if format == "" {
			format = "csv"
		}

		// Simulate data export
		totalRecords := 1000
		for i := range 10 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(200 * time.Millisecond)
				progress := (i + 1) * 10
				updateProgress(progress)
			}
		}

		job.Result = map[string]any{
			"format":        format,
			"total_records": totalRecords,
			"file_url":      fmt.Sprintf("/downloads/export_%s.%s", job.ID, format),
		}

		return nil
	})

	// Data cleanup job handler
	mgr.RegisterHandler("cleanup", func(ctx context.Context, job *structs.Job, updateProgress func(int)) error {
		olderThan, _ := job.Payload["older_than"].(string)
		if olderThan == "" {
			olderThan = "30d"
		}

		// Simulate cleanup
		deletedCount := 0
		for i := range 5 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(300 * time.Millisecond)
				deletedCount += 50
				progress := (i + 1) * 20
				updateProgress(progress)
			}
		}

		job.Result = map[string]any{
			"deleted_count": deletedCount,
			"criteria":      olderThan,
		}

		return nil
	})

	// Report generation job handler
	mgr.RegisterHandler("report", func(ctx context.Context, job *structs.Job, updateProgress func(int)) error {
		reportType, _ := job.Payload["type"].(string)
		dateRange, _ := job.Payload["date_range"].(string)

		// Simulate report generation
		stages := []string{"Fetching data", "Processing", "Generating charts", "Creating PDF", "Uploading"}
		for i := range stages {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				time.Sleep(400 * time.Millisecond)
				progress := (i + 1) * 20
				updateProgress(progress)
			}
		}

		job.Result = map[string]any{
			"type":       reportType,
			"date_range": dateRange,
			"report_url": fmt.Sprintf("/reports/report_%s.pdf", job.ID),
			"pages":      15,
		}

		return nil
	})
}
