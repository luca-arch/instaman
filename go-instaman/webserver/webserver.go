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
func Create(ctx context.Context, jobService jobservice, igservice igservice, logger *slog.Logger) (*http.Server, error) {
	// wrapped := WrapInstagramClient(igClient)
	relay := DefaultPicturesRelay(logger)

	mux := &http.ServeMux{}

	mux.Handle("GET /instaman/instagram/me", Handle(logger, igservice.GetAccount))
	mux.Handle("GET /instaman/instagram/account/{name}", HandleWithInput(logger, igservice.GetUser))
	mux.Handle("GET /instaman/instagram/account-id/{id}", HandleWithInput(logger, igservice.GetUserByID))
	mux.Handle("GET /instaman/instagram/followers/{id}", HandleWithInput(logger, igservice.GetFollowers))
	mux.Handle("GET /instaman/instagram/following/{id}", HandleWithInput(logger, igservice.GetFollowing))

	mux.Handle("GET /instaman/instagram/picture", relay)

	mux.Handle("GET /instaman/jobs/all", HandleWithInput(logger, jobService.FindJobs))
	mux.Handle("GET /instaman/jobs/copy", HandleWithInput(logger, jobService.FindCopyJob))
	mux.Handle("GET /instaman/jobs", HandleWithInput(logger, jobService.FindJob))
	mux.Handle("POST /instaman/jobs/copy", HandleWithInput(logger, jobService.NewCopyJob))

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
