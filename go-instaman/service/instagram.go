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

package service

import (
	"context"
	"errors"

	"github.com/luca-arch/instaman/instaproxy"
)

var (
	// The user ID in request's path is not a valid integer.
	ErrInvalidUserID = errors.New("invalid user ID")

	// The username in request's path is empty.
	ErrInvalidUserName = errors.New("invalid username")
)

// Instagram wraps an instaproxy.Client to call its methods passing arguments that are read from an HTTP request.
type Instagram struct {
	client igclient
}

// igclient describes an instaproxy.Client.
type igclient interface {
	GetAccount(context.Context) (*instaproxy.Account, error)
	GetFollowers(context.Context, int64, *string) (*instaproxy.Connections, error)
	GetFollowing(context.Context, int64, *string) (*instaproxy.Connections, error)
	GetUser(context.Context, string) (*instaproxy.User, error)
	GetUserByID(context.Context, int64) (*instaproxy.User, error)
}

// GetConnectionInput defines input parameters for GetFollowers and GetFollowing methods.
type GetConnectionInput struct {
	Cursor *string `in:"next_cursor,omitempty"`
	UserID int64   `in:"id,path,required"`
}

// GetUserByIDInput defines input parameters for GetFollowers and GetFollowing methods.
type GetUserByIDInput struct {
	UserID int64 `in:"id,path,required"`
}

// GetUserInput defines input parameters for GetFollowers and GetFollowing methods.
type GetUserInput struct {
	Handler string `in:"name,path,required"`
}

// NewInstagramService sets up and returns a new Instaproxy Service.
func NewInstagramService(client igclient) *Instagram {
	return &Instagram{
		client: client,
	}
}

// GetAccount wraps the client's GetAccount method.
func (i *Instagram) GetAccount(ctx context.Context) (*instaproxy.Account, error) {
	return i.client.GetAccount(ctx) //nolint:wrapcheck // Wraps invocation
}

// GetFollowers wraps the client's GetFollowers method.
func (i *Instagram) GetFollowers(ctx context.Context, in GetConnectionInput) (*instaproxy.Connections, error) {
	return i.client.GetFollowers(ctx, in.UserID, in.Cursor) //nolint:wrapcheck // Wraps invocation
}

// GetFollowing wraps the client's GetFollowing method.
func (i *Instagram) GetFollowing(ctx context.Context, in GetConnectionInput) (*instaproxy.Connections, error) {
	return i.client.GetFollowing(ctx, in.UserID, in.Cursor) //nolint:wrapcheck // Wraps invocation
}

// GetUser wraps the client's GetUser method.
func (i *Instagram) GetUser(ctx context.Context, in GetUserInput) (*instaproxy.User, error) {
	return i.client.GetUser(ctx, in.Handler) //nolint:wrapcheck // Wraps invocation
}

// GetUserByID wraps the client's GetUserByID method.
func (i *Instagram) GetUserByID(ctx context.Context, in GetUserByIDInput) (*instaproxy.User, error) {
	return i.client.GetUserByID(ctx, in.UserID) //nolint:wrapcheck // Wraps invocation
}
