package routes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidRole(t *testing.T) {
	if !validRole("admin") {
		t.Fatal("admin should be valid")
	}
	if !validRole("user") {
		t.Fatal("user should be valid")
	}
	if validRole("superadmin") {
		t.Fatal("superadmin should be invalid")
	}
}

func TestNormalizePages(t *testing.T) {
	adminPages := normalizePages([]string{"reports"}, "admin")
	if len(adminPages) != 5 {
		t.Fatalf("expected 5 admin pages, got %d", len(adminPages))
	}

	userPages := normalizePages([]string{" reports ", "photos", "photos", "invalid"}, "user")
	if len(userPages) != 2 {
		t.Fatalf("expected 2 user pages, got %d", len(userPages))
	}

	defaultPages := normalizePages(nil, "user")
	if len(defaultPages) != 1 || defaultPages[0] != "reports" {
		t.Fatalf("expected default reports page, got %#v", defaultPages)
	}
}

func TestExtractPages(t *testing.T) {
	pages, err := extractPages(nil)
	if err != nil {
		t.Fatalf("extractPages(nil) unexpected error: %v", err)
	}
	if len(pages) != 1 || pages[0] != "reports" {
		t.Fatalf("unexpected default pages: %#v", pages)
	}

	pages, err = extractPages([]any{"photos", "reports"})
	if err != nil {
		t.Fatalf("extractPages([]any) unexpected error: %v", err)
	}
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d", len(pages))
	}

	if _, err := extractPages([]any{"photos", 3}); err == nil {
		t.Fatal("expected error for invalid pages claim")
	}
}

func TestWithAnyPageAndAdminOnly(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// user with reports permission
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx := context.WithValue(req.Context(), "role", "user")
	ctx = context.WithValue(ctx, "pages", []string{"reports"})
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	withAnyPage(next, "reports").ServeHTTP(rr, req)
	if rr.Code != http.StatusOK || !nextCalled {
		t.Fatalf("expected allowed access, code=%d called=%v", rr.Code, nextCalled)
	}

	// user without permission
	nextCalled = false
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx2 := context.WithValue(req2.Context(), "role", "user")
	ctx2 = context.WithValue(ctx2, "pages", []string{"photos"})
	req2 = req2.WithContext(ctx2)
	rr2 := httptest.NewRecorder()
	withAnyPage(next, "reports").ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusForbidden || nextCalled {
		t.Fatalf("expected forbidden access, code=%d called=%v", rr2.Code, nextCalled)
	}

	// admin only middleware
	nextCalled = false
	req3 := httptest.NewRequest(http.MethodGet, "/", nil)
	ctx3 := context.WithValue(req3.Context(), "role", "admin")
	req3 = req3.WithContext(ctx3)
	rr3 := httptest.NewRecorder()
	withAdminOnly(next).ServeHTTP(rr3, req3)
	if rr3.Code != http.StatusOK || !nextCalled {
		t.Fatalf("expected admin access, code=%d called=%v", rr3.Code, nextCalled)
	}
}

func TestUsersControllerRepositoryNotInitialized(t *testing.T) {
	ctlr := UserController{}

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	rr := httptest.NewRecorder()
	ctlr.UsersCollection(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr.Code)
	}

	req2 := httptest.NewRequest(http.MethodPatch, "/users/123", nil)
	rr2 := httptest.NewRecorder()
	ctlr.UserByID(rr2, req2)
	if rr2.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", rr2.Code)
	}
}
