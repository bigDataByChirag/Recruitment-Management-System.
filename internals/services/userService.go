package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	"rms/internals/models"
	//"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

// UserService handles business logic related to user operations.
type UserService struct {
	db *sql.DB
}

// NewUserService creates a new UserService instance.
func NewUserService(db *sql.DB) (*UserService, error) {
	// Check if the database connection is nil
	if db == nil {
		return nil, errors.New("db connection cannot be nil")
	}
	return &UserService{db: db}, nil
}

func (us *UserService) Create(newUser *models.NewUser) (*models.User, error) {
	// Convert email and role to lowercase
	user := models.User{

		Name:        newUser.Name,
		Email:       strings.ToLower(newUser.Email),
		Address:     newUser.Address,
		UserType:    strings.ToLower(newUser.UserType),
		ProfileHead: newUser.ProfileHeadline,
	}

	// Hash the user's password using bcrypt
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(newUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	user.PasswordHash = string(hashedBytes)

	// Prepare the statement
	stmt, err := us.db.Prepare(`
	INSERT INTO users (
		Name, Email, Address, UserType, PasswordHash, ProfileHead
	) VALUES (?, ?, ?, ?, ?, ?)
`)
	if err != nil {
		return nil, fmt.Errorf("prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute the statement with parameters
	result, err := stmt.Exec(user.Name, user.Email, user.Address, user.UserType, user.PasswordHash, user.ProfileHead)
	if err != nil {
		return nil, fmt.Errorf("execute statement: %w", err)
	}

	// Get the ID of the inserted row
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert ID: %w", err)
	}

	// Assign the ID to the user struct
	user.ID = int(id)

	return &user, nil
}

// Authenticate verifies user credentials and returns the user if authentication is successful.
func (us *UserService) Authenticate(email, userType, password string) (*models.User, error) {
	// Convert email to lowercase
	email = strings.ToLower(email)

	// Create a user instance to store authentication details
	user := models.User{
		Email:    email,
		UserType: userType,
	}

	// Execute the SQL query to retrieve user information by email
	row := us.db.QueryRow("SELECT id, PasswordHash, UserType FROM users WHERE email = ?", email)
	err := row.Scan(&user.ID, &user.PasswordHash, &user.UserType)
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	// Compare the provided password with the hashed password in the database
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("authenticate: %w", err)
	}

	// Authentication successful, return the user
	return &user, nil
}

// GetApplicants retrieves a list of all users in the system along with their profiles.
func (us *UserService) GetApplicants() ([]models.UserWithProfile, error) {
	// Prepare the SQL query to fetch all users and their profiles
	query := `SELECT
    u.ID, u.Name, u.Email, u.Address, u.UserType, u.ProfileHead, p.Skills, p.Education, p.Experience, p.Phone
FROM
    users as u
JOIN
    resumes as p ON u.ID = p.userID`

	// Execute the query
	rows, err := us.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applicants: %v", err)
	}
	defer rows.Close()

	var results []models.UserWithProfile
	var jsonEducation []map[string]any
	var jsonSkills []string
	var jsonExperience []map[string]any
	for rows.Next() {
		var record struct {
			ID          int
			Name        string
			Email       string
			Address     string
			UserType    string
			ProfileHead string
			Skills      []byte
			Education   []byte
			Experience  []byte
			Phone       string
		}
		err := rows.Scan(&record.ID, &record.Name, &record.Email, &record.Address,
			&record.UserType, &record.ProfileHead, &record.Skills,
			&record.Education, &record.Experience, &record.Phone)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(record.Education, &jsonEducation)
		json.Unmarshal(record.Skills, &jsonSkills)
		json.Unmarshal(record.Experience, &jsonExperience)
		rec := models.UserWithProfile{
			ID:          record.ID,
			Name:        record.Name,
			Email:       record.Email,
			Address:     record.Address,
			UserType:    record.UserType,
			ProfileHead: record.ProfileHead,
			// assuming Skills, Education, and Experience in UserWithProfile are of type string
			Skills:     jsonSkills,
			Education:  jsonEducation,
			Experience: jsonExperience,
			Phone:      record.Phone,
		}
		results = append(results, rec)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return results, nil
}

// GetApplicants retrieves a list of all users in the system along with their profiles.
func (us *UserService) GetApplicant(applicantId int) ([]models.UserWithProfile, error) {
	// Prepare the SQL query to fetch all users and their profiles
	query := `SELECT
    u.ID, u.Name, u.Email, u.Address, u.UserType, u.ProfileHead, p.Skills, p.Education, p.Experience, p.Phone
FROM
    users as u
JOIN
    resumes as p ON u.ID = p.userID
WHERE u.ID = ?`

	// Execute the query
	rows, err := us.db.Query(query, applicantId)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch applicants: %v", err)
	}
	defer rows.Close()

	var results []models.UserWithProfile
	var jsonEducation []map[string]any
	var jsonSkills []string
	var jsonExperience []map[string]any
	for rows.Next() {
		var record struct {
			ID          int
			Name        string
			Email       string
			Address     string
			UserType    string
			ProfileHead string
			Skills      []byte
			Education   []byte
			Experience  []byte
			Phone       string
		}
		err := rows.Scan(&record.ID, &record.Name, &record.Email, &record.Address,
			&record.UserType, &record.ProfileHead, &record.Skills,
			&record.Education, &record.Experience, &record.Phone)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(record.Education, &jsonEducation)
		json.Unmarshal(record.Skills, &jsonSkills)
		json.Unmarshal(record.Experience, &jsonExperience)
		rec := models.UserWithProfile{
			ID:          record.ID,
			Name:        record.Name,
			Email:       record.Email,
			Address:     record.Address,
			UserType:    record.UserType,
			ProfileHead: record.ProfileHead,
			// assuming Skills, Education, and Experience in UserWithProfile are of type string
			Skills:     jsonSkills,
			Education:  jsonEducation,
			Experience: jsonExperience,
			Phone:      record.Phone,
		}

		results = append(results, rec)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return results, nil
}
