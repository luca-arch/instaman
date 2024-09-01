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

// Package instaproxy provides the HTTP connector to the instaproxy Python service.
package instaproxy

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	DefaultBaseURL   = "http://instaproxy:15000"
	DefaultUserAgent = "go-instaman"
)

var (
	ErrHTTPFailure   = errors.New("request failed")
	ErrInvalidArgs   = errors.New("illegal function invocation")
	ErrInvalidJSON   = errors.New("malformed response")
	ErrInvalidStatus = errors.New("unexpected status code")
	ErrInvalidURL    = errors.New("invalid URL")
	ErrNoProtocol    = errors.New("missing HTTP/HTTPS protocol")
	ErrNotFound      = errors.New("resource not found")
	ErrTransport     = errors.New("transport error")
)

// httpDoer defines an interface to make HTTP requests.
type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// Client is an instaproxy API client.
type Client struct {
	base   string
	client httpDoer
	logger *slog.Logger
}

// NewClient instantiates a new instaproxy API client.
func NewClient(client httpDoer, logger *slog.Logger) *Client {
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}

	return &Client{
		base:   DefaultBaseURL,
		client: client,
		logger: logger,
	}
}

// BaseURL sets the client's base URL.
func (c *Client) BaseURL(base string) error {
	u, err := url.Parse(base)
	if err != nil {
		return ErrInvalidURL
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrNoProtocol
	}

	c.base, _ = strings.CutSuffix(u.String(), "/")

	return nil
}

// GetAccount sends a GET request to instaproxy's `/me` endpoint and returns the primary account's information.
func (c *Client) GetAccount(ctx context.Context) (*Account, error) {
	return get[Account](ctx, c, "/me")
}

// GetFollowers sends a GET request to instaproxy's `/followers/{id}` endpoint and returns that user's connections.
func (c *Client) GetFollowers(ctx context.Context, userID int64, cursor *string) (*Connections, error) {
	endpoint := "/followers/" + strconv.FormatInt(userID, 10)

	if cursor != nil {
		endpoint = endpoint + "?next_cursor=" + url.QueryEscape(*cursor)
	}

	return get[Connections](ctx, c, endpoint)
}

// GetFollowing sends a GET request to instaproxy's `/following/{id}` endpoint and returns that user's connections.
func (c *Client) GetFollowing(ctx context.Context, userID int64, cursor *string) (*Connections, error) {
	endpoint := "/following/" + strconv.FormatInt(userID, 10)

	if cursor != nil {
		endpoint = endpoint + "?next_cursor=" + url.QueryEscape(*cursor)
	}

	return get[Connections](ctx, c, endpoint)
}

// GetUser sends a GET request to instaproxy's `/account/{username}` endpoint and returns that user's information.
func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	return get[User](ctx, c, "/account/"+username)
}

// GetUserByID sends a GET request to instaproxy's `/account-id/{id}` endpoint and returns that user's information.
func (c *Client) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	return get[User](ctx, c, "/account-id/"+strconv.FormatInt(userID, 10))
}

// Get sends a GET request to the instaproxy service.
func get[T Account | Connections | User](ctx context.Context, c *Client, endpoint string) (*T, error) {
	var out T

	c.logger.Info("instaproxy request", "http.request.method", http.MethodGet, "http.route", endpoint)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+endpoint, nil)
	if err != nil {
		return nil, errors.Join(ErrHTTPFailure, err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", DefaultUserAgent)

	resp, err := c.client.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}

	switch {
	case err != nil:
		return nil, errors.Join(ErrHTTPFailure, err)
	case resp.StatusCode == http.StatusNotFound:
		return nil, ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return nil, ErrInvalidStatus
	default:
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return nil, errors.Join(ErrInvalidJSON, err)
		}
	}

	return &out, nil
}
