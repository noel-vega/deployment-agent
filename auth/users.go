package auth

import (
	"fmt"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// User represents a user account
type User struct {
	Username     string
	PasswordHash string
}

// users contains the registered user accounts
var users = make(map[string]*User)

// InitializeUsers loads users from environment variables
func InitializeUsers() error {
	// Load admin user from environment - REQUIRED
	adminUsername := os.Getenv("ADMIN_USERNAME")
	if adminUsername == "" {
		return fmt.Errorf("ADMIN_USERNAME environment variable is required")
	}

	adminPassword := os.Getenv("ADMIN_PASSWORD")
	if adminPassword == "" {
		return fmt.Errorf("ADMIN_PASSWORD environment variable is required")
	}

	// Validate password strength (minimum requirements)
	if len(adminPassword) < 8 {
		return fmt.Errorf("ADMIN_PASSWORD must be at least 8 characters long")
	}

	// Generate bcrypt hash for the admin password
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Add admin user
	users[adminUsername] = &User{
		Username:     adminUsername,
		PasswordHash: string(hash),
	}

	log.Printf("Admin user initialized: %s", adminUsername)
	return nil
}

// ValidateCredentials checks if username and password are correct
func ValidateCredentials(username, password string) error {
	user, exists := users[username]
	if !exists {
		return fmt.Errorf("invalid credentials")
	}

	// Compare provided password with stored hash
	err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return fmt.Errorf("invalid credentials")
	}

	return nil
}

// AddUser adds a new user (helper for runtime user management)
func AddUser(username, password string) error {
	// Check if user already exists
	if _, exists := users[username]; exists {
		return fmt.Errorf("user already exists")
	}

	// Generate bcrypt hash
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	users[username] = &User{
		Username:     username,
		PasswordHash: string(hash),
	}

	return nil
}

// GetUserCount returns the number of registered users
func GetUserCount() int {
	return len(users)
}

// UserExists checks if a username exists
func UserExists(username string) bool {
	_, exists := users[username]
	return exists
}
