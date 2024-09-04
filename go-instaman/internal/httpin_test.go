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

package internal_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luca-arch/instaman/internal"
	"github.com/stretchr/testify/assert"
)

type StructInt struct {
	IntNum   int   `in:"val"`
	Int16Num int16 `in:"val16"`
	Int32Num int32 `in:"val32"`
	Int64Num int64 `in:"val64"`
}

type StructPtr struct {
	IntNum   *int    `in:"val"`
	Int16Num *int16  `in:"val16"`
	Int32Num *int32  `in:"val32"`
	Int64Num *int64  `in:"val64"`
	String   *string `in:"valStr"`
}

type StructRequired struct {
	Param string `in:"sentence,required"`
}

func TestInputFromRequest(t *testing.T) {
	t.Parallel()

	var (
		intNum         = 10
		int16Num int16 = 20
		int32Num int32 = 30
		int64Num int64 = 40
		strVal         = "my string"
	)

	type args struct {
		url string
	}

	type fields struct {
		call func(*http.Request) (any, error)
	}

	type wants struct {
		err string
		out any
	}

	tests := map[string]struct {
		args
		fields
		wants
	}{
		"Struct with numeric types": {
			args{
				url: "https://example.com/?val=10&val16=20&val32=30&val64=40",
			},
			fields{
				call: func(r *http.Request) (any, error) {
					return internal.InputFromRequest[StructInt](r)
				},
			},
			wants{
				out: StructInt{
					IntNum:   intNum,
					Int16Num: int16Num,
					Int32Num: int32Num,
					Int64Num: int64Num,
				},
			},
		},
		"ok - struct with required value": {
			args{
				url: "https://example.com/?sentence=my+string",
			},
			fields{
				call: func(r *http.Request) (any, error) {
					return internal.InputFromRequest[StructRequired](r)
				},
			},
			wants{
				out: StructRequired{
					Param: strVal,
				},
			},
		},
		"ok - struct with pointers": {
			args{
				url: "https://example.com/?val=10&val32=30&val64=40&valStr=my+string",
			},
			fields{
				call: func(r *http.Request) (any, error) {
					return internal.InputFromRequest[StructPtr](r)
				},
			},
			wants{
				out: StructPtr{
					IntNum:   &intNum,
					Int16Num: nil,
					Int32Num: &int32Num,
					Int64Num: &int64Num,
					String:   &strVal,
				},
			},
		},
		"error - struct with required value": {
			args{
				url: "https://example.com/",
			},
			fields{
				call: func(r *http.Request) (any, error) {
					return internal.InputFromRequest[StructRequired](r)
				},
			},
			wants{
				err: "missing required field: sentence",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			r := httptest.NewRequest(http.MethodGet, test.args.url, nil)

			out, err := test.fields.call(r)

			if test.wants.err != "" {
				assert.EqualError(t, err, test.wants.err)

				return
			}

			assert.Equal(t, test.wants.out, out)
		})
	}
}
