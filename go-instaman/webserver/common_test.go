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

package webserver_test

import (
	"context"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/luca-arch/instaman/database"
	"github.com/luca-arch/instaman/database/models"
	"github.com/luca-arch/instaman/instaproxy"
	"github.com/luca-arch/instaman/service"
)

// igservice implements webserver.igservice.
type igservice struct{}

func (c *igservice) GetAccount(_ context.Context) (*instaproxy.Account, error) {
	picURL, _ := url.Parse("https://example.com/avatar.png")

	return &instaproxy.Account{
		Biography:  "account bio",
		FullName:   "John Doe",
		Handler:    "john_doe",
		ID:         123,
		PictureURL: &instaproxy.URLField{URL: *picURL},
	}, nil
}

func (c *igservice) GetFollowers(_ context.Context, _ service.GetConnectionInput) (*instaproxy.Connections, error) {
	picURL0, _ := url.Parse("https://example.com/avatar-0.png")
	picURL1, _ := url.Parse("https://example.com/avatar-1.png")

	return &instaproxy.Connections{
		Next: strPtr("next-cursor-001"),
		Users: []instaproxy.User{
			{
				FullName:   "John Doe",
				Handler:    "johndoe",
				ID:         12,
				PictureURL: &instaproxy.URLField{URL: *picURL0},
			},
			{
				FullName:   "Jane Doe",
				Handler:    "janedoe",
				ID:         23,
				PictureURL: &instaproxy.URLField{URL: *picURL1},
			},
			{
				FullName:   "Doe John",
				Handler:    "doejohn",
				ID:         34,
				PictureURL: &instaproxy.URLField{URL: *picURL0},
			},
			{
				FullName:   "Doe Jane",
				Handler:    "doejane",
				ID:         45,
				PictureURL: &instaproxy.URLField{URL: *picURL1},
			},
		},
	}, nil
}

func (c *igservice) GetFollowing(_ context.Context, _ service.GetConnectionInput) (*instaproxy.Connections, error) {
	picURL0, _ := url.Parse("https://example.com/avatar-2.png")
	picURL1, _ := url.Parse("https://example.com/avatar-3.png")

	return &instaproxy.Connections{
		Next: strPtr("next-cursor-002"),
		Users: []instaproxy.User{
			{
				FullName:   "John Doe",
				Handler:    "johndoe",
				ID:         45,
				PictureURL: &instaproxy.URLField{URL: *picURL0},
			},
			{
				FullName:   "Jane Doe",
				Handler:    "janedoe",
				ID:         56,
				PictureURL: &instaproxy.URLField{URL: *picURL1},
			},
			{
				FullName:   "Doe John",
				Handler:    "doejohn",
				ID:         67,
				PictureURL: &instaproxy.URLField{URL: *picURL0},
			},
			{
				FullName:   "Doe Jane",
				Handler:    "doejane",
				ID:         78,
				PictureURL: &instaproxy.URLField{URL: *picURL1},
			},
		},
	}, nil
}

func (c *igservice) GetUser(_ context.Context, _ service.GetUserInput) (*instaproxy.User, error) {
	picURL, _ := url.Parse("https://example.com/user.png")

	return &instaproxy.User{
		FullName:   "User Name",
		Handler:    "user_name",
		ID:         123,
		PictureURL: &instaproxy.URLField{URL: *picURL},
	}, nil
}

func (c *igservice) GetUserByID(_ context.Context, _ service.GetUserByIDInput) (*instaproxy.User, error) {
	picURL, _ := url.Parse("https://example.com/user.png")

	return &instaproxy.User{
		FullName:   "User Name",
		Handler:    "user_name",
		ID:         456,
		PictureURL: &instaproxy.URLField{URL: *picURL},
	}, nil
}

// jobsvc implements webserver.jobservice.
type jobsvc struct{}

func (j *jobsvc) FindCopyJob(context.Context, database.FindCopyJobParams) (*models.CopyJob, error) {
	t, err := time.Parse(time.RFC3339, "2025-01-01T12:00:00Z")
	if err != nil {
		panic(err)
	}

	return &models.CopyJob{
		Job: &models.Job{
			ID:       123,
			Checksum: "test:123456",
			Type:     "jobtype",
			Label:    "Test label",
			LastRun:  &t,
			NextRun:  &t,
			State:    "paused",
		},
		Results: []models.User{},
		Total:   0,
	}, nil
}

func (j *jobsvc) FindJob(context.Context, database.FindJobParams) (*models.Job, error) {
	t, err := time.Parse(time.RFC3339, "2026-01-01T12:00:00Z")
	if err != nil {
		panic(err)
	}

	return &models.Job{
		ID:       456,
		Checksum: "test:abcdef",
		Type:     "jobtype",
		Label:    "Test job",
		LastRun:  &t,
		NextRun:  &t,
		State:    "suspended",
	}, nil
}

func (j *jobsvc) FindJobs(context.Context, database.FindJobsParams) ([]models.Job, error) {
	t, err := time.Parse(time.RFC3339, "2026-01-01T12:00:00Z")
	if err != nil {
		panic(err)
	}

	return []models.Job{
		{
			ID:       123,
			Checksum: "test:123456",
			Type:     "jobtype",
			Label:    "Test label",
			LastRun:  nil,
			NextRun:  nil,
			State:    "paused",
		},
		{
			ID:       456,
			Checksum: "test:abcdef",
			Type:     "jobtype",
			Label:    "Test job",
			LastRun:  &t,
			NextRun:  &t,
			State:    "suspended",
		},
	}, nil
}

func (j *jobsvc) NewCopyJob(context.Context, database.NewCopyJobParams) (*models.CopyJob, error) {
	t, err := time.Parse(time.RFC3339, "2025-01-01T12:00:00Z")
	if err != nil {
		panic(err)
	}

	return &models.CopyJob{
		Job: &models.Job{
			ID:       123,
			Checksum: "test:123456",
			Type:     "jobtype",
			Label:    "Test label",
			LastRun:  nil,
			NextRun:  &t,
			State:    "new",
		},
		Results: []models.User{},
		Total:   0,
	}, nil
}

func fixture(t *testing.T, path string) []byte {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	return data
}

func strPtr(s string) *string {
	return &s
}
