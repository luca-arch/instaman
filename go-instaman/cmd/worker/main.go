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

// The main package for the worker executable.
package main

import (
	"context"
	"flag"
	"log/slog"
	"os"

	"github.com/luca-arch/instaman/internal"
	"github.com/luca-arch/instaman/service"
)

// Boot sets up the worker and its dependencies.
func Boot(ctx context.Context, devMode bool) (*service.Worker, *slog.Logger) {
	isDocker := os.Getenv("ISDOCKER") == "1"
	logger := internal.Logger(devMode)

	// Set up dependencies.
	db := internal.Database(ctx, logger, isDocker)
	instaproxy := internal.Instaproxy(logger, isDocker)

	// Init worker.
	worker := service.NewWorkerService(db, logger, instaproxy)

	return worker, logger
}

func main() {
	devMode := flag.Bool("dev", false, "enable debug logger")
	flag.Parse()

	ctx := context.Background()

	worker, logger := Boot(ctx, *devMode)

	logger.Info("starting worker...")

	worker.StartCopying(ctx)
}
