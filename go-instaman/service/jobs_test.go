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

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/luca-arch/instaman/database"
	"github.com/luca-arch/instaman/database/models"
	"github.com/luca-arch/instaman/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var errMock = errors.New("mock error")

type mockDBJobs struct {
	mock.Mock
}

func (m *mockDBJobs) FindCopyJob(ctx context.Context, params database.FindCopyJobParams) (*models.CopyJob, error) {
	args := m.Called(ctx, params)

	return args.Get(0).(*models.CopyJob), args.Error(1)
}

func (m *mockDBJobs) FindJob(ctx context.Context, p database.FindJobParams) (*models.Job, error) {
	args := m.Called(ctx, p)

	return args.Get(0).(*models.Job), args.Error(1)
}

func (m *mockDBJobs) FindJobs(ctx context.Context, p database.FindJobsParams) ([]models.Job, error) {
	args := m.Called(ctx, p)

	return args.Get(0).([]models.Job), args.Error(1)
}

func (m *mockDBJobs) NewCopyJob(ctx context.Context, p database.NewCopyJobParams) (*models.CopyJob, error) {
	args := m.Called(ctx, p)

	return args.Get(0).(*models.CopyJob), args.Error(1)
}

func TestFindCopyJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	// Dummy params to assert FindCopyJob's specific arguments.
	params := database.FindCopyJobParams{
		Direction: "mock",
		UserID:    1,
	}

	type field struct {
		db func() *mockDBJobs
	}

	type wants struct {
		err error
		out *models.CopyJob
	}

	tests := map[string]struct {
		field
		wants
	}{
		"method FindCopyJob - ok": {
			field{
				db: func() *mockDBJobs {
					t.Helper()

					db := &mockDBJobs{}
					db.On("FindCopyJob", ctx, params).
						Return(&models.CopyJob{
							Job: &models.Job{
								ID:       123,
								Checksum: "abcde",
							},
						}, nil)

					return db
				},
			},
			wants{
				out: &models.CopyJob{
					Job: &models.Job{
						ID:       123,
						Checksum: "abcde",
					},
				},
			},
		},
		"method FindCopyJob - error": {
			field{
				db: func() *mockDBJobs {
					t.Helper()

					db := &mockDBJobs{}
					db.On("FindCopyJob", ctx, params).
						Return(&models.CopyJob{}, errMock)

					return db
				},
			},
			wants{
				err: errMock,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := service.NewJobsService(test.field.db())

			out, err := svc.FindCopyJob(ctx, params)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)
				assert.ErrorIs(t, err, service.ErrDBFailure)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, out)
		})
	}
}

func TestFindJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	// Dummy params to assert FindJob's specific arguments.
	params := database.FindJobParams{
		Checksum: "mock checksum",
		ID:       1,
	}

	type field struct {
		db func() *mockDBJobs
	}

	type wants struct {
		err error
		out *models.Job
	}

	tests := map[string]struct {
		field
		wants
	}{
		"method FindJob - ok": {
			field{
				db: func() *mockDBJobs {
					t.Helper()

					db := &mockDBJobs{}
					db.On("FindJob", ctx, params).
						Return(&models.Job{
							ID:       456,
							Checksum: "abcde",
						}, nil)

					return db
				},
			},
			wants{
				out: &models.Job{
					ID:       456,
					Checksum: "abcde",
				},
			},
		},
		"method FindJob - error": {
			field{
				db: func() *mockDBJobs {
					t.Helper()

					db := &mockDBJobs{}
					db.On("FindJob", ctx, params).
						Return(&models.Job{}, errMock)

					return db
				},
			},
			wants{
				err: errMock,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := service.NewJobsService(test.field.db())

			out, err := svc.FindJob(ctx, params)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)
				assert.ErrorIs(t, err, service.ErrDBFailure)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, out)
		})
	}
}

func TestFindJobs(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	// Dummy params to assert FindJob's specific arguments.
	params := database.FindJobsParams{
		Order: "order",
		Page:  1,
		State: "status",
	}

	type field struct {
		db func() *mockDBJobs
	}

	type wants struct {
		err error
		out []models.Job
	}

	tests := map[string]struct {
		field
		wants
	}{
		"method FindJobs - ok": {
			field{
				db: func() *mockDBJobs {
					t.Helper()

					db := &mockDBJobs{}
					db.On("FindJobs", ctx, params).
						Return([]models.Job{
							{
								ID:       123,
								Checksum: "abcde",
							},
							{
								ID:       456,
								Checksum: "wxyz",
							},
						}, nil)

					return db
				},
			},
			wants{
				out: []models.Job{
					{
						ID:       123,
						Checksum: "abcde",
					},
					{
						ID:       456,
						Checksum: "wxyz",
					},
				},
			},
		},
		"method FindJobs - error": {
			field{
				db: func() *mockDBJobs {
					t.Helper()

					db := &mockDBJobs{}
					db.On("FindJobs", ctx, params).
						Return([]models.Job{}, errMock)

					return db
				},
			},
			wants{
				err: errMock,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := service.NewJobsService(test.field.db())

			out, err := svc.FindJobs(ctx, params)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)
				assert.ErrorIs(t, err, service.ErrDBFailure)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, out)
		})
	}
}

func TestNewCopyJob(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	// Dummy params to assert NewCopyJob's specific arguments.
	params := database.NewCopyJobParams{
		Label: "test label",
		Type:  "test job type",
	}

	type field struct {
		db func() *mockDBJobs
	}

	type wants struct {
		err error
		out *models.CopyJob
	}

	tests := map[string]struct {
		field
		wants
	}{
		"method NewCopyJob - ok": {
			field{
				db: func() *mockDBJobs {
					t.Helper()

					db := &mockDBJobs{}
					db.On("NewCopyJob", ctx, params).
						Return(&models.CopyJob{
							Job: &models.Job{
								ID:       123,
								Checksum: "abcde",
							},
						}, nil)

					return db
				},
			},
			wants{
				out: &models.CopyJob{
					Job: &models.Job{
						ID:       123,
						Checksum: "abcde",
					},
				},
			},
		},
		"method NewCopyJob - error": {
			field{
				db: func() *mockDBJobs {
					t.Helper()

					db := &mockDBJobs{}
					db.On("NewCopyJob", ctx, params).
						Return(&models.CopyJob{}, errMock)

					return db
				},
			},
			wants{
				err: errMock,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			svc := service.NewJobsService(test.field.db())

			out, err := svc.NewCopyJob(ctx, params)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)
				assert.ErrorIs(t, err, service.ErrDBFailure)

				return
			}

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, out)
		})
	}
}
