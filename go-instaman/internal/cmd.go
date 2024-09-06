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

// Package internal provides utilities that are only intended to be used by the go-instaman app itself.
package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/luca-arch/instaman/database"
	"github.com/luca-arch/instaman/instaproxy"
)

const (
	instaproxyTimeout = 90 // The instaproxy client's timeout. High value to account for latency due to retries and login attempts.
	psqlMaxPoolSize   = 5  // Postgres pool size (max)
	psqlMinPoolSize   = 2  // Postgres pool size (min)
)

// Database builds a DSN to create and return a new database connection.
func Database(ctx context.Context, logger *slog.Logger, isDocker bool) *database.Database {
	var dsn string

	if isDocker {
		// Build DSN reading values from the environment.
		user, pass, db, host := os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB"), "postgres"

		dsn = fmt.Sprintf("postgres://%s:%s@%s/%s?pool_max_conns=%d&pool_min_conns=%d",
			user,
			pass,
			net.JoinHostPort(host, "5432"),
			db,
			psqlMaxPoolSize,
			psqlMinPoolSize,
		)
	} else {
		// Hardcoded DSN string, with values from the original docker-compose.yml file
		dsn = "postgres://postgresuser:postgressecret@127.0.0.1:5432/database001?pool_max_conns=5&pool_min_conns=1"
	}

	return database.
		NewPool(ctx, dsn).
		WithLogger(logger)
}

// Logger sets up a new slog.Logger and returns it.
func Logger(debug bool) *slog.Logger {
	lvl := new(slog.LevelVar)
	opts := &slog.HandlerOptions{
		AddSource:   debug,
		Level:       lvl,
		ReplaceAttr: nil,
	}

	if !debug {
		return slog.New(slog.NewJSONHandler(os.Stdout, opts))
	}

	lvl.Set(slog.LevelDebug)

	return slog.New(slog.NewTextHandler(os.Stdout, opts))
}

// Instaproxy sets up a new instaproxy client and returns it.
func Instaproxy(logger *slog.Logger, isDocker bool) *instaproxy.Client {
	httpClient := &http.Client{Timeout: instaproxyTimeout * time.Second} //nolint:exhaustruct // Defaults are ok

	// Set up Instaproxy client and service.
	igClient := instaproxy.NewClient(httpClient, logger)
	if !isDocker {
		if err := igClient.BaseURL("http://127.0.0.1:15000"); err != nil {
			panic(err)
		}
	}

	return igClient
}
