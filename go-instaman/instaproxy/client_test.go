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
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/luca-arch/instaman/instaproxy"
	"github.com/stretchr/testify/assert"
)

type httpDoer struct {
	httpGet func(*http.Request) (*http.Response, error)
}

func (h *httpDoer) Do(r *http.Request) (*http.Response, error) {
	return h.httpGet(r)
}

func mockHTTPDoer(t *testing.T, expectedURL, respStubPath string) *httpDoer {
	t.Helper()

	body := fixture(t, respStubPath)

	h := new(httpDoer)

	h.httpGet = func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, expectedURL, req.URL.String())

		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer(body)),
			Status:     fmt.Sprintf("%d %s", http.StatusOK, http.StatusText(http.StatusOK)),
			StatusCode: http.StatusOK,
		}, nil
	}

	return h
}

func mockErrorDoer(t *testing.T, status int, err error) *httpDoer {
	t.Helper()

	h := new(httpDoer)

	h.httpGet = func(_ *http.Request) (*http.Response, error) {
		if err != nil {
			return nil, err
		}

		return &http.Response{
			Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
			Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
			StatusCode: status,
		}, nil
	}

	return h
}

func TestBaseURL(t *testing.T) {
	t.Parallel()

	type args struct {
		baseURL string
	}

	type wants struct {
		err error
	}

	tests := map[string]struct {
		args
		wants
	}{
		"error - invalid protocol": {
			args{
				baseURL: "//backend:8000",
			},
			wants{
				err: instaproxy.ErrNoProtocol,
			},
		},
		"error - invalid URL": {
			args{
				baseURL: ":smile:",
			},
			wants{
				err: instaproxy.ErrInvalidURL,
			},
		},
		"ok": {
			args{
				baseURL: instaproxy.DefaultBaseURL,
			},
			wants{
				err: nil,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := instaproxy.NewClient(&httpDoer{}, nil)

			err := client.BaseURL(test.args.baseURL)

			switch {
			case test.wants.err == nil:
				assert.NoError(t, err)
			default:
				assert.ErrorIs(t, err, test.wants.err)
			}
		})
	}
}

func TestGetErrors(t *testing.T) {
	t.Parallel()

	type fields struct {
		httpDoer *httpDoer
	}

	type wants struct {
		err error
	}

	tests := map[string]struct {
		fields
		wants
	}{
		"client receives 404": {
			fields{
				httpDoer: mockErrorDoer(t, http.StatusNotFound, nil),
			},
			wants{
				err: instaproxy.ErrNotFound,
			},
		},
		"client receives 429": {
			fields{
				httpDoer: mockErrorDoer(t, http.StatusTooManyRequests, nil),
			},
			wants{
				err: instaproxy.ErrInvalidStatus,
			},
		},
		"network failure": {
			fields{
				httpDoer: mockErrorDoer(t, 0, errors.New("broken")),
			},
			wants{
				err: instaproxy.ErrHTTPFailure,
			},
		},
		"invalid json": {
			fields{
				httpDoer: mockErrorDoer(t, http.StatusOK, nil),
			},
			wants{
				err: instaproxy.ErrInvalidJSON,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := instaproxy.NewClient(test.fields.httpDoer, nil)

			// Call `GetAccount()` to test `get()`.
			out, err := client.GetAccount(context.TODO())

			assert.ErrorIs(t, err, test.wants.err)
			assert.Nil(t, out)
		})
	}
}

func TestMethods(t *testing.T) {
	t.Parallel()

	stubUser := &instaproxy.User{
		FullName:   "John Doe",
		Handler:    "john_doe",
		ID:         123,
		PictureURL: urlField(t, "https://example.com/pic.png"),
	}

	stubUsers := []instaproxy.User{
		{
			FullName: "John Doe",
			Handler:  "johndoe",
			ID:       45,
		},
		{
			FullName: "Jane Doe",
			Handler:  "janedoe",
			ID:       56,
		},
		{
			FullName: "Name Surname",
			Handler:  "name_surname",
			ID:       67,
		},
	}

	type fields struct {
		callMethod func(*instaproxy.Client) (any, error)
		httpDoer   *httpDoer
	}

	type wants struct {
		out any
	}

	tests := map[string]struct {
		fields
		wants
	}{
		"GetAccount": {
			fields{
				callMethod: func(c *instaproxy.Client) (any, error) {
					return c.GetAccount(context.TODO())
				},
				httpDoer: mockHTTPDoer(t, instaproxy.DefaultBaseURL+"/me", "testdata/me.json"),
			},
			wants{
				out: &instaproxy.Account{
					Biography:  "my account bio",
					FullName:   "Test Account",
					Handler:    "test_account",
					ID:         123,
					PictureURL: urlField(t, "https://example.com/pic.png"),
				},
			},
		},
		"GetFollowers (paginated)": {
			fields{
				callMethod: func(c *instaproxy.Client) (any, error) {
					return c.GetFollowers(context.TODO(), int64(1234), strPtr(t, "abcdef"))
				},
				httpDoer: mockHTTPDoer(t, instaproxy.DefaultBaseURL+"/followers/1234?next_cursor=abcdef", "testdata/followers.json"),
			},
			wants{
				out: &instaproxy.Connections{
					Next:  strPtr(t, "wxyz123"),
					Users: stubUsers,
				},
			},
		},
		"GetFollowers": {
			fields{
				callMethod: func(c *instaproxy.Client) (any, error) {
					return c.GetFollowers(context.TODO(), int64(456), nil)
				},
				httpDoer: mockHTTPDoer(t, instaproxy.DefaultBaseURL+"/followers/456", "testdata/followers.json"),
			},
			wants{
				out: &instaproxy.Connections{
					Next:  strPtr(t, "wxyz123"),
					Users: stubUsers,
				},
			},
		},
		"GetFollowing (paginated)": {
			fields{
				callMethod: func(c *instaproxy.Client) (any, error) {
					return c.GetFollowing(context.TODO(), int64(123), strPtr(t, "wxyz"))
				},
				httpDoer: mockHTTPDoer(t, instaproxy.DefaultBaseURL+"/following/123?next_cursor=wxyz", "testdata/following.json"),
			},
			wants{
				out: &instaproxy.Connections{
					Next:  strPtr(t, "abcdef012345"),
					Users: stubUsers,
				},
			},
		},
		"GetFollowing": {
			fields{
				callMethod: func(c *instaproxy.Client) (any, error) {
					return c.GetFollowing(context.TODO(), int64(456), nil)
				},
				httpDoer: mockHTTPDoer(t, instaproxy.DefaultBaseURL+"/following/456", "testdata/following.json"),
			},
			wants{
				out: &instaproxy.Connections{
					Next:  strPtr(t, "abcdef012345"),
					Users: stubUsers,
				},
			},
		},
		"GetUser": {
			fields{
				callMethod: func(c *instaproxy.Client) (any, error) {
					return c.GetUser(context.TODO(), "johndoe")
				},
				httpDoer: mockHTTPDoer(t, instaproxy.DefaultBaseURL+"/account/johndoe", "testdata/user.json"),
			},
			wants{
				out: stubUser,
			},
		},
		"GetUserByID": {
			fields{
				callMethod: func(c *instaproxy.Client) (any, error) {
					return c.GetUserByID(context.TODO(), int64(12345))
				},
				httpDoer: mockHTTPDoer(t, instaproxy.DefaultBaseURL+"/account-id/12345", "testdata/user.json"),
			},
			wants{
				out: stubUser,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := instaproxy.NewClient(test.fields.httpDoer, nil)
			out, err := test.fields.callMethod(client)

			assert.NoError(t, err)
			assert.Equal(t, test.wants.out, out)
		})
	}
}
