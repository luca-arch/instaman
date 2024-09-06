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

	"github.com/luca-arch/instaman/internal"
	"github.com/luca-arch/instaman/service"
	"github.com/luca-arch/instaman/webserver"
)

// Boot sets up the api webserver and its dependencies.
func Boot(ctx context.Context, devMode bool) (*http.Server, *slog.Logger) {
	isDocker := os.Getenv("ISDOCKER") == "1"
	logger := internal.Logger(devMode)

	// Set up dependencies.
	db := internal.Database(ctx, logger, isDocker)
	igService := service.NewInstagramService(internal.Instaproxy(logger, isDocker))
	jobService := service.NewJobsService(db)

	// Init server with routes.
	server, err := webserver.Create(ctx, jobService, igService, logger)
	if err != nil {
		logger.Error("could not bootstrap api-server", "error", err)
		panic(err)
	}

	return server, logger
}

func main() {
	devMode := flag.Bool("dev", false, "enable debug logger")
	flag.Parse()

	server, logger := Boot(context.Background(), *devMode)

	logger.Info("api-server listening on " + server.Addr)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}
