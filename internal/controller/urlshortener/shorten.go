package urlshortener

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"go.uber.org/zap"
)

type ShortenRequest struct {
	LongURL string `json:"long_url"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

func (req *ShortenRequest) validate() error {
	if req.LongURL == "" {
		return errors.New("long_url is required")
	}

	parsedURL, err := url.ParseRequestURI(req.LongURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("invalid long_url format")
	}

	return nil
}

func (u *URLHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		u.logger.Warn("failed to decode shorten request body", zap.Error(err))
		http.Error(w, "bad request: invalid json body", http.StatusBadRequest)
		return
	}

	if err := req.validate(); err != nil {
		u.logger.Warn("shorten request validation failed", zap.Error(err))
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	shortKey, err := u.urlService.Shorten(r.Context(), req.LongURL)
	if err != nil {
		u.logger.Error("failed to shorten url in service", zap.Error(err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	resp := ShortenResponse{
		ShortURL: "http://localhost:8080/" + shortKey,
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		u.logger.Error("failed to encode shorten response", zap.Error(err))
		return
	}
}
