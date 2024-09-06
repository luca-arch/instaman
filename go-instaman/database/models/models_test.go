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

package models_test

import (
	"testing"

	"github.com/luca-arch/instaman/database/models"
	"github.com/stretchr/testify/assert"
)

func TestNewCopyJob(t *testing.T) {
	t.Parallel()

	type args struct {
		in  string
		typ string
	}

	type wants struct {
		err error
		out *models.CopyJobMetadata
	}

	tests := map[string]struct {
		args
		wants
	}{
		"invalid - blank": {
			args{
				in:  "",
				typ: "copy-followers",
			},
			wants{
				err: models.ErrInvalidMetadata,
				out: nil,
			},
		},
		"invalid - negative user id": {
			args{
				in:  `{"userID": -1}`,
				typ: "copy-followers",
			},
			wants{
				err: models.ErrInvalidUserID,
				out: nil,
			},
		},
		"invalid - no user id": {
			args{
				in:  `{}`,
				typ: "copy-followers",
			},
			wants{
				err: models.ErrInvalidUserID,
				out: nil,
			},
		},
		"invalid - wrong type": {
			args{
				in:  "{}",
				typ: "copy-something",
			},
			wants{
				err: models.ErrInvalidCopy,
				out: nil,
			},
		},
		"invalid - zero user id": {
			args{
				in:  `{"userID": 0}`,
				typ: "copy-following",
			},
			wants{
				err: models.ErrInvalidUserID,
				out: nil,
			},
		},
		"valid - with default frequency": {
			args{
				in:  `{"userID": 1}`,
				typ: "copy-following",
			},
			wants{
				out: &models.CopyJobMetadata{
					Cursor:    nil,
					Frequency: "daily",
					UserID:    1,
				},
			},
		},
		"valid - with normalised frequency": {
			args{
				in:  `{"frequency":"wrong", "userID":1}`,
				typ: "copy-following",
			},
			wants{
				out: &models.CopyJobMetadata{
					Cursor:    nil,
					Frequency: "daily",
					UserID:    1,
				},
			},
		},
		"valid - with cursor": {
			args{
				in:  `{"cursor":"abcdefg", "userID":1}`,
				typ: "copy-following",
			},
			wants{
				out: &models.CopyJobMetadata{
					Cursor:    strPtr(t, "abcdefg"),
					Frequency: "daily",
					UserID:    1,
				},
			},
		},
		"valid - with empty cursor": {
			args{
				in:  `{"cursor":"", "userID":1}`,
				typ: "copy-following",
			},
			wants{
				out: &models.CopyJobMetadata{
					Cursor:    nil,
					Frequency: "daily",
					UserID:    1,
				},
			},
		},
		"valid - with null cursor": {
			args{
				in:  `{"cursor":null, "userID":1}`,
				typ: "copy-following",
			},
			wants{
				out: &models.CopyJobMetadata{
					Cursor:    nil,
					Frequency: "daily",
					UserID:    1,
				},
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			job := &models.Job{
				BinData: []byte(test.args.in),
				ID:      123,
				Type:    test.args.typ,
			}

			jc, err := models.NewCopyJob(job)

			if test.wants.err != nil {
				assert.ErrorIs(t, err, test.wants.err)

				return
			}

			assert.Equal(t, int64(123), jc.ID)
			assert.Equal(t, test.wants.out, &jc.Metadata)
		})
	}
}

func strPtr(t *testing.T, str string) *string {
	t.Helper()

	return &str
}
