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
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/luca-arch/instaman/webserver"
	"github.com/stretchr/testify/assert"
)

type args struct {
	endpoint string
	method   string
}

type wants struct {
	body   []byte
	status int
}

func TestEndpointsResponses(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.TODO())

	server, _ := webserver.Create(ctx, &jobsvc{}, &igservice{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	testServer := httptest.NewServer(server.Handler)

	t.Cleanup(testServer.Close)
	t.Cleanup(cancel)

	tests := map[string]struct {
		args
		wants
	}{
		"GET /instaman/instagram/me": {
			args{endpoint: "/instaman/instagram/me"},
			wants{
				body:   fixture(t, "testdata/instagram-me.json"),
				status: http.StatusOK,
			},
		},
		"GET /instaman/instagram/account/{name}": {
			args{endpoint: "/instaman/instagram/account/name"},
			wants{
				body:   fixture(t, "testdata/instagram-account-name.json"),
				status: http.StatusOK,
			},
		},
		"GET /instaman/instagram/account-id/{id}": {
			args{endpoint: "/instaman/instagram/account-id/123"},
			wants{
				body:   fixture(t, "testdata/instagram-account-id.json"),
				status: http.StatusOK,
			},
		},
		"GET /instaman/instagram/followers/{id}": {
			args{endpoint: "/instaman/instagram/followers/123"},
			wants{
				body:   fixture(t, "testdata/instagram-followers.json"),
				status: http.StatusOK,
			},
		},
		"GET /instaman/instagram/following/{id}": {
			args{endpoint: "/instaman/instagram/following/123"},
			wants{
				body:   fixture(t, "testdata/instagram-following.json"),
				status: http.StatusOK,
			},
		},
		"GET /instaman/jobs": {
			args{endpoint: "/instaman/jobs"},
			wants{
				body:   fixture(t, "testdata/jobs-job.json"),
				status: http.StatusOK,
			},
		},
		"GET /instaman/jobs/copy (followers)": {
			args{endpoint: "/instaman/jobs/copy?direction=followers&userID=123"},
			wants{
				body:   fixture(t, "testdata/jobs-copy.json"),
				status: http.StatusOK,
			},
		},
		"GET /instaman/jobs/copy (following)": {
			args{endpoint: "/instaman/jobs/copy?direction=following&userID=123"},
			wants{
				body:   fixture(t, "testdata/jobs-copy.json"),
				status: http.StatusOK,
			},
		},
		"GET /instaman/jobs/copy (error, no direction)": {
			args{endpoint: "/instaman/jobs/copy"},
			wants{
				body:   expectedErr(t, "missing required field: direction"),
				status: http.StatusBadRequest,
			},
		},
		"GET /instaman/jobs/copy (error, no user)": {
			args{endpoint: "/instaman/jobs/copy?direction=followers"},
			wants{
				body:   expectedErr(t, "missing required field: userID"),
				status: http.StatusBadRequest,
			},
		},
		"GET /instaman/jobs/all": {
			args{endpoint: "/instaman/jobs/all"},
			wants{
				body:   fixture(t, "testdata/jobs-all.json"),
				status: http.StatusOK,
			},
		},
		"POST /instaman/jobs/copy": {
			args{
				endpoint: "/instaman/jobs/copy",
				method:   http.MethodPost,
			},
			wants{
				body:   fixture(t, "testdata/jobs-copy-new.json"),
				status: http.StatusOK,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var (
				res *http.Response
				err error
			)

			//nolint:noctx // Ok when testing
			switch test.args.method {
			case http.MethodPost:
				// Empty body as the webserver's services are mocked in common_test.go.
				b := bytes.NewReader([]byte("{}"))
				//nolint:bodyclose // False positive.
				res, err = http.Post(testServer.URL+test.args.endpoint, "application/json", b)
			default:
				//nolint:bodyclose // False positive.
				res, err = http.Get(testServer.URL + test.args.endpoint)
			}

			assert.NoError(t, err)

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			res.Body.Close()

			assert.Equal(t, test.wants.status, res.StatusCode)
			assert.Equal(t, test.wants.body, body, "Actual: "+string(body))
		})
	}
}

func expectedErr(t *testing.T, msg string) []byte {
	t.Helper()

	b, err := json.Marshal(struct {
		Err string `json:"error"`
	}{
		Err: msg,
	})
	if err != nil {
		t.Fatal(err)
	}

	return append(b, byte(0xa)) // Append newline!
}
