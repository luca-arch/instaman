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

package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/luca-arch/instaman/database"
	"github.com/luca-arch/instaman/database/models"
	"github.com/luca-arch/instaman/instaproxy"
)

var (
	ErrInstaproxy      = errors.New("instaproxy failure")
	ErrInvalidMetadata = errors.New("could not parse metadata")
	ErrNoRetry         = errors.New("instaproxy fatal")
)

const (
	attempts             = 4 // How many pages of followers/following to consecutively fetch before pausing the job.
	pauseBetweenAttempts = 5 // How many seconds to sleep between each fetch.
)

type dbworker interface {
	InsertJobEvent(ctx context.Context, jobID int64, event string) error
	NextJob(context.Context, string) (*models.Job, error)
	ScheduleJob(context.Context, int64, time.Duration) error
	StoreCopyJobResults(context.Context, *models.CopyJob, *instaproxy.Connections) error
	TouchJob(context.Context, int64) error
	UpdateJob(context.Context, database.UpdateJobParams) error
}

// Worker is the service that abstracts scheduled jobs operations from the database layer.
type Worker struct {
	db        dbworker
	instagram igclient
	logger    *slog.Logger
}

// NewWorkerService sets up and returns a new Worker Service.
func NewWorkerService(db dbworker, logger *slog.Logger, instagramClient igclient) *Worker {
	return &Worker{
		db:        db,
		instagram: instagramClient,
		logger:    logger,
	}
}

func (w *Worker) StartCopying(ctx context.Context) {
	// Start first loop immediately.
	delay := time.Millisecond

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("shutting down worker...")

			return
		case <-time.After(delay):
			job, err := w.NextCopyJob(ctx)

			// Wait one minute between each iteration.
			delay = time.Minute

			switch {
			case err != nil:
				w.logger.Error("could not fetch job", "error", err)
			case job == nil:
				continue
			case w.db.TouchJob(ctx, job.ID) != nil:
				w.logger.Error("could not update job timestamp", "job.id", job.ID, "job.label", job.Label)
			default:
				w.logger.Info("starting job", "job.id", job.ID, "job.label", job.Label, "job.type", job.Type)

				if err := w.RunCopyJob(ctx, job); err != nil {
					w.logger.Error("could not execute job", "error", err, "job.id", job.ID, "job.label", job.Label)

					if err := w.db.InsertJobEvent(ctx, job.ID, err.Error()); err != nil {
						w.logger.Error("could not log job event", "error", err)
					}
				}

				//nolint:durationcheck // Pause for 10~15 minutes not to flood the api.
				sleep := time.Minute * randDuration(10, 15) //nolint:mnd
				time.Sleep(sleep)
			}
		}
	}
}

// NextCopyJob returns the next scheduled CopyJob that is ready for execution.
func (w *Worker) NextCopyJob(ctx context.Context) (*models.CopyJob, error) {
	j, err := w.db.NextJob(ctx, models.JobTypeCopyFollowers)

	switch {
	case err != nil:
		return nil, errors.Join(ErrDBFailure, err)
	case j == nil:
		j, err = w.db.NextJob(ctx, models.JobTypeCopyFollowing)
	}

	switch {
	case err != nil:
		return nil, errors.Join(ErrDBFailure, err)
	case j == nil:
		return nil, nil //nolint:nilnil // It means not found.
	}

	cj, err := models.NewCopyJob(j)
	if err != nil {
		return nil, errors.Join(ErrDBFailure, err)
	}

	return cj, nil
}

// RunCopyJob executes a CopyJob.
func (w *Worker) RunCopyJob(ctx context.Context, cj *models.CopyJob) error {
	if err := w.db.InsertJobEvent(ctx, cj.ID, "job picked up for execution"); err != nil {
		w.logger.Error("could not log job event", "error", err)
	}

	cursor, done := cj.Metadata.Cursor, false

Loop:
	for a := range attempts {
		res, err := w.instagram.GetFollowers(ctx, cj.Metadata.UserID, cursor)
		if err != nil {
			return errors.Join(
				w.db.UpdateJob(ctx, database.UpdateJobParams{ //nolint:exhaustruct
					ID:    cj.ID,
					State: models.JobStateError,
				}),
				w.db.InsertJobEvent(ctx, cj.ID, err.Error()),
				err,
				ErrNoRetry,
			)
		}

		cursor = res.Next

		if err := w.db.StoreCopyJobResults(ctx, cj, res); err != nil {
			return errors.Join(ErrDBFailure, err)
		}

		if err := w.db.InsertJobEvent(ctx, cj.ID, fmt.Sprintf("Copied %d users. Next cursor: %v", len(res.Users), cursor)); err != nil {
			w.logger.Error("could not log job event", "error", err)
		}

		switch {
		case cursor == nil, *cursor == "":
			done = true

			break Loop
		case a != attempts:
			time.Sleep(time.Duration(pauseBetweenAttempts) * time.Second)
		}
	}

	//nolint:durationcheck // Pause for 20~30 minutes not to flood the api.
	freq := time.Minute * randDuration(20, 30) //nolint:mnd

	if done {
		if err := w.db.InsertJobEvent(ctx, cj.ID, "Sync completed"); err != nil {
			w.logger.Error("could not log job event", "error", err)
		}

		switch cj.Metadata.Frequency {
		case models.JobFrequencyDaily:
			freq = time.Hour * 24 //nolint:mnd
		case models.JobFrequencyWeekly:
			freq = time.Hour * 24 * 7 //nolint:mnd
		}
	}

	if err := w.db.ScheduleJob(ctx, cj.ID, freq); err != nil {
		return errors.Join(ErrDBFailure, err)
	}

	return nil
}

// randDuration returns a random duration in between two values.
func randDuration(from, to int) time.Duration {
	d := from + rand.IntN(to-from) //nolint:gosec

	return time.Duration(d)
}
