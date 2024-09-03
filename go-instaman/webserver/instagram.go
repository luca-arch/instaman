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

package webserver

import (
	"context"
	"errors"

	"github.com/luca-arch/instaman/instaproxy"
	"github.com/luca-arch/instaman/service"
)

var (
	// The user ID in request's path is not a valid integer.
	ErrInvalidUserID = errors.New("invalid user ID")

	// The username in request's path is empty.
	ErrInvalidUserName = errors.New("invalid username")
)

// igservice describes a service that can interact with instaproxy.
type igservice interface {
	GetAccount(context.Context) (*instaproxy.Account, error)
	GetFollowers(context.Context, service.GetConnectionInput) (*instaproxy.Connections, error)
	GetFollowing(context.Context, service.GetConnectionInput) (*instaproxy.Connections, error)
	GetUser(context.Context, service.GetUserInput) (*instaproxy.User, error)
	GetUserByID(context.Context, service.GetUserByIDInput) (*instaproxy.User, error)
}
