package models

import (
	"time"
)

// Job represents a job opening
type Job struct {
	ID                int       `json:"id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	PostedOn          time.Time `json:"posted_on"`
	TotalApplications int       `json:"total_applications"`
	CompanyName       string    `json:"company_name"`
	PostedBy          int       `json:"posted_by"`
}

// JobResponse represents the JSON response for a job
type NewJob struct {
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	PostedOn          time.Time `json:"posted_on"`
	TotalApplications int       `json:"total_applications"`
	CompanyName       string    `json:"company_name"`
	PostedBy          int       `json:"posted_by"`
}

// Applicant represents a user who applied for a job.
type Applicant struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}
