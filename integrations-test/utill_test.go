package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/Sushka21/url-shortener/internal/config"
	"github.com/caarlos0/env/v10"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
)

type ShortenRequest struct {
	LongURL string `json:"long_url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

type configEnv struct {
	db      *sql.DB
	BaseURL string
}

type testClients struct {
	HTTPClient *http.Client
}

func setup(t *testing.T) (*configEnv, *testClients) {
	t.Helper()

	appCfg := loadConfig(t)
	db := initDB(t, appCfg)
	cleanupDB(t, db)

	baseURL := fmt.Sprintf("http://localhost:%s", appCfg.HTTP.Port)
	waitForServices(t, baseURL, 15*time.Second)

	cfg := &configEnv{
		db:      db,
		BaseURL: baseURL,
	}

	clients := &testClients{
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}

	return cfg, clients
}

func loadConfig(t *testing.T) *config.Config {
	t.Helper()

	var cfg config.Config
	err := env.Parse(&cfg)
	require.NoError(t, err)

	return &cfg
}

func initDB(t *testing.T, cfg *config.Config) *sql.DB {
	t.Helper()

	db, err := sql.Open("pgx", cfg.ConstructPostgresURL())
	require.NoError(t, err)

	return db
}

func cleanupDB(t *testing.T, db *sql.DB) {
	t.Helper()

	t.Cleanup(func() {
		require.NoError(t, cleanDB(context.Background(), db))
		db.Close()
	})
}

func waitForServices(t *testing.T, baseURL string, timeout time.Duration) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	client := &http.Client{Timeout: 1 * time.Second}

	for {
		select {
		case <-ctx.Done():
			t.Fatalf("URL Shortener server %s did not become ready in %v", baseURL, timeout)
		default:
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/ping_test_health", nil)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}

		resp, err := client.Do(req)
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode < 500 {
			return
		}

		time.Sleep(200 * time.Millisecond)
	}
}

func cleanDB(ctx context.Context, db *sql.DB) error {
	var tables string

	err := db.QueryRowContext(ctx, `
		SELECT string_agg(
			quote_ident(schemaname) || '.' || quote_ident(tablename),
			', '
		)
		FROM pg_tables
		WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
		AND tablename NOT IN ('goose_db_version')
	`).Scan(&tables)

	if err != nil {
		return err
	}

	if tables == "" {
		return nil
	}

	query := "TRUNCATE TABLE " + tables + " RESTART IDENTITY CASCADE"
	_, err = db.ExecContext(ctx, query)
	return err
}

func shortenURL(client *http.Client, baseURL, longURL string) (string, error) {
	reqBody := ShortenRequest{LongURL: longURL}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(reqBody); err != nil {
		return "", fmt.Errorf("encode req body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, baseURL+"/shorten", &buf)
	if err != nil {
		return "", fmt.Errorf("create post request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("do post request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("expected 201, got %d", resp.StatusCode)
	}

	var respBody ShortenResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return "", fmt.Errorf("decode resp body: %w", err)
	}

	parts := strings.Split(respBody.ShortURL, "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid short URL format: %s", respBody.ShortURL)
	}

	shortKey := parts[len(parts)-1]
	return shortKey, nil
}

func resolveURL(client *http.Client, baseURL, shortKey string) (string, int, error) {
	req, err := http.NewRequest(http.MethodGet, baseURL+"/"+shortKey, nil)
	if err != nil {
		return "", 0, fmt.Errorf("create get request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("do get request: %w", err)
	}
	defer resp.Body.Close()

	return resp.Header.Get("Location"), resp.StatusCode, nil
}
