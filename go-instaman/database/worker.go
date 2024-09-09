/*
 * Instaman - Simple Instagram account manager.
 *
 * Copyright (C) 2024 Luca Contini
 *
 * This program is free software: you can redistribute it and/or modify it
 * under the terms of the GNU General Public License as published by the Free
 * Software Foundation, either version 3 of the License, or (at your option)
 * any later version.
 *
 * This program is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for
 * more details.
 *
 * You should have received a copy of the GNU General Public License along with
 * this program. If not, see <http://www.gnu.org/licenses/>.
 */

package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luca-arch/instaman/database/models"
	"github.com/luca-arch/instaman/instaproxy"
)

// InsertJobEvent registers a new event in the jobs' audit logs table.
func (d *Database) InsertJobEvent(ctx context.Context, jobID int64, event string) error {
	sqlEvent := `INSERT INTO jobs_events (event_msg, job_id, ts) VALUES ($1, $2, NOW())`

	if err := d.querier.Execute(ctx, d, sqlEvent, event, jobID); err != nil {
		return err //nolint:wrapcheck // Error from the same package
	}

	return nil
}

// NextJob returns the first job that is ready for execution.
func (d *Database) NextJob(ctx context.Context, jobType string) (*models.Job, error) {
	sql := `
	SELECT
		id,
		checksum,
		job_type,
		label,
		last_run,
		metadata,
		next_run,
		state
	FROM
		jobs
	WHERE
		job_type = $1
		AND next_run IS NOT NULL
		AND next_run < NOW()
		AND state IN ($2, $3)
	ORDER BY
		next_run ASC
	LIMIT 1
	`

	job, err := d.querier.SelectJob(ctx, d, sql, jobType, models.JobStateActive, models.JobStateNew)

	switch {
	case err == nil:
		return job, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil //nolint:nilnil // It means not found.
	default:
		return nil, err //nolint:wrapcheck // Error from the same package
	}
}

// ScheduleJob updates a job's `next_run` column.
func (d *Database) ScheduleJob(ctx context.Context, jobID int64, nextRun time.Duration) error {
	interval := fmt.Sprintf("%d SECOND", int(nextRun.Seconds()))
	sqlUpdate := `
		UPDATE jobs
			SET next_run = NOW() + INTERVAL '` + interval + `',
			state = $1
		WHERE id = $2
	`

	if err := d.querier.Execute(ctx, d, sqlUpdate, models.JobStateActive, jobID); err != nil {
		return err //nolint:wrapcheck // Error from the same package
	}

	return nil
}

// StoreCopyJobResults updates the `user_followers` or `user_following` tables and the `jobs.metadata.cursor` value.
func (d *Database) StoreCopyJobResults(ctx context.Context, job *models.CopyJob, results *instaproxy.Connections) error {
	table := "user_followers"
	if job.Type == models.JobTypeCopyFollowing {
		table = "user_following"
	}

	sql := fmt.Sprintf(`
		INSERT INTO %s (account_id, first_seen, handler, last_seen, pic_url, user_id)
			VALUES ($1, NOW(), $2, NOW(), $3, $4)
		ON CONFLICT (account_id, user_id) DO UPDATE
			SET last_seen = NOW(), handler = $2, pic_url = $3
	`, table)

	for _, u := range results.Users {
		d.logger.Debug("upsert "+table, "job.id", job.ID, "user", u)

		if err := d.querier.Execute(ctx, d, sql, job.Metadata.UserID, u.Handler, urlStringPtr(u.PictureURL), u.ID); err != nil {
			return err //nolint:wrapcheck // Error from the same package
		}
	}

	if results.Next == nil {
		sql = `
			UPDATE jobs SET
				metadata = jsonb_set(metadata, '{cursor}', 'null'::jsonb),
				state = $1
			WHERE id = $2
		`

		return d.querier.Execute(ctx, d, sql, models.JobStateActive, job.ID) //nolint:wrapcheck // Error from the same package
	}

	sql = `
		UPDATE jobs SET
			metadata = jsonb_set(metadata, '{cursor}', to_jsonb($1::text)),
			state = $2
		WHERE id = $3
	`

	return d.querier.Execute(ctx, d, sql, results.Next, models.JobStateActive, job.ID) //nolint:wrapcheck // Error from the same package
}

// TouchJob updates the job's last_run value.
func (d *Database) TouchJob(ctx context.Context, jobID int64) error {
	if err := d.querier.Execute(ctx, d, "UPDATE jobs SET last_run = NOW() WHERE id = $1", jobID); err != nil {
		return err //nolint:wrapcheck // Error from the same package
	}

	return nil
}

// urlStringPtr returns a pointer to a string represented by a non-empty URLField.
func urlStringPtr(u *instaproxy.URLField) *string {
	if u == nil {
		return nil
	}

	s := u.String()

	if s == "" {
		return nil
	}

	return &s
}
