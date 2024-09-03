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

// Package service provides several services for communicating between different layers of the application.
package service

import (
	"context"
	"errors"

	"github.com/luca-arch/instaman/database"
	"github.com/luca-arch/instaman/database/models"
)

const MaxCopyResults = 500 // The maximum number of users per page to retrieve with copy-followers and copy-following jobs.

var ErrDBFailure = errors.New("db error") // Generic error wrapper for db failures.

// Jobs is the service that abstracts jobs operations from the database layer.
type Jobs struct {
	db *database.Database
}

// NewJobsService sets up and returns a new Job Service.
func NewJobsService(db *database.Database) *Jobs {
	return &Jobs{
		db: db,
	}
}

// FindCopyJob finds a job of type `copy-followers` or `copy-following`.
// This method does not error if the job isn't found, it returns a nil pointer.
func (j *Jobs) FindCopyJob(ctx context.Context, params database.FindCopyJobParams) (*models.CopyJob, error) {
	cj, err := database.FindCopyJob(ctx, j.db, params)
	if err != nil {
		return nil, errors.Join(err, ErrDBFailure)
	}

	return cj, nil
}

// FindJob finds a job by its ID or checksum.
// This method does not error if the job isn't found, it returns a nil pointer.
func (j *Jobs) FindJob(ctx context.Context, params database.FindJobParams) (*models.Job, error) {
	jj, err := database.FindJob(ctx, j.db, params)
	if err != nil {
		return nil, errors.Join(err, ErrDBFailure)
	}

	return jj, nil
}
