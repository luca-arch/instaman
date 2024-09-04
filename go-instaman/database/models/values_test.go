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

const (
	JobTypeCopyFollowers = "copy-followers"
	JobTypeCopyFollowing = "copy-following"
)

func TestIsValidJobType(t *testing.T) {
	t.Parallel()

	type args struct {
		in string
	}

	type wants struct {
		out bool
	}

	tests := map[string]struct {
		args
		wants
	}{
		"valid - copy-followers": {
			args{
				in: "copy-followers",
			},
			wants{
				out: true,
			},
		},
		"valid - copy-following": {
			args{
				in: "copy-following",
			},
			wants{
				out: true,
			},
		},
		"invalid - blank": {
			args{
				in: "",
			},
			wants{
				out: false,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, test.out, models.IsValidJobType(test.in))
		})
	}
}
