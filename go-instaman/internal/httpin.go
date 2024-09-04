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

// Package internal.
package internal

import (
	"errors"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// InputFromRequest hydrates a struct reading from the request args and path.
// Behaviour is defined via struct tags, eg:
//   - `in:"pk,path,required"` will search for the pathvalue named pk, and return an error if not found.
//   - `in:"job_id,omitempty"` will search for the query arg named job_id, allowing it to be empty.
func InputFromRequest[T any](r *http.Request) (T, error) { //nolint:ireturn
	var (
		err error
		in  T
	)

	// Get the reflect.Value of the struct
	inValue := reflect.ValueOf(&in).Elem()
	inType := inValue.Type()

	// Iterate over all the fields of the struct
	for i := 0; i < inType.NumField(); i++ {
		field := inType.Field(i)
		tag := field.Tag.Get("in")

		// Skip the field if there is no "in" tag
		if tag == "" || tag == "-" {
			continue
		}

		var queryValue string

		// Parse tag options
		tagParts := strings.Split(tag, ",")
		tagName := tagParts[0]
		isRequired := false
		omitEmpty := false
		inPath := false

		for _, option := range tagParts[1:] {
			switch option {
			case "path":
				inPath = true
			case "required":
				isRequired = true
			case "omitempty":
				omitEmpty = true
			}
		}

		if inPath {
			// Get the value from the path.
			queryValue = r.PathValue(tagName)
		} else {
			// Get the value from the URL query parameters.
			queryValue = r.URL.Query().Get(tagName)
		}

		// Handle required fields.
		if queryValue == "" {
			if isRequired {
				return in, errors.New("missing required field: " + tagName) //nolint:err113
			}

			if omitEmpty {
				continue
			}
		}

		// Set the field value.
		fieldValue := inValue.Field(i)
		switch fieldValue.Kind() { //nolint:exhaustive // The default should cover enough.
		case reflect.Ptr:
			err = hydratePointer(&fieldValue, &field, tagName, queryValue)
		default:
			err = hydrateValue(&fieldValue, tagName, queryValue)
		}

		if err != nil {
			return in, err
		}
	}

	return in, nil
}

// hydratePointer sets the pointer's value based on its type and the queryValue.
func hydratePointer(fieldValue *reflect.Value, field *reflect.StructField, tagName, queryValue string) error {
	fieldType := field.Type
	elemType := fieldType.Elem()

	if queryValue == "" {
		fieldValue.Set(reflect.Zero(fieldType))

		return nil
	}

	elemValue := reflect.New(elemType).Elem()

	switch elemType.Kind() { //nolint:exhaustive
	case reflect.String:
		elemValue.SetString(queryValue)
	case reflect.Int, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(queryValue, 10, elemType.Bits())
		if err != nil {
			return errors.New("invalid integer value for field: " + tagName) //nolint:err113
		}

		elemValue.SetInt(intVal)
	case reflect.Struct:
		if elemType == reflect.TypeOf(time.Time{}) {
			timeVal, err := time.Parse(time.RFC3339, queryValue)
			if err != nil {
				return errors.New("invalid time format for field: " + tagName) //nolint:err113
			}

			elemValue.Set(reflect.ValueOf(timeVal))
		} else if elemType == reflect.TypeOf(url.URL{}) { //nolint:exhaustruct // Needed only for type-checking
			urlVal, err := url.Parse(queryValue)
			if err != nil {
				return errors.New("invalid URL format for field: " + tagName) //nolint:err113
			}

			elemValue.Set(reflect.ValueOf(*urlVal))
		}
	}

	fieldValue.Set(elemValue.Addr())

	return nil
}

// hydrateValue sets the value based on its type and the queryValue.
func hydrateValue(fieldValue *reflect.Value, tagName, queryValue string) error {
	switch fieldValue.Kind() { //nolint:exhaustive
	case reflect.String:
		fieldValue.SetString(queryValue)
	case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
		if queryValue == "" {
			fieldValue.SetInt(0)
		} else {
			intVal, err := strconv.ParseInt(queryValue, 10, fieldValue.Type().Bits())
			if err != nil {
				return errors.New("invalid number for field: " + tagName) //nolint:err113
			}

			fieldValue.SetInt(intVal)
		}
	case reflect.Struct:
		switch fieldValue.Type() {
		case reflect.TypeOf(time.Time{}):
			if queryValue == "" {
				fieldValue.Set(reflect.Zero(fieldValue.Type()))
			} else {
				timeVal, err := time.Parse(time.RFC3339, queryValue)
				if err != nil {
					return errors.New("invalid time format for field: " + tagName) //nolint:err113
				}

				fieldValue.Set(reflect.ValueOf(timeVal))
			}
		case reflect.TypeOf(url.URL{}): //nolint:exhaustruct // Needed only for type-checking
			if queryValue == "" {
				fieldValue.Set(reflect.Zero(fieldValue.Type()))
			} else {
				urlVal, err := url.Parse(queryValue)
				if err != nil {
					return errors.New("invalid URL format for field: " + tagName) //nolint:err113
				}

				fieldValue.Set(reflect.ValueOf(*urlVal))
			}
		}
	default:
		return errors.New("cannot parse " + tagName + ": " + fieldValue.Kind().String()) //nolint:err113
	}

	return nil
}
