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

package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/luca-arch/instaman/database"
	"github.com/luca-arch/instaman/database/models"
	"github.com/luca-arch/instaman/instaproxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInsertJobEvent(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	mockErr := errors.New("mock error")

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
	}

	tests := map[string]struct {
		fields
		wants
	}{
		"insert - ok": {
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL1 := oneLineSQL(`INSERT INTO jobs_events (event_msg, job_id, ts) VALUES ($1, $2, NOW())`)

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL1, "something happened", int64(1)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
		"insert - error": {
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL1 := oneLineSQL(`INSERT INTO jobs_events (event_msg, job_id, ts) VALUES ($1, $2, NOW())`)

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL1, "something happened", int64(1)).
						Return(mockErr)

					return q
				},
			},
			wants{
				err: mockErr,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			err := db.InsertJobEvent(ctx, int64(1), "something happened")

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestNextJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	expectedSQL := oneLineSQL(`
	SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
	FROM jobs
	WHERE
		job_type = $1
		AND next_run IS NOT NULL
		AND next_run < NOW()
		AND state IN ($2, $3)
	ORDER BY next_run ASC LIMIT 1
	`)

	mockErr := errors.New("mock error")
	mockJob := &models.Job{
		BinData: []byte(`{"dummy":true, "data":[]}`),
		ID:      123,
		Type:    "mock-job-type",
	}

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
		job *models.Job
	}

	tests := map[string]struct {
		fields
		wants
	}{
		"select - ok": {
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "mock-job-type", "active", "new").
						Return(mockJob, nil)

					return q
				},
			},
			wants{
				err: nil,
				job: mockJob,
			},
		},
		"none found - ok": {
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					var j *models.Job

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "mock-job-type", "active", "new").
						Return(j, pgx.ErrNoRows)

					return q
				},
			},
			wants{
				err: nil,
				job: nil,
			},
		},
		"error": {
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					var j *models.Job

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "mock-job-type", "active", "new").
						Return(j, mockErr)

					return q
				},
			},
			wants{
				err: mockErr,
				job: nil,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			job, err := db.NextJob(ctx, "mock-job-type")

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.job, job)
		})
	}
}

func TestScheduleJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	mockErr := errors.New("mock error")

	type args struct {
		jobID   int64
		nextRun time.Duration
	}

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
	}

	tests := map[string]struct {
		args
		fields
		wants
	}{
		"next minute - ok": {
			args{
				jobID:   123,
				nextRun: time.Minute,
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`UPDATE jobs SET next_run = NOW() + INTERVAL '60 SECOND', state = $1 WHERE id = $2`)

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "active", int64(123)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
		"next hour - ok": {
			args{
				jobID:   456,
				nextRun: time.Hour,
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`UPDATE jobs SET next_run = NOW() + INTERVAL '3600 SECOND', state = $1 WHERE id = $2`)

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "active", int64(456)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
		"error": {
			args{
				jobID:   456,
				nextRun: time.Minute * 4,
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`UPDATE jobs SET next_run = NOW() + INTERVAL '240 SECOND', state = $1 WHERE id = $2`)

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "active", int64(456)).
						Return(mockErr)

					return q
				},
			},
			wants{
				err: mockErr,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			err := db.ScheduleJob(ctx, test.args.jobID, test.args.nextRun)

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestStoreCopyJobResults(t *testing.T) { //nolint:maintidx // this is maintainable at the minute
	t.Parallel()

	ctx := context.TODO()

	var nilString *string

	mockErr := errors.New("mock error")
	mockUsers := []instaproxy.User{
		{
			FullName:   "john doe",
			Handler:    "johndoe",
			ID:         100,
			PictureURL: nil,
		},
		{
			FullName:   "jane doe",
			Handler:    "janedoe",
			ID:         200,
			PictureURL: urlField(t, "https://example.com/pic.jpeg"),
		},
	}

	expectedSQLWithCursor := oneLineSQL(`
		UPDATE jobs SET
			metadata = jsonb_set(metadata, '{cursor}', to_jsonb($1::text)),
			state = $2
		WHERE id = $3`)

	expectedSQLWithoutCursor := oneLineSQL(`
		UPDATE jobs SET
			metadata = jsonb_set(metadata, '{cursor}', 'null'::jsonb),
			state = $1
		WHERE id = $2`)

	expectedSQLForFollowers := oneLineSQL(`
		INSERT INTO user_followers (account_id, first_seen, handler, last_seen, pic_url, user_id)
			VALUES ($1, NOW(), $2, NOW(), $3, $4)
		ON CONFLICT (account_id, user_id) DO UPDATE
			SET last_seen = NOW(), handler = $2, pic_url = $3`)

	expectedSQLForFollowing := oneLineSQL(`
		INSERT INTO user_following (account_id, first_seen, handler, last_seen, pic_url, user_id)
			VALUES ($1, NOW(), $2, NOW(), $3, $4)
		ON CONFLICT (account_id, user_id) DO UPDATE
			SET last_seen = NOW(), handler = $2, pic_url = $3`)

	type args struct {
		job     *models.CopyJob
		results *instaproxy.Connections
	}

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
	}

	tests := map[string]struct {
		args
		fields
		wants
	}{
		"insert two followers, copy done - ok": {
			args{
				job: &models.CopyJob{
					Job: &models.Job{
						ID:   123,
						Type: "copy-followers",
					},
					Metadata: models.CopyJobMetadata{
						Cursor: nil,
						UserID: 1,
					},
				},
				results: &instaproxy.Connections{
					Next:  nil,
					Users: mockUsers,
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowers, int64(1), "johndoe", nilString, int64(100)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowers, int64(1), "janedoe", strPtr("https://example.com/pic.jpeg"), int64(200)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLWithoutCursor, "active", int64(123)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
		"insert two followers - ok": {
			args{
				job: &models.CopyJob{
					Job: &models.Job{
						ID:   123,
						Type: "copy-followers",
					},
					Metadata: models.CopyJobMetadata{
						Cursor: nil,
						UserID: 1,
					},
				},
				results: &instaproxy.Connections{
					Next:  strPtr("next-cursor-123"),
					Users: mockUsers,
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowers, int64(1), "johndoe", nilString, int64(100)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowers, int64(1), "janedoe", strPtr("https://example.com/pic.jpeg"), int64(200)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLWithCursor, strPtr("next-cursor-123"), "active", int64(123)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
		"insert two following, copy done - ok": {
			args{
				job: &models.CopyJob{
					Job: &models.Job{
						ID:   456,
						Type: "copy-following",
					},
					Metadata: models.CopyJobMetadata{
						Cursor: nil,
						UserID: 2,
					},
				},
				results: &instaproxy.Connections{
					Next:  nil,
					Users: mockUsers,
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowing, int64(2), "johndoe", nilString, int64(100)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowing, int64(2), "janedoe", strPtr("https://example.com/pic.jpeg"), int64(200)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLWithoutCursor, "active", int64(456)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
		"error inserting users": {
			args{
				job: &models.CopyJob{
					Job: &models.Job{
						ID:   456,
						Type: "copy-following",
					},
					Metadata: models.CopyJobMetadata{
						Cursor: nil,
						UserID: 2,
					},
				},
				results: &instaproxy.Connections{
					Next:  nil,
					Users: mockUsers,
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowing, int64(2), "johndoe", nilString, int64(100)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowing, int64(2), "janedoe", strPtr("https://example.com/pic.jpeg"), int64(200)).
						Return(mockErr)

					return q
				},
			},
			wants{
				err: mockErr,
			},
		},
		"error updating cursor": {
			args{
				job: &models.CopyJob{
					Job: &models.Job{
						ID:   456,
						Type: "copy-following",
					},
					Metadata: models.CopyJobMetadata{
						Cursor: nil,
						UserID: 2,
					},
				},
				results: &instaproxy.Connections{
					Next:  nil,
					Users: mockUsers,
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowing, int64(2), "johndoe", nilString, int64(100)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLForFollowing, int64(2), "janedoe", strPtr("https://example.com/pic.jpeg"), int64(200)).
						Return(nil)

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQLWithoutCursor, "active", int64(456)).
						Return(mockErr)

					return q
				},
			},
			wants{
				err: mockErr,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			err := db.StoreCopyJobResults(ctx, test.args.job, test.args.results)

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func TestTouchJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()
	mockErr := errors.New("mock error")

	expectedSQL := "UPDATE jobs SET last_run = NOW() WHERE id = $1"

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
	}

	tests := map[string]struct {
		fields
		wants
	}{
		"insert - ok": {
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, int64(1234)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
		"insert - error": {
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, int64(1234)).
						Return(mockErr)

					return q
				},
			},
			wants{
				err: mockErr,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			err := db.TouchJob(ctx, int64(1234))

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
		})
	}
}
