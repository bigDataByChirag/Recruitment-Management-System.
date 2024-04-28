package handlers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"log/slog"
	"net/http"
	"rms/internals/auth"
	"rms/internals/models"
	"rms/internals/services"
	"strconv"
)

type Users struct {
	userService *services.UserService
	a           *auth.Auth
}

// NewUsers creates a new Users handler with the provided services and authentication
func NewUsers(us *services.UserService, a *auth.Auth) (*Users, error) {
	if us == nil || a == nil {
		return nil, errors.New("please provide all the values")
	}
	return &Users{
		userService: us,
		a:           a,
	}, nil
}

// CreateUser handles the creation of a new user via HTTP POST request
func (u Users) CreateUser(c *gin.Context) {
	// Set response header
	c.Header("Content-Type", "application/json")

	var newUser models.NewUser

	// Bind request body to the newUser struct
	if err := c.BindJSON(&newUser); err != nil {
		slog.Error("bind json:", slog.String("Error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		fmt.Println()
		return
	}

	// Create a new validator and register the custom role validator
	validate := validator.New()

	// Validate the newUser struct
	if err := validate.Struct(newUser); err != nil {
		slog.Error("validate struct:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Validation error"})
		fmt.Println()
		return
	}

	// Create the user using the user service
	_, err := u.userService.Create(&newUser)
	if err != nil {
		slog.Error("create:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
		fmt.Println()
		return
	}

	// If user creation is successful, return a success response
	c.JSON(http.StatusCreated, gin.H{"message": "User Registered Successfully"})
}

// customRoleValidator is a custom validation function for user roles
func customRoleValidator(fl validator.FieldLevel) bool {
	role := fl.Field().String()
	return role == "Admin" || role == "Applicant"
}

func (u Users) ProcessLoginIn(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	var authUser struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,password"`
		UserType string `json:"UserType" validate:"required"`
	}

	if err := c.BindJSON(&authUser); err != nil {
		slog.Error("bind json:", slog.String("Error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		fmt.Println()
		return
	}

	// Authenticate the user using the user service
	user, err := u.userService.Authenticate(authUser.Email, authUser.UserType, authUser.Password)
	if err != nil {
		slog.Error("autheticate:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		fmt.Println()
		return
	}

	// Generate a JWT token for the authenticated user
	tkn, err := u.a.GenerateToken(user.ID, user.UserType)
	if err != nil {
		slog.Error("generate token:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token Generate error"})
		fmt.Println()
		return
	}

	// Set the JWT token in the response header
	c.Header("Authorization", tkn)

	c.JSON(http.StatusCreated, gin.H{"message": tkn})
}

// GetApplicantsHandler retrieves a list of all users in the system (applicants).
func (u Users) GetApplicantsHandler(c *gin.Context) {
	// Extract the authenticated user from the context
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	authenticatedUser, ok := claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if the authenticated user is an admin
	if authenticatedUser.Roles != "Admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Only Admin users are allowed to access this API"})
		return
	}

	// Get the list of applicants
	applicants, err := u.userService.GetApplicants()
	if err != nil {
		slog.Error("get applicants:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Println()
		return
	}

	// Return the list of applicants
	c.JSON(http.StatusOK, gin.H{"applicants": applicants})
}

// GetApplicantsHandler retrieves a list of all users in the system (applicants).
func (u Users) GetApplicantHandler(c *gin.Context) {
	// Extract the authenticated user from the context
	claims, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	authenticatedUser, ok := claims.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Check if the authenticated user is an admin
	if authenticatedUser.Roles != "Admin" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Only Admin users are allowed to access this API"})
		return
	}

	// Get the applicant ID from the request URL
	applicantId, err := strconv.Atoi(c.Param("applicant_id"))
	if err != nil {
		slog.Error("string conversion:", slog.String("Error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid applicant ID"})
		fmt.Println()
		return
	}

	// Get the applicant by ID
	applicant, err := u.userService.GetApplicant(applicantId)
	if err != nil {
		slog.Error("get applicant:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		fmt.Println()
		return
	}

	// Return the applicant
	c.JSON(http.StatusOK, gin.H{"applicant": applicant})
}
