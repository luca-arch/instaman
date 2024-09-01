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
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/luca-arch/instaman/instaproxy"
	"github.com/luca-arch/instaman/webserver"
	"github.com/stretchr/testify/assert"
)

type client struct{}

func (c *client) GetAccount(_ context.Context) (*instaproxy.Account, error) {
	picURL, _ := url.Parse("https://example.com/avatar.png")

	return &instaproxy.Account{
		Biography:  "account bio",
		FullName:   "John Doe",
		Handler:    "john_doe",
		ID:         123,
		PictureURL: &instaproxy.URLField{URL: *picURL},
	}, nil
}

func (c *client) GetFollowers(_ context.Context, _ int64, _ *string) (*instaproxy.Connections, error) {
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

func (c *client) GetFollowing(_ context.Context, _ int64, _ *string) (*instaproxy.Connections, error) {
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

func (c *client) GetUser(_ context.Context, _ string) (*instaproxy.User, error) {
	picURL, _ := url.Parse("https://example.com/user.png")

	return &instaproxy.User{
		FullName:   "User Name",
		Handler:    "user_name",
		ID:         123,
		PictureURL: &instaproxy.URLField{URL: *picURL},
	}, nil
}

func (c *client) GetUserByID(_ context.Context, _ int64) (*instaproxy.User, error) {
	picURL, _ := url.Parse("https://example.com/user.png")

	return &instaproxy.User{
		FullName:   "User Name",
		Handler:    "user_name",
		ID:         456,
		PictureURL: &instaproxy.URLField{URL: *picURL},
	}, nil
}

func TestInstagramEndpoints(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.TODO())

	server, _ := webserver.Create(ctx, &client{}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	testServer := httptest.NewServer(server.Handler)

	t.Cleanup(testServer.Close)
	t.Cleanup(cancel)

	type args struct {
		endpoint string
	}

	type wants struct {
		body   []byte
		status int
	}

	tests := map[string]struct {
		args
		wants
	}{
		"GET /instagram/me": {
			args{endpoint: "/instagram/me"},
			wants{
				body:   fixture(t, "testdata/instagram-me.json"),
				status: http.StatusOK,
			},
		},
		"GET /instagram/account/{name}": {
			args{endpoint: "/instagram/account/name"},
			wants{
				body:   fixture(t, "testdata/instagram-account-name.json"),
				status: http.StatusOK,
			},
		},
		"GET /instagram/account-id/{id}": {
			args{endpoint: "/instagram/account-id/123"},
			wants{
				body:   fixture(t, "testdata/instagram-account-id.json"),
				status: http.StatusOK,
			},
		},
		"GET /instagram/followers/{id}": {
			args{endpoint: "/instagram/followers/123"},
			wants{
				body:   fixture(t, "testdata/instagram-followers.json"),
				status: http.StatusOK,
			},
		},
		"GET /instagram/following/{id}": {
			args{endpoint: "/instagram/following/123"},
			wants{
				body:   fixture(t, "testdata/instagram-following.json"),
				status: http.StatusOK,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			//nolint:noctx // Ok when testing
			res, err := http.Get(testServer.URL + test.args.endpoint)
			assert.NoError(t, err)

			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			res.Body.Close()

			assert.Equal(t, test.wants.status, res.StatusCode)
			assert.Equal(t, test.wants.body, body, "Actual: "+string(body))
		})
	}
}
