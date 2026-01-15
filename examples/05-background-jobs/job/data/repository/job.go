// Package JobRepository stores job state in SQLite.
package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ncobase/ncore/examples/05-background-jobs/job/structs"
)

type JobRepository interface {
	Create(ctx context.Context, job *structs.Job) error
	Update(ctx context.Context, job *structs.Job) error
	FindByID(ctx context.Context, id string) (*structs.Job, error)
	List(ctx context.Context) ([]*structs.Job, error)
	Stats(ctx context.Context) (map[structs.JobStatus]int, error)
}

type jobRepository struct {
	db *sql.DB
}

func NewJobRepository(db *sql.DB) (JobRepository, error) {
	repo := &jobRepository{db: db}
	if err := repo.initSchema(context.Background()); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *jobRepository) initSchema(ctx context.Context) error {
	_, err := r.db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS jobs (
			id TEXT PRIMARY KEY,
			type TEXT NOT NULL,
			payload TEXT NOT NULL,
			status TEXT NOT NULL,
			progress INTEGER NOT NULL,
			result TEXT,
			error TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL,
			started_at TEXT,
			ended_at TEXT
		);
	`)
	return err
}

func (r *jobRepository) Create(ctx context.Context, job *structs.Job) error {
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	resultJSON, err := json.Marshal(job.Result)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO jobs (
			id, type, payload, status, progress, result, error, created_at, updated_at, started_at, ended_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		job.ID,
		job.Type,
		string(payloadJSON),
		string(job.Status),
		job.Progress,
		string(resultJSON),
		job.Error,
		job.CreatedAt.UTC().Format(time.RFC3339Nano),
		job.UpdatedAt.UTC().Format(time.RFC3339Nano),
		formatTime(job.StartedAt),
		formatTime(job.EndedAt),
	)
	return err
}

func (r *jobRepository) Update(ctx context.Context, job *structs.Job) error {
	payloadJSON, err := json.Marshal(job.Payload)
	if err != nil {
		return err
	}

	resultJSON, err := json.Marshal(job.Result)
	if err != nil {
		return err
	}

	_, err = r.db.ExecContext(ctx, `
		UPDATE jobs
		SET type = ?, payload = ?, status = ?, progress = ?, result = ?, error = ?, updated_at = ?, started_at = ?, ended_at = ?
		WHERE id = ?
	`,
		job.Type,
		string(payloadJSON),
		string(job.Status),
		job.Progress,
		string(resultJSON),
		job.Error,
		job.UpdatedAt.UTC().Format(time.RFC3339Nano),
		formatTime(job.StartedAt),
		formatTime(job.EndedAt),
		job.ID,
	)
	return err
}

func (r *jobRepository) FindByID(ctx context.Context, id string) (*structs.Job, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, type, payload, status, progress, result, error, created_at, updated_at, started_at, ended_at
		FROM jobs WHERE id = ?
	`, id)

	var payloadJSON string
	var resultJSON string
	var status string
	var createdAt string
	var updatedAt string
	var startedAt sql.NullString
	var endedAt sql.NullString

	j := &structs.Job{}
	if err := row.Scan(
		&j.ID,
		&j.Type,
		&payloadJSON,
		&status,
		&j.Progress,
		&resultJSON,
		&j.Error,
		&createdAt,
		&updatedAt,
		&startedAt,
		&endedAt,
	); err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(payloadJSON), &j.Payload); err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(resultJSON), &j.Result); err != nil {
		return nil, err
	}

	j.Status = structs.JobStatus(status)

	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return nil, err
	}
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return nil, err
	}

	j.CreatedAt = parsedCreatedAt
	j.UpdatedAt = parsedUpdatedAt
	j.StartedAt, err = parseTime(startedAt)
	if err != nil {
		return nil, err
	}
	j.EndedAt, err = parseTime(endedAt)
	if err != nil {
		return nil, err
	}

	return j, nil
}

func (r *jobRepository) List(ctx context.Context) ([]*structs.Job, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, payload, status, progress, result, error, created_at, updated_at, started_at, ended_at
		FROM jobs ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []*structs.Job
	for rows.Next() {
		var payloadJSON string
		var resultJSON string
		var status string
		var createdAt string
		var updatedAt string
		var startedAt sql.NullString
		var endedAt sql.NullString

		j := &structs.Job{}
		if err := rows.Scan(
			&j.ID,
			&j.Type,
			&payloadJSON,
			&status,
			&j.Progress,
			&resultJSON,
			&j.Error,
			&createdAt,
			&updatedAt,
			&startedAt,
			&endedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal([]byte(payloadJSON), &j.Payload); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(resultJSON), &j.Result); err != nil {
			return nil, err
		}

		j.Status = structs.JobStatus(status)

		parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
		if err != nil {
			return nil, err
		}
		parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
		if err != nil {
			return nil, err
		}
		j.CreatedAt = parsedCreatedAt
		j.UpdatedAt = parsedUpdatedAt
		j.StartedAt, err = parseTime(startedAt)
		if err != nil {
			return nil, err
		}
		j.EndedAt, err = parseTime(endedAt)
		if err != nil {
			return nil, err
		}

		jobs = append(jobs, j)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobs, nil
}

func (r *jobRepository) Stats(ctx context.Context) (map[structs.JobStatus]int, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT status, COUNT(*)
		FROM jobs
		GROUP BY status
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := map[structs.JobStatus]int{
		structs.StatusPending:   0,
		structs.StatusRunning:   0,
		structs.StatusCompleted: 0,
		structs.StatusFailed:    0,
	}

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		stats[structs.JobStatus(status)] = count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return stats, nil
}

func formatTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func parseTime(value sql.NullString) (*time.Time, error) {
	if !value.Valid || value.String == "" {
		return nil, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil, fmt.Errorf("invalid time value: %w", err)
	}
	return &parsed, nil
}
