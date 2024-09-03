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

package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/luca-arch/instaman/instaproxy"
	"github.com/luca-arch/instaman/service"
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

//nolint:maintidx // test all methods
func TestMethods(t *testing.T) {
	t.Parallel()

	testCtx := context.TODO()

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
		callMethod func(*service.Instagram) (any, error)
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
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetAccount(testCtx)
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetAccount", testCtx).
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
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetAccount(testCtx)
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetAccount", testCtx).
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
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetFollowers(testCtx, service.GetConnectionInput{
						UserID: 5678,
					})
				},
				setupMock: func() *mockInstagramClient {
					out := &instaproxy.Connections{
						Next:  &stubCursor1,
						Users: stubUsers,
					}

					client := &mockInstagramClient{}
					client.On("GetFollowers", testCtx, int64(5678), (*string)(nil)).
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
		"method GetFollowers - error": {
			fields{
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetFollowers(testCtx, service.GetConnectionInput{
						UserID: 1234,
					})
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetFollowers", testCtx, int64(1234), (*string)(nil)).
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
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetFollowing(testCtx, service.GetConnectionInput{
						UserID: 1234,
					})
				},
				setupMock: func() *mockInstagramClient {
					out := &instaproxy.Connections{
						Next:  &stubCursor2,
						Users: stubUsers,
					}

					client := &mockInstagramClient{}
					client.On("GetFollowing", testCtx, int64(1234), (*string)(nil)).
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
		"method GetFollowing - error": {
			fields{
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetFollowing(testCtx, service.GetConnectionInput{
						UserID: 1234,
					})
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetFollowing", testCtx, int64(1234), (*string)(nil)).
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
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetUser(testCtx, service.GetUserInput{
						Handler: "johndoe",
					})
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetUser", testCtx, "johndoe").
						Return(&stubUsers[0], nil)

					return client
				},
			},
			wants{
				err: nil,
				out: &stubUsers[0],
			},
		},
		"method GetUser - error": {
			fields{
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetUser(testCtx, service.GetUserInput{
						Handler: "johndoe",
					})
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetUser", testCtx, "johndoe").
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
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetUserByID(testCtx, service.GetUserByIDInput{
						UserID: 1234,
					})
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetUserByID", testCtx, int64(1234)).
						Return(&stubUsers[0], nil)

					return client
				},
			},
			wants{
				err: nil,
				out: &stubUsers[0],
			},
		},
		"method GetUserByID - error": {
			fields{
				callMethod: func(ic *service.Instagram) (any, error) {
					return ic.GetUserByID(testCtx, service.GetUserByIDInput{})
				},
				setupMock: func() *mockInstagramClient {
					client := &mockInstagramClient{}
					client.On("GetUserByID", testCtx, int64(0)).
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
			svc := service.NewInstagramService(client)

			res, err := test.fields.callMethod(svc)

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
