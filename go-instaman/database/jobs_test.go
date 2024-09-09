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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestFindCopyJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	mockCopyFollowersJob := &models.Job{
		BinData: []byte(`{"userID":123, "frequency":"daily"}`),
		ID:      1,
		Type:    "copy-followers",
	}

	mockCopyFollowingJob := &models.Job{
		BinData: []byte(`{"userID":456, "frequency":"weekly"}`),
		ID:      2,
		Type:    "copy-following",
	}

	type args struct {
		in database.FindCopyJobParams
	}

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
		out *models.CopyJob
	}

	tests := map[string]struct {
		args
		fields
		wants
	}{
		"followers - ok": {
			args{
				in: database.FindCopyJobParams{
					Direction: "followers",
					UserID:    123,
					WithPage:  nil,
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL1 := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE checksum = $1 AND job_type = $2`)

					expectedSQL2 := oneLineSQL(`SELECT COUNT(*) FROM user_followers WHERE account_id = $1`)

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL1, "copy-followers:123", "copy-followers").
						Return(mockCopyFollowersJob, nil)

					q.On("Count", ctx, mock.AnythingOfType("*database.Database"), expectedSQL2, int64(123)).
						Return(int32(10), nil)

					return q
				},
			},
			wants{
				out: &models.CopyJob{
					Job: mockCopyFollowersJob,
					Metadata: models.CopyJobMetadata{
						Frequency: "daily",
						UserID:    123,
					},
					Results: nil,
					Total:   10,
				},
			},
		},
		"following - ok": {
			args{
				in: database.FindCopyJobParams{
					Direction: "following",
					UserID:    456,
					WithPage:  nil,
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL1 := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE checksum = $1 AND job_type = $2`)

					expectedSQL2 := oneLineSQL(`SELECT COUNT(*) FROM user_following WHERE account_id = $1`)

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL1, "copy-following:456", "copy-following").
						Return(mockCopyFollowingJob, nil)

					q.On("Count", ctx, mock.AnythingOfType("*database.Database"), expectedSQL2, int64(456)).
						Return(int32(20), nil)

					return q
				},
			},
			wants{
				out: &models.CopyJob{
					Job: mockCopyFollowingJob,
					Metadata: models.CopyJobMetadata{
						Frequency: "weekly",
						UserID:    456,
					},
					Results: nil,
					Total:   20,
				},
			},
		},
		"followers with results - ok": {
			args{
				in: database.FindCopyJobParams{
					Direction: "followers",
					UserID:    123,
					WithPage:  intPtr(t, 4),
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL1 := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE checksum = $1 AND job_type = $2`)

					expectedSQL2 := oneLineSQL(`SELECT COUNT(*) FROM user_followers WHERE account_id = $1`)

					expectedSQL3 := oneLineSQL(`
					SELECT user_id, first_seen, handler, last_seen, pic_url
					FROM user_followers
					WHERE account_id = $1
					ORDER BY first_seen DESC LIMIT $2 OFFSET $3`)

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL1, "copy-followers:123", "copy-followers").
						Return(mockCopyFollowersJob, nil)

					q.On("Count", ctx, mock.AnythingOfType("*database.Database"), expectedSQL2, int64(123)).
						Return(int32(2), nil)

					q.On("SelectUsers", ctx, mock.AnythingOfType("*database.Database"), expectedSQL3, int64(123), 100, 400).
						Return([]models.User{
							{
								AccountID: 1,
								Handler:   "johndoe",
							},
							{
								AccountID: 2,
								Handler:   "janedoe",
							},
						}, nil)

					return q
				},
			},
			wants{
				out: &models.CopyJob{
					Job: mockCopyFollowersJob,
					Metadata: models.CopyJobMetadata{
						Frequency: "daily",
						UserID:    123,
					},
					Results: []models.User{
						{
							AccountID: 1,
							Handler:   "johndoe",
						},
						{
							AccountID: 2,
							Handler:   "janedoe",
						},
					},
					Total: 2,
				},
			},
		},
		"not found - ok": {
			args{
				in: database.FindCopyJobParams{
					Direction: "following",
					UserID:    1,
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					var j *models.Job

					expectedSQL1 := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE checksum = $1 AND job_type = $2`)

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL1, "copy-following:1", "copy-following").
						Return(j, pgx.ErrNoRows)

					return q
				},
			},
			wants{
				err: nil,
				out: nil,
			},
		},
		"invalid direction - err": {
			args{
				in: database.FindCopyJobParams{
					Direction: "fololo",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					return &mockQuerier{}
				},
			},
			wants{
				err: database.ErrFindCopyJobParams,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			job, err := db.FindCopyJob(ctx, test.args.in)

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, job)
		})
	}
}

func TestFindJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	mockJob := &models.Job{
		BinData: []byte(`{"dummy":true}`),
		ID:      1,
		Type:    "some-type",
	}

	type args struct {
		in database.FindJobParams
	}

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
		out *models.Job
	}

	tests := map[string]struct {
		args
		fields
		wants
	}{
		"all params - ok": {
			args{
				in: database.FindJobParams{
					Checksum: "job:checksum",
					ID:       123,
					Type:     "job-type",
					State:    "job-state",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE id = $1 AND checksum = $2 AND state = $3 AND job_type = $4`)

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, int64(123), "job:checksum", "job-state", "job-type").
						Return(mockJob, nil)

					return q
				},
			},
			wants{
				out: mockJob,
			},
		},
		"not found - ok": {
			args{
				in: database.FindJobParams{
					ID:    123,
					State: "job-state",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					var j *models.Job

					expectedSQL := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE id = $1 AND state = $2`)

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, int64(123), "job-state").
						Return(j, pgx.ErrNoRows)

					return q
				},
			},
			wants{
				err: nil,
				out: nil,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			job, err := db.FindJob(ctx, test.args.in)

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, job)
		})
	}
}

func TestFindJobs(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	mockErr := errors.New("mock error")

	mockJobs := []models.Job{
		{
			BinData: []byte(`{"dummy":true}`),
			ID:      1,
			Type:    "some-type",
		},
		{
			BinData: []byte(`{"dummy":true}`),
			ID:      2,
			Type:    "some-other-type",
		},
	}

	type args struct {
		in database.FindJobsParams
	}

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
		out []models.Job
	}

	tests := map[string]struct {
		args
		fields
		wants
	}{
		"order by last_run, desc - ok": {
			args{
				in: database.FindJobsParams{
					Order: "-last_run",
					Type:  "job-type",
					State: "job-state",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE state = $1 AND job_type = $2 ORDER BY last_run DESC LIMIT 20 OFFSET 0`)

					q := &mockQuerier{}

					q.On("SelectJobs", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "job-state", "job-type").
						Return(mockJobs, nil)

					return q
				},
			},
			wants{
				out: mockJobs,
			},
		},
		"order by last_run, asc - ok": {
			args{
				in: database.FindJobsParams{
					Order: "last_run",
					Type:  "thetype",
					State: "thestate",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE state = $1 AND job_type = $2 ORDER BY last_run ASC LIMIT 20 OFFSET 0`)

					q := &mockQuerier{}

					q.On("SelectJobs", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "thestate", "thetype").
						Return(mockJobs, nil)

					return q
				},
			},
			wants{
				out: mockJobs,
			},
		},
		"order by next_run, desc - ok": {
			args{
				in: database.FindJobsParams{
					Order: "-next_run",
					Type:  "job-type",
					State: "job-state",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE state = $1 AND job_type = $2 ORDER BY next_run DESC LIMIT 20 OFFSET 0`)

					q := &mockQuerier{}

					q.On("SelectJobs", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "job-state", "job-type").
						Return(mockJobs, nil)

					return q
				},
			},
			wants{
				out: mockJobs,
			},
		},
		"order by next_run, asc - ok": {
			args{
				in: database.FindJobsParams{
					Order: "next_run",
					Type:  "thetype",
					State: "thestate",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					WHERE state = $1 AND job_type = $2 ORDER BY next_run ASC LIMIT 20 OFFSET 0`)

					q := &mockQuerier{}

					q.On("SelectJobs", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "thestate", "thetype").
						Return(mockJobs, nil)

					return q
				},
			},
			wants{
				out: mockJobs,
			},
		},
		"no params and generic error": {
			args{
				in: database.FindJobsParams{},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`
					SELECT id, checksum, job_type, label, last_run, metadata, next_run, state
					FROM jobs
					ORDER BY last_run DESC LIMIT 20 OFFSET 0`)

					q := &mockQuerier{}

					q.On("SelectJobs", ctx, mock.AnythingOfType("*database.Database"), expectedSQL).
						Return([]models.Job{}, mockErr)

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

			job, err := db.FindJobs(ctx, test.args.in)

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, job)
		})
	}
}

func TestNewCopyJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	mockFollowersJob := &models.Job{
		BinData: []byte(`{"userID":111, "frequency":"weekly"}`),
		ID:      1,
		Type:    "copy-followers",
	}

	mockFollowersMetadata := struct {
		Cursor    string "json:\"-\""
		Frequency string "json:\"frequency\""
		UserID    int64  "json:\"userID\""
	}{
		Cursor:    "should be ignored",
		Frequency: "weekly",
		UserID:    111,
	}

	mockFollowingJob := &models.Job{
		BinData: []byte(`{"userID":222, "frequency":"daily"}`),
		ID:      2,
		Type:    "copy-following",
	}

	mockFollowingMetadata := struct {
		Cursor    string "json:\"-\""
		Frequency string "json:\"frequency\""
		UserID    int64  "json:\"userID\""
	}{
		Cursor:    "should be ignored",
		Frequency: "daily",
		UserID:    222,
	}

	type args struct {
		in database.NewCopyJobParams
	}

	type fields struct {
		querier func() *mockQuerier
	}

	type wants struct {
		err error
		out *models.CopyJob
	}

	tests := map[string]struct {
		args
		fields
		wants
	}{
		"new copy followers - ok": {
			args{
				in: database.NewCopyJobParams{
					Label:    "my label",
					NextRun:  nil,
					Metadata: mockFollowersMetadata,
					Type:     "copy-followers",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					var nextRun *time.Time

					expectedSQL := oneLineSQL(`
					INSERT INTO jobs ( checksum, job_type, label, last_run, metadata, next_run, state )
					VALUES ($1, $2, $3, NULL, $4, $5, $6)
					RETURNING *`)

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "copy-followers:111", "copy-followers", "my label", mockFollowersMetadata, nextRun, "new").
						Return(mockFollowersJob, nil)

					return q
				},
			},
			wants{
				out: &models.CopyJob{
					Job: mockFollowersJob,
					Metadata: models.CopyJobMetadata{
						Frequency: "weekly",
						UserID:    111,
					},
				},
			},
		},
		"new copy following - ok": {
			args{
				in: database.NewCopyJobParams{
					Label:    "my label",
					NextRun:  nil,
					Metadata: mockFollowingMetadata,
					Type:     "copy-following",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					var nextRun *time.Time

					expectedSQL := oneLineSQL(`
					INSERT INTO jobs ( checksum, job_type, label, last_run, metadata, next_run, state )
					VALUES ($1, $2, $3, NULL, $4, $5, $6)
					RETURNING *`)

					q := &mockQuerier{}

					q.On("SelectJob", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "copy-following:222", "copy-following", "my label", mockFollowingMetadata, nextRun, "new").
						Return(mockFollowingJob, nil)

					return q
				},
			},
			wants{
				out: &models.CopyJob{
					Job: mockFollowingJob,
					Metadata: models.CopyJobMetadata{
						Frequency: "daily",
						UserID:    222,
					},
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			job, err := db.NewCopyJob(ctx, test.args.in)

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, job)
		})
	}
}

// TestNewJob only tests for errors, as TestNewCopyJob already covers `NewJob()`.
func TestNewJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	type args struct {
		in database.NewJobParams
	}

	type wants struct {
		err error
	}

	tests := map[string]struct {
		args
		wants
	}{
		"invalid type - error": {
			args{
				in: database.NewJobParams{
					Type: "not a valid type",
				},
			},
			wants{
				err: database.ErrInvalidType,
			},
		},
		"invalid state - error": {
			args{
				in: database.NewJobParams{
					State: "not a valid state",
					Type:  "copy-followers",
				},
			},
			wants{
				err: database.ErrInvalidState,
			},
		},
		"blank checksum - error": {
			args{
				in: database.NewJobParams{
					Checksum: "",
					State:    "new",
					Type:     "copy-followers",
				},
			},
			wants{
				err: database.ErrInvalidChecksum,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(&mockQuerier{})

			job, err := db.NewJob(ctx, test.args.in)

			assert.ErrorIs(t, err, test.wants.err)
			assert.Nil(t, job)
		})
	}
}

func TestUpdateJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	type args struct {
		in database.UpdateJobParams
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
		"update all fields - ok": {
			args{
				in: database.UpdateJobParams{
					Frequency: "weekly",
					ID:        100,
					Label:     "my label",
					State:     "pause",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`
					UPDATE jobs SET
						metadata = jsonb_set(metadata, '{frequency}', to_jsonb($1::text)),state = $2,label = $3
					WHERE id = $4`)

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "weekly", "pause", "my label", int64(100)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
		"discard invalid fields - ok": {
			args{
				in: database.UpdateJobParams{
					Frequency: "wrong",
					ID:        100,
					Label:     "my label",
					State:     "wrong",
				},
			},
			fields{
				querier: func() *mockQuerier {
					t.Helper()

					expectedSQL := oneLineSQL(`
					UPDATE jobs SET label = $1
					WHERE id = $2`)

					q := &mockQuerier{}

					q.On("Execute", ctx, mock.AnythingOfType("*database.Database"), expectedSQL, "my label", int64(100)).
						Return(nil)

					return q
				},
			},
			wants{
				err: nil,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			q := test.fields.querier()
			db := database.NewPool(ctx, "postgres://user:pass@127.0.0.1:5432/db1").
				WithQuerier(q)

			err := db.UpdateJob(ctx, test.args.in)

			q.AssertExpectations(t)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.NoError(t, err)
		})
	}
}

func intPtr(t *testing.T, i int) *int {
	t.Helper()

	return &i
}
