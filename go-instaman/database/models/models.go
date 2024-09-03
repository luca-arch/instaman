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
	"time"
)

// User represents an Instagram user as stored in the `user_followers` and `user_following` tables.
type User struct {
	AccountID  int64     `description:"Account ID (relationship owner)" json:"-" db:"account_id"`
	ID         int64     `description:"User's Instagram ID" json:"id" db:"user_id"`
	FirstSeen  time.Time `description:"First time the connection was indexed" json:"firstSeen" db:"first_seen"`
	Handler    string    `description:"User's Instagram handler" json:"handler" db:"handler"`
	LastSeen   time.Time `description:"Last time the connection was indexed" json:"lastSeen" db:"last_seen"`
	PictureURL *string   `description:"Profile picture URL" json:"pictureURL" db:"pic_url"` //nolint:tagliatelle // Make it consistent
}

// CopyJob represents a record of the `jobs` table of which the type is either `copy-followers` or `copy-following`.
type CopyJob struct {
	*Job

	Results []User `json:"results"`
	Total   int32  `json:"resultsCount"`
}

// Job represents a record of the `jobs` table.
type Job struct {
	ID       int64          `description:"Record PK" json:"id" db:"id"`
	Checksum string         `description:"Job checksum to avoid duplicates" json:"checksum" db:"checksum"`
	Type     string         `description:"Job type (copy-followers, copy-following)" json:"type" db:"job_type"`
	Label    string         `description:"Human readable label" json:"label" db:"label"`
	LastRun  *time.Time     `description:"Last execution time" json:"lastRun" db:"last_run"`
	Metadata map[string]any `description:"Job's metadata" json:"metadata" db:"metadata"`
	NextRun  *time.Time     `description:"Next scheduled time" json:"nextRun" db:"next_run"`
	State    string         `description:"Execution's state (new, paused, suspended)" json:"state" db:"state"`
}
