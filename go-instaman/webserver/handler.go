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
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/luca-arch/instaman/instaproxy"
	"github.com/luca-arch/instaman/internal"
)

type errResponse struct {
	Error string `json:"error"`
}

// TargetFunc is an HTTP handler that takes a generic input and returns a generic output.
// https://www.willem.dev/articles/generic-http-handlers/
type TargetFunc[Out any] func(context.Context) (Out, error)

// Handle takes a TargetFunc and uses it to create an HTTP handler.
// https://www.willem.dev/articles/generic-http-handlers/
func Handle[Out any](logger *slog.Logger, f TargetFunc[Out]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("HTTP request", "http.method", r.Method, "http.url", r.URL)

		// Call out to target function.
		out, err := f(r.Context())

		// Serve response.
		writeResponse(w, logger, out, err)
	})
}

// TargetFunc is an HTTP handler that takes a generic input and returns a generic output.
type TargetFuncWithInput[In any, Out any] func(context.Context, In) (Out, error)

// HandleWithInput takes a TargetFuncWithInput and uses it to create an HTTP handler that reads the request's body.
func HandleWithInput[In any, Out any](logger *slog.Logger, f TargetFuncWithInput[In, Out]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			in  In
			err error
		)

		logger.Info("HTTP request", "http.method", r.Method, "http.url", r.URL)

		switch r.Method {
		case http.MethodGet, http.MethodHead:
			// Read request's query/path.
			in, err = internal.InputFromRequest[In](r)
		default:
			// Read request's body.
			err = json.NewDecoder(r.Body).Decode(&in)
		}

		if err != nil {
			writeErrResponse(w, err, http.StatusBadRequest)

			return
		}

		// Call out to target function.
		out, err := f(r.Context(), in)

		// Serve response.
		writeResponse(w, logger, out, err)
	})
}

// TargetFuncWithRequest is an HTTP handler that takes a generic input + an HTTP request, and returns a generic output.
type TargetFuncWithRequest[Out any] func(*http.Request) (Out, error)

// HandleWithRequest takes a TargetFuncWithRequest and uses it to create an HTTP handler.
func HandleWithRequest[Out any](logger *slog.Logger, f TargetFuncWithRequest[Out]) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("HTTP request", "http.method", r.Method, "http.url", r.URL)

		// Call out to target function.
		out, err := f(r)

		// Serve response.
		writeResponse(w, logger, out, err)
	})
}

// writeResponse is an helper that writes JSON-encoded data into the ResponseWriter.
func writeResponse[T any](w http.ResponseWriter, logger *slog.Logger, out T, err error) {
	w.Header().Set("Content-Type", "application/json")

	var wErr error

	switch {
	case err == nil:
		w.WriteHeader(http.StatusOK)
		wErr = json.NewEncoder(w).Encode(out)
	case errors.Is(err, instaproxy.ErrInvalidStatus):
		w.WriteHeader(http.StatusBadGateway)
	case errors.Is(err, instaproxy.ErrNotFound):
		w.WriteHeader(http.StatusNotFound)
		wErr = json.NewEncoder(w).Encode(errResponse{Error: err.Error()})
	default:
		w.WriteHeader(http.StatusInternalServerError)
		wErr = json.NewEncoder(w).Encode(errResponse{Error: err.Error()})
	}

	if wErr != nil {
		logger.Warn("failed to serve HTTP response", "error", wErr)
	}
}

func writeErrResponse(w http.ResponseWriter, err error, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	//nolint:errchkjson // Bad client!
	json.NewEncoder(w).Encode(errResponse{Error: err.Error()}) //nolint:errcheck
}
