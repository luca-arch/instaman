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

package instaproxy_test

import (
	"testing"

	"github.com/luca-arch/instaman/instaproxy"
	"github.com/stretchr/testify/assert"
)

func TestURLFieldMarshal(t *testing.T) {
	t.Parallel()

	type fields struct {
		obj *instaproxy.URLField
	}

	type wants struct {
		err error
		out string
	}

	tests := map[string]struct {
		fields
		wants
	}{
		"success": {
			fields{
				obj: urlField(t, "https://example.com/"),
			},
			wants{
				out: `"https://example.com/"`,
			},
		},
		"error - relative URL": {
			fields{
				obj: urlField(t, "/path/user.png"),
			},
			wants{
				err: instaproxy.ErrInvalidPictureURL,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			data, err := test.fields.obj.MarshalJSON()

			switch {
			case test.wants.err != nil:
				assert.ErrorIs(t, err, test.wants.err)
			default:
				assert.Equal(t, test.wants.out, string(data))
				assert.Nil(t, err)
			}
		})
	}
}

func TestURLFieldUnmarshal(t *testing.T) {
	t.Parallel()

	type args struct {
		in []byte
	}

	type wants struct {
		empty bool
		err   error
		out   string
	}

	tests := map[string]struct {
		args
		wants
	}{
		"ok - HTTP URL": {
			args{
				in: []byte(`"http://example.com"`),
			},
			wants{
				out: "http://example.com",
			},
		},
		"ok - HTTPS URL": {
			args{
				in: []byte(`"https://example.com"`),
			},
			wants{
				out: "https://example.com",
			},
		},
		"err - invalid URL": {
			args{
				in: []byte(`"not an url"`),
			},
			wants{
				err: instaproxy.ErrInvalidPictureURL,
			},
		},
		"err - protocol": {
			args{
				in: []byte(`"://example.com"`),
			},
			wants{
				err: instaproxy.ErrInvalidPictureURL,
			},
		},
		"err - relative URL": {
			args{
				in: []byte(`"/example"`),
			},
			wants{
				err: instaproxy.ErrInvalidPictureURL,
			},
		},
		"ok - empty bytes": {
			args{
				in: []byte(""),
			},
			wants{
				empty: true,
			},
		},
		"ok - empty URL": {
			args{
				in: []byte(`""`),
			},
			wants{
				empty: true,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			u := &instaproxy.URLField{}

			err := u.UnmarshalJSON(test.args.in)

			switch {
			case test.wants.err != nil:
				assert.ErrorIs(t, err, test.wants.err)
				assert.Empty(t, u)
			case test.wants.empty:
				assert.Nil(t, err)
				assert.Empty(t, u)
			default:
				assert.Nil(t, err)
				assert.Equal(t, test.wants.out, u.String())
			}
		})
	}
}
