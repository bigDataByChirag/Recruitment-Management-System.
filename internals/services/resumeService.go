package services

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
)

// ResumeService handles operations related to resume data
type ResumeService struct {
	db *sql.DB
}

// NewUserService creates a new UserService instance.
func NewResumeService(db *sql.DB) (*ResumeService, error) {
	// Check if the database connection is nil
	if db == nil {
		return nil, errors.New("db connection cannot be nil")
	}
	return &ResumeService{db: db}, nil
}

// SaveResume saves the parsed resume data into the database
func (rs *ResumeService) SaveResume(userId string, parsedData map[string]interface{}, fileName string) error {
	// Convert parsedData to JSON strings
	educationJSON, err := json.Marshal(parsedData["education"])
	fmt.Println("save resume", string(educationJSON))
	if err != nil {
		return err
	}

	experienceJSON, err := json.Marshal(parsedData["experience"])
	if err != nil {
		return err
	}

	skillsJSON, err := json.Marshal(parsedData["skills"])
	if err != nil {
		return err
	}

	// Insert the parsed resume data into the database
	_, err = rs.db.Exec("INSERT INTO resumes (userId , resumeFileAddr, name, email, phone, education, experience, skills) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		userId, fileName, parsedData["name"], parsedData["email"], parsedData["phone"], educationJSON, experienceJSON, skillsJSON)
	if err != nil {
		return err
	}
	return nil
}
