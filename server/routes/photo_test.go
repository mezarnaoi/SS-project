package routes_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"fmt"


	"go.uber.org/mock/gomock"

	"mqtt-streaming-server/domain"
	mock_domain "mqtt-streaming-server/mocks"
	"mqtt-streaming-server/routes"
)

func TestPhotoController_GetPhotos(t *testing.T) {
	tests := []struct {
		name             string
		userEmail        string
		mockPhotos       []*domain.Photo
		mockError        error
		expectedStatus   int
		expectedContains string
	}{
		{
			name:             "no photos",
			userEmail:        "empty@example.com",
			mockPhotos:       []*domain.Photo{},
			expectedStatus:   http.StatusOK,
			expectedContains: "[]",
		},
		{
			name:             "repository error",
			userEmail:        "error@example.com",
			mockError:        errors.New("db error"),
			expectedStatus:   http.StatusInternalServerError,
			expectedContains: "Failed to fetch photos",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
			ctlr := routes.PhotoController{PhotoRepository: mockRepo}

			req := httptest.NewRequest(http.MethodGet, "/photos", nil)
			ctx := context.WithValue(req.Context(), "email", tt.userEmail)
			req = req.WithContext(ctx)
			rr := httptest.NewRecorder()

			if tt.mockPhotos != nil || tt.mockError != nil {
				mockRepo.EXPECT().
					GetPhotos(ctx, gomock.Any()).
					Return(tt.mockPhotos, tt.mockError)
			}

			ctlr.GetPhotos(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}
			if tt.expectedContains != "" && !strings.Contains(rr.Body.String(), tt.expectedContains) {
				t.Errorf("expected body to contain %q, got %q", tt.expectedContains, rr.Body.String())
			}
		})
	}
}


func TestPhotoController_GetPhotos_MethodNotAllowed(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
	ctlr := routes.PhotoController{PhotoRepository: mockRepo}

	req := httptest.NewRequest(http.MethodPost, "/photos", nil) // Using POST instead of GET
	rr := httptest.NewRecorder()

	ctlr.GetPhotos(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Method not allowed") {
		t.Errorf("expected body to contain 'Method not allowed', got %q", rr.Body.String())
	}
}

func TestPhotoController_GetPhotos_InvalidTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
	ctlr := routes.PhotoController{PhotoRepository: mockRepo}

	req := httptest.NewRequest(http.MethodGet, "/photos?start=invalid&end=invalid", nil)
	rr := httptest.NewRecorder()

	ctlr.GetPhotos(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Invalid start timestamp") && !strings.Contains(rr.Body.String(), "Invalid end timestamp") {
		t.Errorf("expected body to contain 'Invalid start timestamp' or 'Invalid end timestamp', got %q", rr.Body.String())
	}
}

func FuzzPhotoController_GetPhotos(f *testing.F) {
	seedInputs := []struct{ start, end string }{
		{"", ""},
		{"1700000000", "1800000000"},
		{"invalid", "invalid"},
		{"0", "0"},
		{strings.Repeat("9", 100), "1"},
	}
	for _, s := range seedInputs {
		f.Add(s.start, s.end)
	}

	f.Fuzz(func(t *testing.T, start string, end string) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := mock_domain.NewMockPhotoRepository(ctrl)
		ctlr := routes.PhotoController{PhotoRepository: mockRepo}

		url := "/photos"
		if start != "" || end != "" {
			url = fmt.Sprintf("/photos?start=%s&end=%s", start, end)
		}

		req := httptest.NewRequest(http.MethodGet, url, nil)
		ctx := context.WithValue(req.Context(), "email", "test@example.com")
		ctx = context.WithValue(ctx, "role", "admin")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		mockRepo.EXPECT().
			GetPhotos(gomock.Any(), gomock.Any()).
			Return([]*domain.Photo{}, nil).
			AnyTimes()

		ctlr.GetPhotos(rr, req)

		if rr.Code == 0 {
			t.Error("handler returned status 0")
		}
	})
}
