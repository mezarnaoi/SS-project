package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func FuzzWithAuth(f *testing.F) {
	seedTokens := []string{
		"",
		"Bearer ",
		"Bearer invalid",
		"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6InRlc3QiLCJyb2xlIjoidXNlciJ9.invalid",
		"NotBearer token",
		"Bearer " + strings.Repeat("a", 500),
	}
	for _, token := range seedTokens {
		f.Add(token)
	}

	f.Fuzz(func(t *testing.T, authHeader string) {
		handler := withAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		if authHeader != "" {
			req.Header.Set("Authorization", authHeader)
		}
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		if rr.Code == 0 {
			t.Error("handler returned status 0")
		}
	})
}
