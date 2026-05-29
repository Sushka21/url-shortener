package urlshortener

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/Sushka21/url-shortener/internal/entity"
	"go.uber.org/zap"
)

//go:generate mockgen -source=urlshortener.go -destination=mocks/urlshortener_mocks.go -package=mocks
type (
	Repository interface {
		Save(ctx context.Context, shortKey string, longURL string) error
		GetByShortKey(ctx context.Context, shortKey string) (string, error)
	}
)

type urlService struct {
	repo   Repository
	logger *zap.Logger
}

func NewURLService(repo Repository, logger *zap.Logger) *urlService {
	return &urlService{
		repo:   repo,
		logger: logger,
	}
}

const (
	alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	base     = len(alphabet)
)

func (u *urlService) Shorten(ctx context.Context, longURL string) (string, error) {
	if longURL == "" {
		return "", errors.New("long_url is required")
	}
	hash := sha256.Sum256([]byte(longURL))
	data := make([]byte, 10)

	for i := range 10 {
		data[i] = alphabet[int(hash[i])%base]
	}
	shortKey := string(data)
	if err := u.repo.Save(ctx, shortKey, longURL); err != nil {
		if errors.Is(err, entity.ErrConflictURL) {
			existingLongURL, getErr := u.repo.GetByShortKey(ctx, shortKey)

			if getErr != nil {
				u.logger.Error("failed to check existing key on conflict", zap.Error(getErr))
				return "", getErr
			}

			if existingLongURL == longURL {
				u.logger.Info("url already exists in database, returning existing key", zap.String("key", shortKey))
				return shortKey, nil
			}

			u.logger.Warn("collision Same short key for different URLs",
				zap.String("key", shortKey),
				zap.String("existing_url", existingLongURL),
				zap.String("new_url", longURL),
			)

			return "", errors.New("short key collision")
		}
		return "", err
	}
	return shortKey, nil
}

func (u *urlService) Resolve(ctx context.Context, shortKey string) (string, error) {
	longURL, err := u.repo.GetByShortKey(ctx, shortKey)
	if err != nil {
		return "", err
	}
	return longURL, nil
}
