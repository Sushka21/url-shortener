package controller

import (
	"net/http"

	"github.com/Sushka21/url-shortener/internal/controller/urlshortener"
	"go.uber.org/zap"
)

type URLHandler interface {
	Shorten(w http.ResponseWriter, r *http.Request)
	Resolve(w http.ResponseWriter, r *http.Request)
}

type API struct {
	urlServer URLHandler
	logger    *zap.Logger
}

func New(urlService urlshortener.URLService, logger *zap.Logger, baseURL string) *API {
	return &API{
		urlServer: urlshortener.NewURLHandler(urlService, logger, baseURL),
		logger:    logger,
	}
}

func (a *API) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /", a.urlServer.Shorten)
	mux.HandleFunc("GET /{shortKey}", a.urlServer.Resolve)
}
