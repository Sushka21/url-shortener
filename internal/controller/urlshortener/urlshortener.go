package urlshortener

import (
	"context"

	"go.uber.org/zap"
)

//go:generate mockgen -source=urlshortener.go -destination=mocks/urlshortener_mocks.go -package=mocks
type URLService interface {
	Shorten(ctx context.Context, longURL string) (string, error)
	Resolve(ctx context.Context, shortKey string) (string, error)
}

// var _ controller.URLServer = (*URLServer)(nil)

type URLHandler struct {
	urlService URLService
	logger     *zap.Logger
	baseURL    string
}

func NewURLHandler(
	urlService URLService,
	logger *zap.Logger,
	baseURL string,
) *URLHandler {
	return &URLHandler{
		urlService: urlService,
		logger:     logger,
		baseURL:    baseURL,
	}
}
