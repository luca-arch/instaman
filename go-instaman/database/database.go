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

// Package database provides types and interfaces for the relational storage.
package database

import (
	"context"
	"errors"
	"io"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDatabaseFailure = errors.New("postgresql error") // Wrapper for pgx/pgxpool errors.

type connectionPool interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

// Database wraps a PostgreSQL connection pool.
type Database struct {
	cnx    connectionPool
	logger *slog.Logger
}

// WithLogger sets the logger.
func (d *Database) WithLogger(logger *slog.Logger) *Database {
	d.logger = logger

	return d
}

// WithPool sets the connection pool. This is only ever useful for testing.
func (d *Database) WithPool(cnx connectionPool) *Database {
	d.cnx = cnx

	return d
}

// NewPool instantiates a new connection pool from the provided DSN string.
func NewPool(ctx context.Context, dsn string) *Database {
	cnx, err := pgxpool.New(ctx, dsn)
	if err != nil {
		// Lazy panic here because it happens only with malformed dsn strings.
		panic(err)
	}

	return &Database{
		cnx:    cnx,
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
	}
}

// Count executes the provided SQL expecting a COUNT.
func Count(ctx context.Context, db *Database, sql string, args ...any) (int, error) {
	db.logger.Debug("Query", "sql", sql, "args", args)

	res, err := db.cnx.Query(ctx, sql, args...)
	if err != nil {
		return -1, errors.Join(ErrDatabaseFailure, err)
	}

	defer res.Close()

	count, err := pgx.CollectExactlyOneRow(res, pgx.RowTo[int])
	if err != nil {
		return -1, errors.Join(ErrDatabaseFailure, err)
	}

	// Rows MUST be closed prior to reading the error.
	// CollectExactlyOneRow does that already.
	if err := res.Err(); err != nil {
		return -1, errors.Join(ErrDatabaseFailure, err)
	}

	return count, nil
}

// Select executes the provided SQL and returns the whole resultset.
func Select[T any](ctx context.Context, db *Database, sql string, args ...any) ([]T, error) {
	db.logger.Debug("Query", "sql", sql, "args", args)

	var out []T

	res, err := db.cnx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Join(ErrDatabaseFailure, err)
	}

	defer res.Close()

	out, err = pgx.CollectRows(res, pgx.RowToStructByNameLax[T])
	if err != nil {
		return nil, errors.Join(ErrDatabaseFailure, err)
	}

	// Rows MUST be closed prior to reading the error.
	// CollectRows does that already.
	if err := res.Err(); err != nil {
		return nil, errors.Join(ErrDatabaseFailure, err)
	}

	return out, nil
}

// Select executes the provided SQL and return the found row.
// It returns an error if none, or if more than one rows are found.
func SelectOne[T any](ctx context.Context, db *Database, sql string, args ...any) (*T, error) {
	db.logger.Debug("Query", "sql", sql, "args", args)

	res, err := db.cnx.Query(ctx, sql, args...)
	if err != nil {
		return nil, errors.Join(ErrDatabaseFailure, err)
	}

	defer res.Close()

	out, err := pgx.CollectExactlyOneRow(res, pgx.RowToStructByPos[T])
	if err != nil {
		return nil, errors.Join(ErrDatabaseFailure, err)
	}

	// Rows MUST be closed prior to reading the error.
	// CollectExactlyOneRow does that already.
	if err := res.Err(); err != nil {
		return nil, errors.Join(ErrDatabaseFailure, err)
	}

	return &out, nil
}
