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

package webserver

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/luca-arch/instaman/instaproxy"
)

var (
	// The user ID in request's path is not a valid integer.
	ErrInvalidUserID = errors.New("invalid user ID")

	// The username in request's path is empty.
	ErrInvalidUserName = errors.New("invalid username")
)

// InstagramClient wraps an instaproxy.Client to call its methods passing arguments that are read from an HTTP request.
type InstagramClient struct {
	client igclient
}

// igclient describes the instaproxy.Client to be wrapped.
type igclient interface {
	GetAccount(context.Context) (*instaproxy.Account, error)
	GetFollowers(context.Context, int64, *string) (*instaproxy.Connections, error)
	GetFollowing(context.Context, int64, *string) (*instaproxy.Connections, error)
	GetUser(context.Context, string) (*instaproxy.User, error)
	GetUserByID(context.Context, int64) (*instaproxy.User, error)
}

// WrapInstagramClient wraps an Instagram client methods.
func WrapInstagramClient(client igclient) *InstagramClient {
	return &InstagramClient{
		client: client,
	}
}

// GetAccount wraps the client's GetAccount method.
func (i *InstagramClient) GetAccount(r *http.Request) (*instaproxy.Account, error) {
	return i.client.GetAccount(r.Context()) //nolint:wrapcheck // Wraps invocation
}

// GetFollowers wraps the client's GetFollowers method.
func (i *InstagramClient) GetFollowers(r *http.Request) (*instaproxy.Connections, error) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	return i.client.GetFollowers(r.Context(), userID, cursor(r)) //nolint:wrapcheck // Wraps invocation
}

// GetFollowing wraps the client's GetFollowing method.
func (i *InstagramClient) GetFollowing(r *http.Request) (*instaproxy.Connections, error) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	return i.client.GetFollowing(r.Context(), userID, cursor(r)) //nolint:wrapcheck // Wraps invocation
}

// GetUser wraps the client's GetUser method.
func (i *InstagramClient) GetUser(r *http.Request) (*instaproxy.User, error) {
	username := r.PathValue("name")
	if username == "" {
		return nil, ErrInvalidUserName
	}

	return i.client.GetUser(r.Context(), username) //nolint:wrapcheck // Wraps invocation
}

// GetUserByID wraps the client's GetUserByID method.
func (i *InstagramClient) GetUserByID(r *http.Request) (*instaproxy.User, error) {
	userID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	return i.client.GetUserByID(r.Context(), userID) //nolint:wrapcheck // Wraps invocation
}

// cursor reads the next_cursor HTTP query argument.
func cursor(r *http.Request) *string {
	c := r.URL.Query().Get("next_cursor")
	if strings.TrimSpace(c) == "" {
		return nil
	}

	return &c
}
