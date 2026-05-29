package urlshortener

import (
	"errors"
	"net/http"

	"github.com/Sushka21/url-shortener/internal/entity"
	"go.uber.org/zap"
)

func (u *URLHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	shortKey := r.PathValue("shortKey")
	if shortKey == "" {
		u.logger.Warn("resolve request missing shortKey parameter")
		http.Error(w, "bad request: missing short key", http.StatusBadRequest)
		return
	}

	if len(shortKey) != 10 {
		u.logger.Warn("invalid shortKey length", zap.String("key", shortKey))
		http.Error(w, "bad request: short key must be 10 characters long", http.StatusBadRequest)
		return
	}

	longURL, err := u.urlService.Resolve(r.Context(), shortKey)
	if err != nil {
		if errors.Is(err, entity.ErrURLNotFound) {
			u.logger.Info("short key not found in database", zap.String("key", shortKey))
			http.Error(w, "url not found", http.StatusNotFound)
			return
		}

		u.logger.Error("failed to resolve short key", zap.Error(err), zap.String("key", shortKey))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	u.logger.Info("successfully resolved URL", zap.String("key", shortKey), zap.String("target", longURL))
	http.Redirect(w, r, longURL, http.StatusTemporaryRedirect)
}
