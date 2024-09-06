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

package models

const (
	JobFrequencyDaily    = "daily"
	JobFrequencyWeekly   = "weekly"
	JobStateActive       = "active"
	JobStateError        = "error"
	JobStateNew          = "new"
	JobStatePaused       = "pause"
	JobTypeCopyFollowers = "copy-followers"
	JobTypeCopyFollowing = "copy-following"
)

// IsValidJobFrequency return whether job frequency is a valid value for the jobs.metadata ->> frequency column.
func IsValidJobFrequency(jobFreq string) bool {
	switch jobFreq {
	case JobFrequencyDaily, JobFrequencyWeekly:
		return true
	default:
		return false
	}
}

// IsValidJobState return whether state is a valid value for the jobs.state column.
func IsValidJobState(jobType string) bool {
	switch jobType {
	case JobStateActive, JobStateError, JobStateNew, JobStatePaused:
		return true
	default:
		return false
	}
}

// IsValidJobType return whether jobType is a valid value for the jobs.job_type column.
func IsValidJobType(jobType string) bool {
	switch jobType {
	case JobTypeCopyFollowers, JobTypeCopyFollowing:
		return true
	default:
		return false
	}
}
