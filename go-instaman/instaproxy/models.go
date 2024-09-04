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

package instaproxy

import (
	"encoding/json"
	"errors"
	"net/url"
)

var ErrInvalidPictureURL = errors.New("invalid pictureURL")

// Account is a struct that mirrors instaproxy's `AccountDict` objetcs.
type Account struct {
	Biography string `description:"Account bio" json:"biography"`
	FullName  string `description:"Full name" json:"fullName"`
	Handler   string `description:"Handler without @" json:"handler"`
	ID        int64  `description:"Account ID" json:"id"`
	//nolint:tagliatelle // Proxy returns pictureURL
	PictureURL *URLField `description:"Avatar URL" json:"pictureURL,omitempty"`
}

// Connections is a struct that mirrors instaproxy's `/followers/<id>` and `/following/<id>` response.
type Connections struct {
	Next  *string `description:"Next cursor for pagination" json:"next,omitempty"`
	Users []User  `description:"List of users" json:"users"`
}

// User is a struct that mirrors instaproxy's `InstagramUserDict` objects.
type User struct {
	FullName string `description:"Full name" json:"fullName"`
	Handler  string `description:"Handler without @" json:"handler"`
	ID       int64  `description:"Account ID" json:"id"`
	//nolint:tagliatelle // Proxy returns pictureURL
	PictureURL *URLField `description:"Avatar URL" json:"pictureURL,omitempty"`
}

// URLField is a type that implements json.Marshaler and json.Unmarshaler for URLs.
type URLField struct {
	url.URL
}

// MarshalJSON satisfies json.Marshaler interface.
func (u *URLField) MarshalJSON() ([]byte, error) {
	if !u.IsAbs() {
		return nil, ErrInvalidPictureURL
	}

	val, err := json.Marshal(u.String())
	if err != nil {
		return nil, errors.Join(err, ErrInvalidPictureURL)
	}

	return val, nil
}

// UnmarshalJSON satisfies json.Unmarshaler interface.
func (u *URLField) UnmarshalJSON(data []byte) error {
	if len(data) == 0 {
		return nil
	}

	s := ""
	err := json.Unmarshal(data, &s)

	switch {
	case err != nil:
		return errors.Join(err, ErrInvalidPictureURL)
	case s == "":
		return nil
	}

	val, err := url.Parse(s)

	switch {
	case err != nil:
		return errors.Join(err, ErrInvalidPictureURL)
	case !val.IsAbs():
		return ErrInvalidPictureURL
	}

	*u = URLField{*val}

	return nil
}
