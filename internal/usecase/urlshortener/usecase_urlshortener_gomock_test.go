package urlshortener_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Sushka21/url-shortener/internal/entity"
	"github.com/Sushka21/url-shortener/internal/usecase/urlshortener"
	"github.com/Sushka21/url-shortener/internal/usecase/urlshortener/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestURLService_Shorten(t *testing.T) {
	t.Parallel()

	const (
		testLongURL      = "https://google.com"
		expectedShortKey = "feWMl_ok1G"
		anotherLongURL   = "https://yandex.ru"
	)

	type mockBehavior func(r *mocks.MockRepository)

	tests := []struct {
		name          string
		longURL       string
		mockBehavior  mockBehavior
		wantShortKey  string
		wantErrString string
	}{
		{
			name:    "Success New URL",
			longURL: testLongURL,
			mockBehavior: func(r *mocks.MockRepository) {
				r.EXPECT().
					Save(gomock.Any(), expectedShortKey, testLongURL).
					Return(nil).
					Times(1)
			},
			wantShortKey: expectedShortKey,
		},
		{
			name:          "Error Empty Long URL",
			longURL:       "",
			mockBehavior:  func(r *mocks.MockRepository) {},
			wantErrString: "long_url is required",
		},
		{
			name:    "Success URL Already Exists (Conflict but same URL)",
			longURL: testLongURL,
			mockBehavior: func(r *mocks.MockRepository) {
				r.EXPECT().
					Save(gomock.Any(), expectedShortKey, testLongURL).
					Return(entity.ErrConflictURL).
					Times(1)
				r.EXPECT().
					GetByShortKey(gomock.Any(), expectedShortKey).
					Return(testLongURL, nil).
					Times(1)
			},
			wantShortKey: expectedShortKey,
		},
		{
			name:    "Error Hash Collision (Conflict with different URL)",
			longURL: testLongURL,
			mockBehavior: func(r *mocks.MockRepository) {
				r.EXPECT().
					Save(gomock.Any(), expectedShortKey, testLongURL).
					Return(entity.ErrConflictURL).
					Times(1)
				r.EXPECT().
					GetByShortKey(gomock.Any(), expectedShortKey).
					Return(anotherLongURL, nil).
					Times(1)
			},
			wantErrString: "short key collision",
		},
		{
			name:    "Error Repository Save Failed",
			longURL: testLongURL,
			mockBehavior: func(r *mocks.MockRepository) {
				r.EXPECT().
					Save(gomock.Any(), expectedShortKey, testLongURL).
					Return(errors.New("db connection failure")).
					Times(1)
			},
			wantErrString: "db connection failure",
		},
		{
			name:    "Error Repository GetByShortKey Failed on Conflict",
			longURL: testLongURL,
			mockBehavior: func(r *mocks.MockRepository) {
				r.EXPECT().
					Save(gomock.Any(), expectedShortKey, testLongURL).
					Return(entity.ErrConflictURL).
					Times(1)
				r.EXPECT().
					GetByShortKey(gomock.Any(), expectedShortKey).
					Return("", errors.New("db internal error")).
					Times(1)
			},
			wantErrString: "db internal error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			logger := zap.NewNop()

			tt.mockBehavior(mockRepo)

			service := urlshortener.NewURLService(mockRepo, logger)

			shortKey, err := service.Shorten(context.Background(), tt.longURL)

			if tt.wantErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrString)
				assert.Empty(t, shortKey)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantShortKey, shortKey)
			}
		})
	}
}

func TestURLService_Resolve(t *testing.T) {
	t.Parallel()

	const (
		testShortKey = "feWMl_ok1G"
		testLongURL  = "https://google.com"
	)

	type mockBehavior func(r *mocks.MockRepository)

	tests := []struct {
		name          string
		shortKey      string
		mockBehavior  mockBehavior
		wantLongURL   string
		wantErrString string
	}{
		{
			name:     "Success URL Found",
			shortKey: testShortKey,
			mockBehavior: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetByShortKey(gomock.Any(), testShortKey).
					Return(testLongURL, nil).
					Times(1)
			},
			wantLongURL: testLongURL,
		},
		{
			name:     "Error Key Not Found or DB Error",
			shortKey: testShortKey,
			mockBehavior: func(r *mocks.MockRepository) {
				r.EXPECT().
					GetByShortKey(gomock.Any(), testShortKey).
					Return("", errors.New("database error or not found")).
					Times(1)
			},
			wantErrString: "database error or not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mocks.NewMockRepository(ctrl)
			logger := zap.NewNop()

			tt.mockBehavior(mockRepo)

			service := urlshortener.NewURLService(mockRepo, logger)
			longURL, err := service.Resolve(context.Background(), tt.shortKey)

			if tt.wantErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrString)
				assert.Empty(t, longURL)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLongURL, longURL)
			}
		})
	}
}
