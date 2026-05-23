package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"

	"mqtt-streaming-server/domain"
	"mqtt-streaming-server/repository"
)

type UserController struct {
	UserRepository    domain.UserRepository
	RawUserRepository *repository.UserRepository
}

func InitUserRoutes(db *mongo.Database, mux *http.ServeMux) {
	rawRepo := repository.NewUserRepository(db)
	userController := &UserController{
		UserRepository:    rawRepo,
		RawUserRepository: rawRepo,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := rawRepo.EnsureDefaultAdmin(ctx); err != nil {
		fmt.Printf("Warning: failed to ensure default admin user: %v\n", err)
	}

	mux.HandleFunc("/register", userController.Register)
	mux.HandleFunc("/login", userController.Login)
	mux.Handle("/profile", withAuth(http.HandlerFunc(userController.GetProfile)))
	mux.Handle("/users", withAuth(withAdminOnly(http.HandlerFunc(userController.UsersCollection))))
	mux.Handle("/users/", withAuth(withAdminOnly(http.HandlerFunc(userController.UserByID))))
}

func (ctlr UserController) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))
	if req.Email == "" || req.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	// look for existing user
	existingUser, err := ctlr.UserRepository.FindByEmail(r.Context(), req.Email)
	if err != nil && err != mongo.ErrNoDocuments {
		http.Error(w, "Failed to check existing user", http.StatusInternalServerError)
		return
	}

	if existingUser != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	// Save the user to the database
	err = ctlr.UserRepository.Save(r.Context(), req.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintln(w, "User registered successfully")
}

func (ctlr UserController) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req domain.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req.Email = strings.ToLower(strings.TrimSpace(req.Email))

	// Check if the user exists
	user, err := ctlr.UserRepository.FindByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid email or password: "+err.Error(), http.StatusUnauthorized)
		return
	}

	// Verify the password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid email or password: "+err.Error(), http.StatusUnauthorized)
		return
	}

	claims := jwt.MapClaims{
		"email": user.Email,
		"role":  user.Role,
		"pages": normalizePages(user.Pages),
		"exp":   time.Now().Add(time.Hour * 24).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := jwtToken.SignedString([]byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"token": tokenString}); err != nil { // #nosec G104 -- encoding errors on http.ResponseWriter are non-actionable
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (ctlr UserController) GetProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	email, ok := r.Context().Value("email").(string)
	if !ok {
		http.Error(w, "Email not found in context", http.StatusUnauthorized)
		return
	}

	// Retrieve the user's profile from the database
	user, err := ctlr.UserRepository.FindByEmail(r.Context(), email)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	// Exclude the password from the response
	user.Password = ""
	if len(user.Pages) == 0 {
		user.Pages = []string{"reports"}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user) // #nosec G104,G117 -- password field has json:"-" tag and is explicitly cleared before encoding; ResponseWriter errors are non-actionable
}

type userMutationRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Role     string   `json:"role"`
	Pages    []string `json:"pages"`
}

type userResponse struct {
	ID    string   `json:"id"`
	Email string   `json:"email"`
	Role  string   `json:"role"`
	Pages []string `json:"pages"`
}

func validRole(role string) bool {
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "admin", "user":
		return true
	default:
		return false
	}
}

func normalizePages(input []string) []string {
	if len(input) == 0 {
		return []string{"reports"}
	}
	allowed := map[string]struct{}{
		"photos":     {},
		"devices":    {},
		"statistics": {},
		"reports":    {},
		"users":      {},
	}
	seen := map[string]struct{}{}
	pages := make([]string, 0, len(input))
	for _, raw := range input {
		page := strings.ToLower(strings.TrimSpace(raw))
		if _, ok := allowed[page]; !ok {
			continue
		}
		if _, exists := seen[page]; exists {
			continue
		}
		seen[page] = struct{}{}
		pages = append(pages, page)
	}
	if len(pages) == 0 {
		return []string{"reports"}
	}
	return pages
}

func (ctlr UserController) UsersCollection(w http.ResponseWriter, r *http.Request) {
	if ctlr.RawUserRepository == nil {
		http.Error(w, "Repository not initialized", http.StatusInternalServerError)
		return
	}

	switch r.Method {
	case http.MethodGet:
		users, err := ctlr.RawUserRepository.List(r.Context())
		if err != nil {
			http.Error(w, "Failed to list users", http.StatusInternalServerError)
			return
		}
		response := make([]userResponse, 0, len(users))
		for _, user := range users {
			response = append(response, userResponse{
				ID:    user.ID.Hex(),
				Email: user.Email,
				Role:  user.Role,
				Pages: normalizePages(user.Pages),
			})
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil { // #nosec G104
			http.Error(w, "Failed to encode users", http.StatusInternalServerError)
		}
	case http.MethodPost:
		var req userMutationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		req.Role = strings.ToLower(strings.TrimSpace(req.Role))
		req.Pages = normalizePages(req.Pages)

		if req.Email == "" || req.Password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}
		if !validRole(req.Role) {
			http.Error(w, "Role must be admin or user", http.StatusBadRequest)
			return
		}
		if req.Role == "admin" {
			req.Pages = []string{"photos", "devices", "statistics", "reports", "users"}
		}

		existing, err := ctlr.UserRepository.FindByEmail(r.Context(), req.Email)
		if err != nil && err != mongo.ErrNoDocuments {
			http.Error(w, "Failed to check existing user", http.StatusInternalServerError)
			return
		}
		if existing != nil {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		created, err := ctlr.RawUserRepository.Create(r.Context(), domain.User{
			Email:    req.Email,
			Password: string(hashedPassword),
			Role:     req.Role,
			Pages:    req.Pages,
		})
		if err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(userResponse{
			ID:    created.ID.Hex(),
			Email: created.Email,
			Role:  created.Role,
			Pages: normalizePages(created.Pages),
		}) // #nosec G104
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ctlr UserController) UserByID(w http.ResponseWriter, r *http.Request) {
	if ctlr.RawUserRepository == nil {
		http.Error(w, "Repository not initialized", http.StatusInternalServerError)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/users/")
	if id == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}
	if _, err := primitive.ObjectIDFromHex(id); err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var req userMutationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		setFields := map[string]any{}
		if strings.TrimSpace(req.Email) != "" {
			setFields["email"] = strings.ToLower(strings.TrimSpace(req.Email))
		}
		if strings.TrimSpace(req.Password) != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to hash password", http.StatusInternalServerError)
				return
			}
			setFields["password"] = string(hashedPassword)
		}
		if strings.TrimSpace(req.Role) != "" {
			role := strings.ToLower(strings.TrimSpace(req.Role))
			if !validRole(role) {
				http.Error(w, "Role must be admin or user", http.StatusBadRequest)
				return
			}
			setFields["role"] = role
			if role == "admin" {
				setFields["pages"] = []string{"photos", "devices", "statistics", "reports", "users"}
			}
		}
		if req.Pages != nil {
			setFields["pages"] = normalizePages(req.Pages)
		}

		if err := ctlr.RawUserRepository.UpdateByID(r.Context(), id, setFields); err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}

		updated, err := ctlr.RawUserRepository.FindByID(r.Context(), id)
		if err != nil {
			http.Error(w, "Failed to fetch updated user", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(userResponse{
			ID:    updated.ID.Hex(),
			Email: updated.Email,
			Role:  updated.Role,
			Pages: normalizePages(updated.Pages),
		}) // #nosec G104
	case http.MethodDelete:
		if err := ctlr.RawUserRepository.DeleteByID(r.Context(), id); err != nil {
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
