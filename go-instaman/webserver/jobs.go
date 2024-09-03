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

	"github.com/luca-arch/instaman/database"
	"github.com/luca-arch/instaman/database/models"
)

// jobservice describes a service that can access and manipulate jobs.
type jobservice interface {
	FindCopyJob(context.Context, database.FindCopyJobParams) (*models.CopyJob, error)
	FindJob(context.Context, database.FindJobParams) (*models.Job, error)
}
