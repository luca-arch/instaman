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
package main_test

import (
	"context"
	"log/slog"
	"testing"

	apiserver "github.com/luca-arch/instaman/cmd/api-server"
	"github.com/stretchr/testify/assert"
)

// This test does almost nothing but increase code coverage.
func TestBoot(t *testing.T) {
	t.Parallel()

	ctx := context.TODO()

	_, logger := apiserver.Boot(ctx, false)
	assert.False(t, logger.Handler().Enabled(ctx, slog.LevelDebug))

	_, logger = apiserver.Boot(ctx, true)
	assert.True(t, logger.Handler().Enabled(ctx, slog.LevelDebug))
}
