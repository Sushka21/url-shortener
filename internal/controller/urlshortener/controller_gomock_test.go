package urlshortener_test

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Sushka21/url-shortener/internal/controller/urlshortener"
	"github.com/Sushka21/url-shortener/internal/controller/urlshortener/mocks"
	"github.com/Sushka21/url-shortener/internal/entity"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"
)

func TestURLHandler_Shorten(t *testing.T) {
	t.Parallel()

	const (
		testLongURL  = "https://google.com"
		testShortKey = "feWMl_ok1G"
	)

	type mockBehavior func(s *mocks.MockURLService)

	tests := []struct {
		name           string
		requestBody    string
		mockBehavior   mockBehavior
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "Success created",
			requestBody: `{"long_url": "https://google.com"}`,
			mockBehavior: func(s *mocks.MockURLService) {
				s.EXPECT().
					Shorten(gomock.Any(), testLongURL).
					Return(testShortKey, nil).
					Times(1)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   `{"short_url":"http://localhost:8080/feWMl_ok1G"}`,
		},
		{
			name:           "Error Invalid JSON Body",
			requestBody:    `{"long_url": `,
			mockBehavior:   func(s *mocks.MockURLService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "bad request: invalid json body",
		},
		{
			name:           "Error validation Failed (Empty URL)",
			requestBody:    `{"long_url": ""}`,
			mockBehavior:   func(s *mocks.MockURLService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "long_url is required",
		},
		{
			name:           "Error validation Failed (Invalid URL Format)",
			requestBody:    `{"long_url": "not-a-valid-url"}`,
			mockBehavior:   func(s *mocks.MockURLService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid long_url format",
		},
		{
			name:        "Error service Internal Failure",
			requestBody: `{"long_url": "https://google.com"}`,
			mockBehavior: func(s *mocks.MockURLService) {
				s.EXPECT().
					Shorten(gomock.Any(), testLongURL).
					Return("", errors.New("some service error")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "internal server error",
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)
			logger := zap.NewNop()

			tt.mockBehavior(mockService)

			handler := urlshortener.NewURLHandler(mockService, logger, "")

			req, err := http.NewRequest(http.MethodPost, "/shorten", bytes.NewBufferString(tt.requestBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()

			handler.Shorten(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code, "The status code doesn't match")

			assert.Contains(t, strings.TrimSpace(w.Body.String()), tt.expectedBody, "The response body does not match")
		})
	}
}

func TestURLHandler_Resolve(t *testing.T) {
	t.Parallel()

	const (
		validShortKey   = "feWMl_ok1G"
		invalidShortKey = "short"
		testLongURL     = "https://google.com"
	)

	type mockBehavior func(s *mocks.MockURLService)

	tests := []struct {
		name           string
		shortKeyParam  string
		mockBehavior   mockBehavior
		expectedStatus int
		expectedHeader string
	}{
		{
			name:          "Success Temporary Redirect",
			shortKeyParam: validShortKey,
			mockBehavior: func(s *mocks.MockURLService) {
				s.EXPECT().
					Resolve(gomock.Any(), validShortKey).
					Return(testLongURL, nil).
					Times(1)
			},
			expectedStatus: http.StatusTemporaryRedirect,
			expectedHeader: testLongURL,
		},
		{
			name:           "Error Missing ShortKey",
			shortKeyParam:  "",
			mockBehavior:   func(s *mocks.MockURLService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Error Invalid Key Length",
			shortKeyParam:  invalidShortKey,
			mockBehavior:   func(s *mocks.MockURLService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:          "Error URL Not Found (404)",
			shortKeyParam: validShortKey,
			mockBehavior: func(s *mocks.MockURLService) {
				s.EXPECT().
					Resolve(gomock.Any(), validShortKey).
					Return("", entity.ErrURLNotFound).
					Times(1)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:          "Error Service Internal Failure (500)",
			shortKeyParam: validShortKey,
			mockBehavior: func(s *mocks.MockURLService) {
				s.EXPECT().
					Resolve(gomock.Any(), validShortKey).
					Return("", errors.New("some db crash")).
					Times(1)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockURLService(ctrl)
			logger := zap.NewNop()

			tt.mockBehavior(mockService)

			handler := urlshortener.NewURLHandler(mockService, logger, "")

			req, err := http.NewRequest(http.MethodGet, "/"+tt.shortKeyParam, nil)
			assert.NoError(t, err)

			if tt.shortKeyParam != "" {
				req.SetPathValue("shortKey", tt.shortKeyParam)
			}

			w := httptest.NewRecorder()

			handler.Resolve(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.expectedHeader, w.Header().Get("Location"))
			}
		})
	}
}
