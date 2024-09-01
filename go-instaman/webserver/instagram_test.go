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
	"errors"
	"net/http"
	"testing"

	"github.com/luca-arch/instaman/instaproxy"
	"github.com/luca-arch/instaman/webserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockInstagramClient struct {
	mock.Mock
}

func (m *mockInstagramClient) GetAccount(ctx context.Context) (*instaproxy.Account, error) {
	args := m.Called(ctx)

	return args.Get(0).(*instaproxy.Account), args.Error(1)
}

func (m *mockInstagramClient) GetFollowers(ctx context.Context, userID int64, cursor *string) (*instaproxy.Connections, error) {
	args := m.Called(ctx, userID, cursor)

	return args.Get(0).(*instaproxy.Connections), args.Error(1)
}

func (m *mockInstagramClient) GetFollowing(ctx context.Context, userID int64, cursor *string) (*instaproxy.Connections, error) {
	args := m.Called(ctx, userID, cursor)

	return args.Get(0).(*instaproxy.Connections), args.Error(1)
}

func (m *mockInstagramClient) GetUser(ctx context.Context, username string) (*instaproxy.User, error) {
	args := m.Called(ctx, username)

	return args.Get(0).(*instaproxy.User), args.Error(1)
}

func (m *mockInstagramClient) GetUserByID(ctx context.Context, userID int64) (*instaproxy.User, error) {
	args := m.Called(ctx, userID)

	return args.Get(0).(*instaproxy.User), args.Error(1)
}

func httpRequest(t *testing.T, pathValues map[string]string) *http.Request {
	t.Helper()

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "https://example.com/any/", nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(pathValues) == 0 {
		return req
	}

	for name, value := range pathValues {
		req.SetPathValue(name, value)
	}

	return req
}

//nolint:maintidx // test all methods
func TestMethods(t *testing.T) {
	t.Parallel()

	stubAccount := &instaproxy.Account{
		Biography:  "Account biography",
		FullName:   "John Doe",
		Handler:    "john_doe",
		ID:         123,
		PictureURL: nil,
	}

	stubCursor1, stubCursor2 := "abcdef-0123456789", "abcdef-0123456789"

	stubErr := errors.New("stub error for mocked responses")

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
		callMethod func(*webserver.InstagramClient) (any, error)
		setupMock  func() *mockInstagramClient
	}

	type wants struct {
		err error
		out any
	}

	tests := map[string]struct {
		fields
		wants
	}{
		"method GetAccount - ok": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					return ic.GetAccount(httpRequest(t, nil))
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetAccount", mock.Anything).
						Return(stubAccount, nil)

					return client
				},
			},
			wants{
				err: nil,
				out: stubAccount,
			},
		},
		"method GetAccount - error": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					return ic.GetAccount(httpRequest(t, nil))
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetAccount", mock.Anything).
						Return(&instaproxy.Account{}, stubErr)

					return client
				},
			},
			wants{
				err: stubErr,
				out: nil,
			},
		},
		"method GetFollowers - ok": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, map[string]string{"id": "5678"})

					return ic.GetFollowers(r)
				},
				setupMock: func() *mockInstagramClient {
					out := &instaproxy.Connections{
						Next:  &stubCursor1,
						Users: stubUsers,
					}

					client := &mockInstagramClient{}
					client.On("GetFollowers", mock.Anything, int64(5678), (*string)(nil)).
						Return(out, nil)

					return client
				},
			},
			wants{
				err: nil,
				out: &instaproxy.Connections{
					Next:  &stubCursor1,
					Users: stubUsers,
				},
			},
		},
		"method GetFollowers - bad call": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, nil)

					return ic.GetFollowers(r)
				},
				setupMock: func() *mockInstagramClient {
					return &mockInstagramClient{}
				},
			},
			wants{
				err: webserver.ErrInvalidUserID,
				out: nil,
			},
		},
		"method GetFollowers - error": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, map[string]string{"id": "1234"})

					return ic.GetFollowers(r)
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetFollowers", mock.Anything, int64(1234), (*string)(nil)).
						Return(&instaproxy.Connections{}, stubErr)

					return client
				},
			},
			wants{
				err: stubErr,
				out: nil,
			},
		},
		"method GetFollowing - ok": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, map[string]string{"id": "1234"})

					return ic.GetFollowing(r)
				},
				setupMock: func() *mockInstagramClient {
					out := &instaproxy.Connections{
						Next:  &stubCursor2,
						Users: stubUsers,
					}

					client := &mockInstagramClient{}
					client.On("GetFollowing", mock.Anything, int64(1234), (*string)(nil)).
						Return(out, nil)

					return client
				},
			},
			wants{
				err: nil,
				out: &instaproxy.Connections{
					Next:  &stubCursor2,
					Users: stubUsers,
				},
			},
		},
		"method GetFollowing - bad call": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, nil)

					return ic.GetFollowing(r)
				},
				setupMock: func() *mockInstagramClient {
					return &mockInstagramClient{}
				},
			},
			wants{
				err: webserver.ErrInvalidUserID,
				out: nil,
			},
		},
		"method GetFollowing - error": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, map[string]string{"id": "1234"})

					return ic.GetFollowing(r)
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetFollowing", mock.Anything, int64(1234), (*string)(nil)).
						Return(&instaproxy.Connections{}, stubErr)

					return client
				},
			},
			wants{
				err: stubErr,
				out: nil,
			},
		},
		"method GetUser - ok": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, map[string]string{"name": "johndoe"})

					return ic.GetUser(r)
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetUser", mock.Anything, "johndoe").
						Return(&stubUsers[0], nil)

					return client
				},
			},
			wants{
				err: nil,
				out: &stubUsers[0],
			},
		},
		"method GetUser - bad call": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, nil)

					return ic.GetUser(r)
				},
				setupMock: func() *mockInstagramClient {
					return &mockInstagramClient{}
				},
			},
			wants{
				err: webserver.ErrInvalidUserName,
				out: nil,
			},
		},
		"method GetUser - error": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, map[string]string{"name": "johndoe"})

					return ic.GetUser(r)
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetUser", mock.Anything, "johndoe").
						Return(&instaproxy.User{}, stubErr)

					return client
				},
			},
			wants{
				err: stubErr,
				out: nil,
			},
		},
		"method GetUserByID - ok": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, map[string]string{"id": "1234"})

					return ic.GetUserByID(r)
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetUserByID", mock.Anything, int64(1234)).
						Return(&stubUsers[0], nil)

					return client
				},
			},
			wants{
				err: nil,
				out: &stubUsers[0],
			},
		},
		"method GetUserByID - bad call": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, nil)

					return ic.GetUserByID(r)
				},
				setupMock: func() *mockInstagramClient {
					return &mockInstagramClient{}
				},
			},
			wants{
				err: webserver.ErrInvalidUserID,
				out: nil,
			},
		},
		"method GetUserByID - error": {
			fields{
				callMethod: func(ic *webserver.InstagramClient) (any, error) {
					r := httpRequest(t, map[string]string{"id": "456"})

					return ic.GetUserByID(r)
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetUserByID", mock.Anything, int64(456)).
						Return(&instaproxy.User{}, stubErr)

					return client
				},
			},
			wants{
				err: stubErr,
				out: nil,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			client := test.setupMock()
			wsClient := webserver.WrapInstagramClient(client)

			res, err := test.fields.callMethod(wsClient)

			if test.wants.err == nil {
				assert.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, test.wants.out, res)

				return
			}

			assert.ErrorIs(t, err, test.wants.err)
		})
	}
}
