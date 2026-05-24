package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidRole(t *testing.T) {
	if !validRole("admin") {
		t.Fatal("expected admin to be valid")
	}
	if !validRole("user") {
		t.Fatal("expected user to be valid")
	}
	if validRole("superadmin") {
		t.Fatal("expected superadmin to be invalid")
	}
}

func TestNormalizePages(t *testing.T) {
	pages := normalizePages([]string{"reports", " photos ", "invalid", "reports"})
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}
	if pages[0] != "reports" || pages[1] != "photos" {
		t.Fatalf("unexpected normalized pages: %#v", pages)
	}

	defaultPages := normalizePages(nil)
	if len(defaultPages) != 1 || defaultPages[0] != "reports" {
		t.Fatalf("expected default reports page, got %#v", defaultPages)
	}
}

func TestExtractPages(t *testing.T) {
	pages, err := extractPages(nil)
	if err != nil {
		t.Fatalf("unexpected error for nil pages: %v", err)
	}
	if len(pages) != 1 || pages[0] != "reports" {
		t.Fatalf("unexpected pages for nil: %#v", pages)
	}

	pages, err = extractPages([]any{"reports", "devices"})
	if err != nil {
		t.Fatalf("unexpected error for valid []any pages: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %#v", pages)
	}

	_, err = extractPages(123)
	if err == nil {
		t.Fatal("expected error for invalid pages type")
	}
}

func TestWithAnyPageAndAdminOnly(t *testing.T) {
	protected := withAnyPage(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), "devices")

	req := httptest.NewRequest(http.MethodGet, "/devices", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", "user"))
	req = req.WithContext(context.WithValue(req.Context(), "pages", []string{"reports", "devices"}))
	rr := httptest.NewRecorder()
	protected.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for page access, got %d", rr.Code)
	}

	reqDenied := httptest.NewRequest(http.MethodGet, "/devices", nil)
	reqDenied = reqDenied.WithContext(context.WithValue(reqDenied.Context(), "role", "user"))
	reqDenied = reqDenied.WithContext(context.WithValue(reqDenied.Context(), "pages", []string{"reports"}))
	rrDenied := httptest.NewRecorder()
	protected.ServeHTTP(rrDenied, reqDenied)
	if rrDenied.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing page access, got %d", rrDenied.Code)
	}

	adminOnly := withAdminOnly(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	adminReq := httptest.NewRequest(http.MethodDelete, "/users/1", nil)
	adminReq = adminReq.WithContext(context.WithValue(adminReq.Context(), "role", "admin"))
	adminRes := httptest.NewRecorder()
	adminOnly.ServeHTTP(adminRes, adminReq)
	if adminRes.Code != http.StatusNoContent {
		t.Fatalf("expected 204 for admin, got %d", adminRes.Code)
	}
}

func TestUsersControllerRepositoryNotInitialized(t *testing.T) {
	controller := UserController{}
	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()
	controller.UsersCollection(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when repository not initialized, got %d", rr.Code)
	}
}
