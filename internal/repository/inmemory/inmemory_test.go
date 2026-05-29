package inmemory_test

import (
	"context"
	"testing"

	"github.com/Sushka21/url-shortener/internal/entity"
	"github.com/Sushka21/url-shortener/internal/repository/inmemory"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryRepository(t *testing.T) {
	repo := inmemory.NewInMemoryRepository()
	ctx := context.Background()

	err := repo.Save(ctx, "key1", "https://google.com")
	assert.NoError(t, err)

	err = repo.Save(ctx, "key1", "https://yandex.ru")
	assert.ErrorIs(t, err, entity.ErrConflictURL)

	url, err := repo.GetByShortKey(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, "https://google.com", url)

	_, err = repo.GetByShortKey(ctx, "unknown")
	assert.ErrorIs(t, err, entity.ErrURLNotFound)
}
