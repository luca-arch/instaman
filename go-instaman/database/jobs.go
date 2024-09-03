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

	"github.com/jackc/pgx/v5"
	"github.com/luca-arch/instaman/database/models"
)

const MaxCopyResults = 100 // The maximum number of users per page to retrieve with copy-followers and copy-following jobs.

var (
	ErrDriverFailure     = errors.New("db error")                // Something went wrong when querying the database.
	ErrFindJobParams     = errors.New("requires id or checksum") // Missing required parameters in FindJob().
	ErrFindCopyJobParams = errors.New("invalid direction")       // Invalid direction passed to FindCopyJob().
)

// FindCopyJobParams defines the search parameters for FindCopyJob().
type FindCopyJobParams struct {
	Direction string `in:"direction,required"`
	UserID    int64  `in:"userID,required"`
	WithPage  *int64 `in:"page,omitempty"`
}

// FindJobParams defines the search parameters for FindJob().
type FindJobParams struct {
	Checksum string `in:"checksum"`
	ID       int64  `in:"id"`
	State    string `in:"state"`
	Type     string `in:"type"`
}

// FindCopyJob finds a job of type `copy-followers` or `copy-following`.
// It calls FindJob and augments the result with the total number of connections already retrieved.
// If WithPage is set, that slice of results is also included in the returned value.
func FindCopyJob(ctx context.Context, db *Database, params FindCopyJobParams) (*models.CopyJob, error) {
	var table string

	p := FindJobParams{} //nolint:exhaustruct // OK

	switch params.Direction {
	case "followers":
		p.Checksum = "copyfollowers:" + strconv.FormatInt(params.UserID, 10)
		p.Type = "copy-followers"
		table = "user_followers"
	case "following":
		p.Checksum = "copyfollowing:" + strconv.FormatInt(params.UserID, 10)
		p.Type = "copy-following"
		table = "user_following"
	default:
		return nil, ErrFindCopyJobParams
	}

	job, err := FindJob(ctx, db, p)

	switch {
	case err != nil:
		return nil, err
	case job == nil:
		return nil, nil //nolint:nilnil // It means not found
	}

	sql := fmt.Sprintf(`SELECT COUNT(*) FROM %s WHERE account_id = $1`, table)
	total, err := Count(ctx, db, sql, params.UserID)

	switch {
	case err != nil:
		return nil, errors.Join(err, ErrDriverFailure)
	case params.WithPage == nil:
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

	results, err := Select[models.User](ctx, db, sql, params.UserID, limit, offset)
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
func FindJob(ctx context.Context, db *Database, params FindJobParams) (*models.Job, error) {
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

	job, err := SelectOne[models.Job](ctx, db, sql, whereV...)

	switch {
	case err == nil:
		return job, nil
	case errors.Is(err, pgx.ErrNoRows):
		return nil, nil //nolint:nilnil // It means not found
	default:
		return nil, err
	}
}

// nextPlaceholder builds prepared statements' placeholders.
func nextPlaceholder(col string, where []string) string {
	return col + " = $" + strconv.Itoa(len(where)+1)
}
