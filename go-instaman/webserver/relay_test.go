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

package webserver_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/luca-arch/instaman/webserver"
	"github.com/stretchr/testify/assert"
)

var (
	pic0 = []byte("binary content 000") //nolint:gochecknoglobals
	pic1 = []byte("binary content 001") //nolint:gochecknoglobals
	pic2 = []byte("binary content 002") //nolint:gochecknoglobals
)

type mockHTTPDoer struct {
	body   string
	err    error
	status int
}

func (m *mockHTTPDoer) Do(_ *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	return &http.Response{
		Body:       io.NopCloser(bytes.NewBuffer([]byte(m.body))),
		Status:     fmt.Sprintf("%d %s", m.status, http.StatusText(m.status)),
		StatusCode: m.status,
	}, nil
}

func TestCache(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.TODO())
	t.Cleanup(cancel)

	cache := webserver.DefaultPicturesRelay(slog.New(slog.NewTextHandler(io.Discard, nil)))
	data := []byte("binary data")
	key := "item-key"

	cache.TTL(0)
	cache.Cache(key, "item content type", data)

	cachedData, cachedContentType, found := cache.Cached(key)

	assert.True(t, found)
	assert.Equal(t, data, cachedData)
	assert.Equal(t, "item content type", cachedContentType)

	_, _, found = cache.Cached("non existent key")
	assert.False(t, found)

	// Force flush, then sleep just enough time for the flush to finish.
	cache.Watch(ctx, 0)
	time.Sleep(50 * time.Millisecond)

	cachedData, cachedContentType, found = cache.Cached(key)
	assert.False(t, found)
	assert.Empty(t, cachedData)
	assert.Empty(t, cachedContentType)
}

func TestServeHTTP(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.TODO())
	t.Cleanup(cancel)

	anyValidURL := "https://example" + webserver.InstagramCDNDomain + "/picture.png"

	type args struct {
		pictureURL string
	}

	type fields struct {
		http mockHTTPDoer
	}

	type wants struct {
		contentType string
		picture     []byte
		status      int
	}

	tests := map[string]struct {
		args
		fields
		wants
	}{
		"success - picture 0": {
			args{
				pictureURL: "https://example" + webserver.InstagramCDNDomain + "/pic0.png",
			},
			fields{},
			wants{
				contentType: "image/png",
				picture:     pic0,
				status:      http.StatusOK,
			},
		},
		"success - picture 1": {
			args{
				pictureURL: "https://example" + webserver.InstagramCDNDomain + "/pic1.jpg",
			},
			fields{},
			wants{
				contentType: "image/jpeg",
				picture:     pic1,
				status:      http.StatusOK,
			},
		},
		"success - picture 2": {
			args{
				pictureURL: "https://example" + webserver.InstagramCDNDomain + "/pic2.png",
			},
			fields{},
			wants{
				contentType: "image/png",
				picture:     pic2,
				status:      http.StatusOK,
			},
		},
		"success - picture downloaded": {
			args{
				pictureURL: anyValidURL,
			},
			fields{
				mockHTTPDoer{
					body:   "downloaded binary content",
					status: http.StatusOK,
				},
			},
			wants{
				picture: []byte("downloaded binary content"),
				status:  http.StatusOK,
			},
		},
		"failure - URL is not HTTPS": {
			args{
				pictureURL: "http://example" + webserver.InstagramCDNDomain + "/picture.png",
			},
			fields{},
			wants{
				status: http.StatusBadRequest,
			},
		},
		"failure - URL is not CDN": {
			args{
				pictureURL: "https://example.com/picture.png",
			},
			fields{},
			wants{
				status: http.StatusForbidden,
			},
		},
		"failure - client error": {
			args{
				pictureURL: anyValidURL,
			},
			fields{
				mockHTTPDoer{
					err: errors.New("client.Do error"),
				},
			},
			wants{
				status: http.StatusBadGateway,
			},
		},
		"failure - 429 error": {
			args{
				pictureURL: anyValidURL,
			},
			fields{
				mockHTTPDoer{
					status: http.StatusTooManyRequests,
				},
			},
			wants{
				status: http.StatusBadGateway,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequestWithContext(ctx, http.MethodGet, "/instagram/picture?pictureURL="+url.QueryEscape(test.pictureURL), nil)
			rr := httptest.NewRecorder()

			picturesRelay(t, &test.fields.http).ServeHTTP(rr, req)

			assert.Equal(t, test.wants.status, rr.Result().StatusCode) //nolint:bodyclose // It will be closed.
			assert.Equal(t, string(test.wants.picture), rr.Body.String())
			assert.Equal(t, test.wants.contentType, rr.Header().Get("Content-Type"))

			rr.Result().Body.Close()
		})
	}
}

func picturesRelay(t *testing.T, mockClient *mockHTTPDoer) *webserver.PicturesRelay {
	t.Helper()

	r := webserver.DefaultPicturesRelay(slog.New(slog.NewTextHandler(io.Discard, nil)))

	r.Cache("https://example"+webserver.InstagramCDNDomain+"/pic0.png", "image/png", pic0)
	r.Cache("https://example"+webserver.InstagramCDNDomain+"/pic1.jpg", "image/jpeg", pic1)
	r.Cache("https://example"+webserver.InstagramCDNDomain+"/pic2.png", "image/png", pic2)

	if mockClient != nil {
		return r.Client(mockClient)
	}

	return r
}
