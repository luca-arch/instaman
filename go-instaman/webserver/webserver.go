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

// Package webserver provides an http.Server that relays HTTP requests to the instaproxy service.
package webserver

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"
)

const (
	// Permissive http.Server timeout values.
	serverIdleTimeout  = 120
	serverReadTimeout  = 10
	serverWriteTimeout = 10
)

// Create sets up an HTTP server with all the app routes mounted.
func Create(ctx context.Context, igClient igclient, logger *slog.Logger) (*http.Server, error) {
	wrapped := WrapInstagramClient(igClient)
	relay := DefaultPicturesRelay(logger)

	mux := &http.ServeMux{}

	mux.Handle("GET /instagram/me", HandleWithRequest(logger, wrapped.GetAccount))
	mux.Handle("GET /instagram/account/{name}", HandleWithRequest(logger, wrapped.GetUser))
	mux.Handle("GET /instagram/account-id/{id}", HandleWithRequest(logger, wrapped.GetUserByID))
	mux.Handle("GET /instagram/followers/{id}", HandleWithRequest(logger, wrapped.GetFollowers))
	mux.Handle("GET /instagram/following/{id}", HandleWithRequest(logger, wrapped.GetFollowing))

	mux.Handle("GET /instagram/picture", relay)

	relay.Watch(ctx, FlushFrequency)

	return &http.Server{ //nolint:exhaustruct // Defaults are ok
		Addr:              ":10000",
		Handler:           mux,
		IdleTimeout:       serverIdleTimeout * time.Second,
		ReadHeaderTimeout: serverReadTimeout * time.Second,
		ReadTimeout:       serverReadTimeout * time.Second,
		WriteTimeout:      serverWriteTimeout * time.Second,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}, nil
}
