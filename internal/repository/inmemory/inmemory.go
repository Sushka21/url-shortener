package inmemory

import (
	"context"
	"sync"

	"github.com/Sushka21/url-shortener/internal/entity"
)

type InMemoryRepository struct {
	mutex sync.Mutex
	data  map[string]string
}

func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		data: make(map[string]string),
	}
}
func (r *InMemoryRepository) Save(ctx context.Context, shortKey string, longURL string) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if _, ok := r.data[shortKey]; ok {
		return entity.ErrConflictURL
	}
	r.data[shortKey] = longURL
	return nil
}

func (r *InMemoryRepository) GetByShortKey(ctx context.Context, shortKey string) (string, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	v, ok := r.data[shortKey]
	if !ok {
		return "", entity.ErrURLNotFound
	}
	return v, nil
}
