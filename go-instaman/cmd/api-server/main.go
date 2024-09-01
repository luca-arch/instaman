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

// The main package for the api-server executable.
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/luca-arch/instaman/instaproxy"
	"github.com/luca-arch/instaman/webserver"
)

const (
	instaproxyTimeout = 90 // The instaproxy client's timeout. High value to account for latency due to retries and login attempts.
)

// Boot set up a JSON logger, a new Instaproxy client, and the api webserver.
func Boot(ctx context.Context, devMode bool) (*http.Server, *slog.Logger) {
	// Set up logger.
	var logger *slog.Logger

	lvl := new(slog.LevelVar)
	opts := &slog.HandlerOptions{
		AddSource:   false,
		Level:       lvl,
		ReplaceAttr: nil,
	}

	if devMode {
		opts.AddSource = true
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))

		lvl.Set(slog.LevelDebug)
	} else {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	// Set up dependencies.
	httpClient := &http.Client{Timeout: instaproxyTimeout * time.Second} //nolint:exhaustruct // Defaults are ok

	// Set up Instaproxy client.
	igClient := instaproxy.NewClient(httpClient, logger)
	if devMode {
		if err := igClient.BaseURL("http://127.0.0.1:15000"); err != nil {
			panic(err)
		}
	}

	// Init server with routes.
	server, err := webserver.Create(ctx, igClient, logger)
	if err != nil {
		logger.Error("could not bootstrap api-server", "error", err)
		panic(err)
	}

	return server, logger
}

func main() {
	devMode := flag.Bool("dev", false, "run in development mode (debug logger, and local instaproxy)")
	flag.Parse()

	server, logger := Boot(context.Background(), *devMode)

	logger.Info("api-server listening on " + server.Addr)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
