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

type dbjobs interface {
	FindCopyJob(context.Context, database.FindCopyJobParams) (*models.CopyJob, error)
	FindJob(context.Context, database.FindJobParams) (*models.Job, error)
	FindJobs(context.Context, database.FindJobsParams) ([]models.Job, error)
	NewCopyJob(context.Context, database.NewCopyJobParams) (*models.CopyJob, error)
}

// Jobs is the service that abstracts jobs operations from the database layer.
type Jobs struct {
	db dbjobs
}

// NewJobsService sets up and returns a new Job Service.
func NewJobsService(db dbjobs) *Jobs {
	return &Jobs{
		db: db,
	}
}

// FindCopyJob finds a job of type `copy-followers` or `copy-following`.
// This method does not error if the job isn't found, it returns a nil pointer.
func (j *Jobs) FindCopyJob(ctx context.Context, params database.FindCopyJobParams) (*models.CopyJob, error) {
	cj, err := j.db.FindCopyJob(ctx, params)
	if err != nil {
		return nil, errors.Join(ErrDBFailure, err)
	}

	return cj, nil
}

// FindJob finds a job by its ID or checksum.
// This method does not error if the job isn't found, it returns a nil pointer.
func (j *Jobs) FindJob(ctx context.Context, params database.FindJobParams) (*models.Job, error) {
	jj, err := j.db.FindJob(ctx, params)
	if err != nil {
		return nil, errors.Join(ErrDBFailure, err)
	}

	return jj, nil
}

// FindJobs retrieves a list of jobs from the database.
func (j *Jobs) FindJobs(ctx context.Context, params database.FindJobsParams) ([]models.Job, error) {
	jobs, err := j.db.FindJobs(ctx, params)
	if err != nil {
		return nil, errors.Join(ErrDBFailure, err)
	}

	return jobs, nil
}

// NewCopyJob creates a new CopyJob in the database and returns it.
func (j *Jobs) NewCopyJob(ctx context.Context, params database.NewCopyJobParams) (*models.CopyJob, error) {
	cj, err := j.db.NewCopyJob(ctx, params)
	if err != nil {
		return nil, errors.Join(ErrDBFailure, err)
	}

	return cj, nil
}
