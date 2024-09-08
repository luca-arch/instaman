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

package database

import (
	"context"

	"github.com/luca-arch/instaman/database/models"
)

// querier defines an interface to abstract all database queries.
// It is only ever useful for testing purposes.
type querier interface {
	Count(context.Context, *Database, string, ...any) (int32, error)
	Execute(context.Context, *Database, string, ...any) error
	SelectJob(context.Context, *Database, string, ...any) (*models.Job, error)
	SelectJobs(context.Context, *Database, string, ...any) ([]models.Job, error)
	SelectUsers(context.Context, *Database, string, ...any) ([]models.User, error)
}

// Querier is the default querier that simply calls Count, Select, SelectOne and Execute.
type Querier struct{}

// Count calls the Count function to return the number of counted records.
func (q *Querier) Count(ctx context.Context, db *Database, sql string, args ...any) (int32, error) {
	return Count(ctx, db, sql, args...)
}

// Execute calls the Execute function to return any error that might occur.
func (q *Querier) Execute(ctx context.Context, db *Database, sql string, args ...any) error {
	return Execute(ctx, db, sql, args...)
}

// SelectJob calls the SelectOne function to return a `Job` objects.
func (q *Querier) SelectJob(ctx context.Context, db *Database, sql string, args ...any) (*models.Job, error) {
	return SelectOne[models.Job](ctx, db, sql, args...)
}

// SelectJobs calls the Select function to return a list of `Job` objects.
func (q *Querier) SelectJobs(ctx context.Context, db *Database, sql string, args ...any) ([]models.Job, error) {
	return Select[models.Job](ctx, db, sql, args...)
}

// SelectUsers calls the Select function to return a list of `User` objects.
func (q *Querier) SelectUsers(ctx context.Context, db *Database, sql string, args ...any) ([]models.User, error) {
	return Select[models.User](ctx, db, sql, args...)
}
