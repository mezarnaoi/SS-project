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
	userRepo := repository.NewUserRepository(db)
	userController := &UserController{
		UserRepository:    userRepo,
		RawUserRepository: userRepo,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := userRepo.EnsureDefaultAdmin(ctx); err != nil {
		fmt.Println("failed to ensure default admin:", err)
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
		"pages": user.Pages,
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
	Password string   `json:"password,omitempty"`
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
	return role == "admin" || role == "user"
}

func normalizePages(pages []string, role string) []string {
	if role == "admin" {
		return []string{"photos", "devices", "statistics", "reports", "users"}
	}
	if len(pages) == 0 {
		return []string{"reports"}
	}

	allowed := map[string]bool{
		"photos":     true,
		"devices":    true,
		"statistics": true,
		"reports":    true,
	}
	seen := map[string]bool{}
	normalized := make([]string, 0, len(pages))
	for _, p := range pages {
		page := strings.ToLower(strings.TrimSpace(p))
		if allowed[page] && !seen[page] {
			seen[page] = true
			normalized = append(normalized, page)
		}
	}
	if len(normalized) == 0 {
		return []string{"reports"}
	}
	return normalized
}

func (ctlr UserController) UsersCollection(w http.ResponseWriter, r *http.Request) {
	if ctlr.RawUserRepository == nil {
		http.Error(w, "User repository not initialized", http.StatusInternalServerError)
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
			pages := user.Pages
			if len(pages) == 0 {
				pages = []string{"reports"}
			}
			if user.Role == "admin" {
				pages = []string{"all"}
			}
			response = append(response, userResponse{
				ID:    user.ID.Hex(),
				Email: user.Email,
				Role:  user.Role,
				Pages: pages,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response) // #nosec G104
	case http.MethodPost:
		var req userMutationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		req.Role = strings.ToLower(strings.TrimSpace(req.Role))
		if !validRole(req.Role) {
			http.Error(w, "Invalid role", http.StatusBadRequest)
			return
		}
		if req.Email == "" || req.Password == "" {
			http.Error(w, "Email and password are required", http.StatusBadRequest)
			return
		}

		existingUser, err := ctlr.RawUserRepository.FindByEmail(r.Context(), req.Email)
		if err != nil && err != mongo.ErrNoDocuments {
			http.Error(w, "Failed to check existing user", http.StatusInternalServerError)
			return
		}
		if existingUser != nil {
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}

		user := domain.User{
			Email:    req.Email,
			Password: string(hashedPassword),
			Role:     req.Role,
			Pages:    normalizePages(req.Pages, req.Role),
		}
		if err := ctlr.RawUserRepository.Create(r.Context(), user); err != nil {
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (ctlr UserController) UserByID(w http.ResponseWriter, r *http.Request) {
	if ctlr.RawUserRepository == nil {
		http.Error(w, "User repository not initialized", http.StatusInternalServerError)
		return
	}
	userID := strings.TrimPrefix(r.URL.Path, "/users/")
	if userID == "" {
		http.Error(w, "User ID required", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case http.MethodPatch:
		var req userMutationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		req.Email = strings.ToLower(strings.TrimSpace(req.Email))
		req.Role = strings.ToLower(strings.TrimSpace(req.Role))
		if !validRole(req.Role) {
			http.Error(w, "Invalid role", http.StatusBadRequest)
			return
		}
		if req.Email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}

		existingUser, err := ctlr.RawUserRepository.FindByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		update := map[string]any{
			"email": req.Email,
			"role":  req.Role,
			"pages": normalizePages(req.Pages, req.Role),
		}
		if req.Password != "" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				http.Error(w, "Failed to hash password", http.StatusInternalServerError)
				return
			}
			update["password"] = string(hashedPassword)
		}

		if existingUser.Email == "admin@test.com" {
			update["role"] = "admin"
			update["pages"] = []string{"photos", "devices", "statistics", "reports", "users"}
		}

		if err := ctlr.RawUserRepository.UpdateByID(r.Context(), userID, update); err != nil {
			http.Error(w, "Failed to update user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		existingUser, err := ctlr.RawUserRepository.FindByID(r.Context(), userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		if existingUser.Email == "admin@test.com" {
			http.Error(w, "Default admin cannot be deleted", http.StatusForbidden)
			return
		}
		if err := ctlr.RawUserRepository.DeleteByID(r.Context(), userID); err != nil {
			http.Error(w, "Failed to delete user", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
