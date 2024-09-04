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
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luca-arch/instaman/database/models"
)

const (
	MaxCopyResults = 100 // The maximum number of users per page to retrieve with copy-followers and copy-following jobs.
	MaxJobsResult  = 20  // The maximum number of jobs per page that are retrieved by FindJobs().
)

var (
	ErrDriverFailure     = errors.New("db error")                // Something went wrong when querying the database.
	ErrFindJobParams     = errors.New("requires id or checksum") // Missing required parameters in FindJob().
	ErrFindCopyJobParams = errors.New("invalid direction")       // Invalid direction passed to FindCopyJob().
	ErrInvalidChecksum   = errors.New("invalid checksum")        // Invalid checksum.
	ErrInvalidID         = errors.New("invalid ID")              // Invalid identifier.
	ErrInvalidType       = errors.New("invalid job tyoe")        // Invalid job type.
)

// FindCopyJobParams defines the search parameters for FindCopyJob().
type FindCopyJobParams struct {
	Direction string `in:"direction,required"`
	UserID    int64  `in:"userID,required"`
	WithPage  *int   `in:"page,omitempty"`
}

// FindJobParams defines the search parameters for FindJob().
type FindJobParams struct {
	Checksum string `in:"checksum"`
	ID       int64  `in:"id"`
	State    string `in:"state"`
	Type     string `in:"type"`
}

// FindJobsParams defines the search parameters for FindJobs().
type FindJobsParams struct {
	// LastRun *time.Time `in:"nextRun"`
	// NextRun *time.Time `in:"nextRun"`
	Order string `in:"order"`
	Page  int32  `in:"page"`
	State string `in:"state"`
	Type  string `in:"type"`
}

// NewCopyJobParams defines the input data for NewCopyJob().
type NewCopyJobParams struct {
	Label    string     `json:"label"`
	NextRun  *time.Time `json:"nextRun"`
	Type     string     `json:"type"`
	Metadata struct {
		Frequency string `json:"frequency"`
		UserID    int64  `json:"userID"` //nolint:tagliatelle // Always capitalise ID suffix.
	} `json:"metadata"`
}

// NewJobParams defines the input data for NewJob().
type NewJobParams struct {
	Checksum string
	Label    string
	Metadata any
	NextRun  *time.Time
	State    string
	Type     string
}

// FindCopyJob finds a job of type `copy-followers` or `copy-following`.
// It calls FindJob and augments the result with the total number of connections already retrieved.
// If WithPage is set, that slice of results is also included in the returned value.
func (d *Database) FindCopyJob(ctx context.Context, params FindCopyJobParams) (*models.CopyJob, error) {
	var table string

	p := FindJobParams{} //nolint:exhaustruct // OK

	switch params.Direction {
	case "followers":
		p.Checksum = models.JobTypeCopyFollowers + ":" + strconv.FormatInt(params.UserID, 10)
		p.Type = models.JobTypeCopyFollowers
		table = "user_followers"
	case "following":
		p.Checksum = models.JobTypeCopyFollowing + ":" + strconv.FormatInt(params.UserID, 10)
		p.Type = models.JobTypeCopyFollowing
		table = "user_following"
	default:
		return nil, ErrFindCopyJobParams
	}

	job, err := d.FindJob(ctx, p)

	switch {
	case err != nil:
		return nil, err
	case job == nil:
		return nil, nil //nolint:nilnil // It means not found
	}

	sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE account_id = $1`, table)
	total, err := Count(ctx, d, sql, params.UserID)

	switch {
	case err != nil:
		return nil, errors.Join(err, ErrDriverFailure)
	case params.WithPage == nil || *params.WithPage < 0:
		return &models.CopyJob{
			Job:     job,
			Results: nil,
			Total:   total,
		}, nil
	}

	limit, offset := *params.WithPage, MaxCopyResults

	sql = `
	SELECT
		user_id,
		first_seen,
		handler,
		last_seen,
		pic_url
	FROM
		` + table + `
	WHERE
		account_id = $1
	ORDER BY
		first_seen DESC
	LIMIT $2 OFFSET $3
	`

	results, err := Select[models.User](ctx, d, sql, params.UserID, limit, offset)
	if err != nil {
		return nil, errors.Join(err, ErrDriverFailure)
	}

	return &models.CopyJob{
		Job:     job,
		Results: results,
		Total:   total,
	}, nil
}

// FindJob finds a job by its ID or checksum.
func (d *Database) FindJob(ctx context.Context, params FindJobParams) (*models.Job, error) {
	if params.ID <= 0 && params.Checksum == "" {
		return nil, ErrFindJobParams
	}

	whereP := make([]string, 0)
	whereV := make([]any, 0)

	if params.ID > 0 {
		whereP = append(whereP, nextPlaceholder("id", whereP))
		whereV = append(whereV, params.ID)
	}

	if params.Checksum != "" {
		whereP = append(whereP, nextPlaceholder("checksum", whereP))
		whereV = append(whereV, params.Checksum)
	}

	if params.State != "" {
		whereP = append(whereP, nextPlaceholder("state", whereP))
		whereV = append(whereV, params.State)
	}

	if params.Type != "" {
		whereP = append(whereP, nextPlaceholder("job_type", whereP))
		whereV = append(whereV, params.Type)
	}

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
	WHERE ` + strings.Join(whereP, " AND ")

	job, err := SelectOne[models.Job](ctx, d, sql, whereV...)

	switch {
	case err == nil:
		return job, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil //nolint:nilnil // It means not found
	default:
		return nil, err
	}
}

// FindJobs returns a list of jobs.
func (d *Database) FindJobs(ctx context.Context, params FindJobsParams) ([]models.Job, error) {
	whereP := make([]string, 0)
	args := make([]any, 0)
	where := ""
	order, dir := "last_run", OrderDesc

	if params.State != "" {
		whereP = append(whereP, nextPlaceholder("state", whereP))
		args = append(args, params.State)
	}

	if params.Type != "" {
		whereP = append(whereP, nextPlaceholder("job_type", whereP))
		args = append(args, params.Type)
	}

	if len(whereP) > 0 {
		where = "WHERE " + strings.Join(whereP, " AND ")
	}

	switch params.Order {
	case "-last_run":
		order, dir = "last_run", OrderDesc
	case "last_run":
		order, dir = "last_run", OrderAsc
	case "-next_run":
		order, dir = "next_run", OrderDesc
	case "next_run":
		order, dir = "next_run", OrderAsc
	case "-state":
		order, dir = "state", OrderDesc
	case "state":
		order, dir = "state", OrderAsc
	case "-label":
		order, dir = "label", OrderDesc
	case "label":
		order, dir = "label", OrderAsc
	}

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
	`

	sql = fmt.Sprintf("%s %s ORDER BY %s %s LIMIT %d OFFSET %d",
		sql, where, order, dir, MaxJobsResult, params.Page*MaxJobsResult)

	jobs, err := Select[models.Job](ctx, d, sql, args...)

	switch {
	case err == nil:
		return jobs, nil
	default:
		return nil, err
	}
}

// NewCopyJob creates a new Job of either type copy-followers or copy-following.
func (d *Database) NewCopyJob(ctx context.Context, params NewCopyJobParams) (*models.CopyJob, error) {
	switch {
	case params.Type != models.JobTypeCopyFollowers && params.Type != models.JobTypeCopyFollowing:
		return nil, ErrFindCopyJobParams
	case params.Metadata.UserID < 1:
		return nil, ErrInvalidID
	}

	j, err := d.NewJob(ctx, NewJobParams{
		Checksum: fmt.Sprintf("%s:%d", params.Type, params.Metadata.UserID),
		Label:    params.Label,
		Metadata: params.Metadata,
		NextRun:  params.NextRun,
		State:    "new",
		Type:     params.Type,
	})
	if err != nil {
		return nil, err
	}

	return &models.CopyJob{
		Job:     j,
		Results: nil,
		Total:   0,
	}, nil
}

// NewJob creates a new Job in the `jobs` table.
func (d *Database) NewJob(ctx context.Context, params NewJobParams) (*models.Job, error) {
	switch {
	case !models.IsValidJobType(params.Type):
		return nil, ErrInvalidType
	case params.Checksum == "":
		return nil, ErrInvalidChecksum
	}

	sql := `
	INSERT INTO jobs (
		checksum,
		job_type,
		label,
		last_run,
		metadata,
		next_run,
		state
	)
	VALUES ($1, $2, $3, NULL, $4, $5, $6)
	RETURNING *
	`

	j, err := SelectOne[models.Job](ctx, d, sql, params.Checksum, params.Type, params.Label, params.Metadata, params.NextRun, params.State)
	if err != nil {
		return nil, errors.Join(err, ErrDriverFailure)
	}

	return j, nil
}

// nextPlaceholder builds prepared statements' placeholders.
func nextPlaceholder(col string, where []string) string {
	return col + " = $" + strconv.Itoa(len(where)+1)
}
