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
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const (
	DefaultCacheTTL     = time.Hour                                                                  // Cached items' expiry.
	FlushFrequency      = 5 * time.Minute                                                            // How often the cache should be checked for stale items.
	InstagramCDNDomain  = ".cdninstagram.com"                                                        // Default domain whence Instagram pictures are served.
	InstagramCDNTimeout = 10 * time.Second                                                           // Maximum time Instagram CDN can take to serve a picture.
	UserAgent           = "YahooMailProxy; https://help.yahoo.com/kb/yahoo-mail-proxy-SLN28749.html" // User-Agent header to use when downloading from Instagram
)

// httpDoer defines an interface to make HTTP requests.
type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

// cacheEntry defines how a picture should be stored in the cached.
type cacheEntry struct {
	contentType string    // File's content type
	data        []byte    // File's binary content
	expiry      time.Time // Entry's expiry date
}

// PicturesRelay is an helper that acts as a proxy for Instagram CDN, working around their CORS restrictions.
type PicturesRelay struct {
	cache    map[string]cacheEntry // Cache items map
	httpDoer httpDoer              // HTTP client
	lock     sync.Mutex            // Lock for flush() method
	logger   *slog.Logger          // Logger
	ttl      time.Duration         // Items' TTL.
}

// Cache stores a picture and its content type in the cache.
func (p *PicturesRelay) Cache(url, contentType string, picture []byte) {
	p.cache[url] = cacheEntry{
		contentType: contentType,
		data:        picture,
		expiry:      time.Now().Add(p.ttl),
	}
}

// Cached retrieves a picture and its content type from the cache.
func (p *PicturesRelay) Cached(url string) ([]byte, string, bool) {
	item, found := p.cache[url]
	if !found {
		return nil, "", false
	}

	return item.data, item.contentType, true
}

// Client overrides the defautl HTTP client that will be downloading files from Instagram.
func (p *PicturesRelay) Client(client httpDoer) *PicturesRelay {
	p.httpDoer = client

	return p
}

// ServeHTTP implements the HandlerFunc interface.
// It reads the picture's URL from the GET querystring (key: pictureURL) and then performs a lookup into its cache.
// If the picture is cached, it will be downloaded from Instagram, stored in the cache, and served to the client as is.
func (p *PicturesRelay) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pictureURL := r.URL.Query().Get("pictureURL")
	u, err := url.Parse(pictureURL)

	// Input validation.
	switch {
	case err != nil, pictureURL == "", u.Scheme != "https":
		p.logger.Debug("invalid URL", "pictureURL", pictureURL)
		w.WriteHeader(http.StatusBadRequest)

		return
	case !strings.HasSuffix(u.Hostname(), InstagramCDNDomain):
		p.logger.Debug("forbidden URL", "domain", u.Hostname(), "pictureURL", pictureURL)
		w.WriteHeader(http.StatusForbidden)

		return
	}

	// Cache hit.
	if data, ctype, found := p.Cached(pictureURL); found {
		w.Header().Set("Content-Type", ctype)
		w.WriteHeader(http.StatusOK)

		if _, err := w.Write(data); err != nil {
			p.logger.Warn("could not relay Instagram picture", "error", err)
		}

		return
	}

	// Cache miss - download from Instagram.
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, u.String(), nil)
	if err != nil {
		p.logger.Warn("could not create HTTP request", "error", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	req.Header.Set("User-Agent", UserAgent)

	res, err := p.httpDoer.Do(req)
	if res != nil && res.Body != nil {
		defer res.Body.Close()
	}

	// Response.
	switch {
	case err != nil:
		p.logger.Warn("could not download Instagram picture", "error", err)
		w.WriteHeader(http.StatusBadGateway)
	case res.StatusCode != http.StatusOK:
		p.logger.Warn("could not download Instagram picture", "http.response.status_code", res.StatusCode)
		w.WriteHeader(http.StatusBadGateway)
	default:
		ctype := res.Header.Get("Content-Type")

		data, err := io.ReadAll(res.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			p.logger.Error("could not relay Instagram picture", "error", err)

			return
		}

		p.Cache(pictureURL, ctype, data)
		w.Header().Set("Content-Type", ctype)

		if _, err := w.Write(data); err != nil {
			p.logger.Warn("could not relay Instagram picture", "error", err)
		}
	}
}

// TTL sets the lifespan of the next cached items.
func (p *PicturesRelay) TTL(ttl time.Duration) {
	p.ttl = ttl
}

// Watch starts a go routine that watches the cache and removes any expire entry.
// The goroutine will automatically terminate when the context is cancelled.
func (p *PicturesRelay) Watch(ctx context.Context, freq time.Duration) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(freq):
				p.flush()
			}
		}
	}()
}

// flush removes expired items from the cache.
func (p *PicturesRelay) flush() {
	p.logger.Debug("start flushing")

	start := time.Now()
	flushed := 0

	p.lock.Lock()
	defer p.lock.Unlock()

	for pictureURL, item := range p.cache {
		if start.Compare(item.expiry) == 1 {
			delete(p.cache, pictureURL)

			flushed++
		}
	}

	p.logger.Debug("done flushing", "count", flushed, "time.ms", time.Since(start).Milliseconds())
}

// DefaultPicturesRelay returns a PicturesRelay with default configuration.
func DefaultPicturesRelay(logger *slog.Logger) *PicturesRelay {
	return &PicturesRelay{
		cache:    make(map[string]cacheEntry, 0),
		httpDoer: &http.Client{Timeout: InstagramCDNTimeout}, //nolint:exhaustruct // defaults are ok
		lock:     sync.Mutex{},
		logger:   logger,
		ttl:      DefaultCacheTTL,
	}
}
