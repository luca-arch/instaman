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

// Package models describes the structures stored in the database.
package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrInvalidUserID   = errors.New("invalid user ID")
	ErrInvalidMetadata = errors.New("job has invalid metadata")
	ErrInvalidCopy     = errors.New("not a CopyJob")
)

// CopyJob represents a record of the `jobs` table of which the type is either `copy-followers` or `copy-following`.
type CopyJob struct {
	*Job

	Metadata CopyJobMetadata `json:"metadata"`
	Results  []User          `json:"results"`
	Total    int32           `json:"resultsCount"`
}

// CopyJobMetadata.
type CopyJobMetadata struct {
	Cursor    *string `json:"cursor,omitempty"`
	Frequency string  `json:"frequency"`
	UserID    int64   `json:"userID"` //nolint:tagliatelle // Always capitalise ID suffix.
}

// Job represents a record of the `jobs` table.
type Job struct {
	BinData  []byte     `description:"Job's metadata as binary stream" json:"metadata" db:"metadata"`
	ID       int64      `description:"Record PK" json:"id" db:"id"`
	Checksum string     `description:"Job checksum to avoid duplicates" json:"checksum" db:"checksum"`
	Type     string     `description:"Job type (copy-followers, copy-following)" json:"type" db:"job_type"`
	Label    string     `description:"Human readable label" json:"label" db:"label"`
	LastRun  *time.Time `description:"Last execution time" json:"lastRun" db:"last_run"`
	NextRun  *time.Time `description:"Next scheduled time" json:"nextRun" db:"next_run"`
	State    string     `description:"Execution's state (active, error, new, pause)" json:"state" db:"state"`
}

// User represents an Instagram user as stored in the `user_followers` and `user_following` tables.
type User struct {
	AccountID  int64     `description:"Account ID (relationship owner)" json:"-" db:"account_id"`
	ID         int64     `description:"User's Instagram ID" json:"id" db:"user_id"`
	FirstSeen  time.Time `description:"First time the connection was indexed" json:"firstSeen" db:"first_seen"`
	Handler    string    `description:"User's Instagram handler" json:"handler" db:"handler"`
	LastSeen   time.Time `description:"Last time the connection was indexed" json:"lastSeen" db:"last_seen"`
	PictureURL *string   `description:"Profile picture URL" json:"pictureURL" db:"pic_url"` //nolint:tagliatelle // Make it consistent
}

// NewCopyJob morphs a Job into a CopyJob validating its metadata.
// This factory is required to avoid a Metadata field of type of `map[string]any` and its bizarre behaviour with int64 being converted to float64.
func NewCopyJob(j *Job) (*CopyJob, error) {
	var m *CopyJobMetadata

	if j.Type != JobTypeCopyFollowers && j.Type != JobTypeCopyFollowing {
		return nil, ErrInvalidCopy
	}

	// Use an encoder with `Number()` so long integers are correctly parsed.
	d := json.NewDecoder(bytes.NewBuffer(j.BinData))
	d.UseNumber()

	if err := d.Decode(&m); err != nil {
		return nil, errors.Join(ErrInvalidMetadata, err)
	}

	if m.UserID < 1 {
		return nil, ErrInvalidUserID
	}

	if m.Cursor != nil && *m.Cursor == "" {
		m.Cursor = nil
	}

	if !IsValidJobFrequency(m.Frequency) {
		m.Frequency = JobFrequencyDaily
	}

	return &CopyJob{
		Job:      j,
		Metadata: *m,
		Results:  nil,
		Total:    0,
	}, nil
}
