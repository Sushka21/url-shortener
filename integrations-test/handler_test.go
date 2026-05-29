package tests

import (
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestURLShortener(t *testing.T) {
	cfg, clients := setup(t)

	clients.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	longURL := "https://ozon.ru/category/elektronika-15500/"

	shortKey, err := shortenURL(clients.HTTPClient, cfg.BaseURL, longURL)
	require.NoError(t, err)
	require.Len(t, shortKey, 10)

	location, status, err := resolveURL(clients.HTTPClient, cfg.BaseURL, shortKey)
	require.NoError(t, err)
	require.Equal(t, http.StatusTemporaryRedirect, status)
	require.Equal(t, longURL, location)
}

func TestURLShortenerConcurrency(t *testing.T) {
	cfg, clients := setup(t)

	const workers = 50
	const requestsPerWorker = 200

	var wg sync.WaitGroup

	clients.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	for i := range workers {
		workerID := i
		wg.Go(func() {
			for j := range requestsPerWorker {
				longURL := fmt.Sprintf("https://ozon.ru/worker/%d/request/%d", workerID, j)

				shortKey, err := shortenURL(clients.HTTPClient, cfg.BaseURL, longURL)
				if err != nil {
					t.Errorf("worker %d failed to shorten URL: %v", workerID, err)
					return
				}

				location, status, err := resolveURL(clients.HTTPClient, cfg.BaseURL, shortKey)
				if err != nil {
					t.Errorf("worker %d failed to resolve URL: %v", workerID, err)
					return
				}

				if status != http.StatusTemporaryRedirect {
					t.Errorf("worker %d expected status 307, got %d", workerID, status)
					return
				}

				if location != longURL {
					t.Errorf("worker %d location mismatch: expected %s, got %s", workerID, longURL, location)
					return
				}
			}
		})
	}
	wg.Wait()
}

func TestURLShortenerSameURLConcurrency(t *testing.T) {
	cfg, clients := setup(t)

	const workers = 50
	const requestsPerWorker = 200

	var wg sync.WaitGroup

	clients.HTTPClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	const targetURL = "https://ozon.ru/shared-concurrency-link"

	for i := range workers {
		workerID := i
		wg.Go(func() {
			for range requestsPerWorker {
				shortKey, err := shortenURL(clients.HTTPClient, cfg.BaseURL, targetURL)
				if err != nil {
					t.Errorf("worker %d failed on duplicate insert: %v", workerID, err)
					return
				}

				if len(shortKey) != 10 {
					t.Errorf("worker %d got invalid key length: %d", workerID, len(shortKey))
					return
				}
			}
		})
	}

	wg.Wait()
}
