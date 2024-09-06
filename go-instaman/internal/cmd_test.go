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
package internal_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/luca-arch/instaman/internal"
	"github.com/stretchr/testify/assert"
)

// This test does almost nothing but increase code coverage.
func TestDatabase(t *testing.T) {
	t.Parallel()

	out := internal.Database(context.TODO(), nopLogger(t), true)
	assert.NotNil(t, out)

	out = internal.Database(context.TODO(), nopLogger(t), false)
	assert.NotNil(t, out)
}

// This test does almost nothing but increase code coverage.
func TestLogger(t *testing.T) {
	t.Parallel()

	out := internal.Logger(true)
	assert.NotNil(t, out)
	assert.True(t, out.Handler().Enabled(context.TODO(), slog.LevelDebug))

	out = internal.Logger(false)
	assert.NotNil(t, out)
	assert.False(t, out.Handler().Enabled(context.TODO(), slog.LevelDebug))
}

// This test does almost nothing but increase code coverage.
func TestInstaproxy(t *testing.T) {
	t.Parallel()

	out := internal.Instaproxy(nopLogger(t), true)
	assert.NotNil(t, out)

	out = internal.Instaproxy(nopLogger(t), false)
	assert.NotNil(t, out)
}

func nopLogger(t *testing.T) *slog.Logger {
	t.Helper()

	discard := slog.NewJSONHandler(io.Discard, nil)

	return slog.New(discard)
}
