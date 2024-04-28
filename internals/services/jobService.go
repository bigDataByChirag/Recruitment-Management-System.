package services

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"rms/internals/models"
	"strconv"
	"time"
)

type JobService struct {
	db *sql.DB
}

func NewJobService(db *sql.DB) *JobService {
	return &JobService{
		db: db,
	}
}

func (js *JobService) CreateJob(newJob *models.NewJob, postedByID string) (*models.Job, error) {
	// Add postedOn timestamp
	userId, err := strconv.Atoi(postedByID)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	job := models.Job{
		Title:       newJob.Title,
		Description: newJob.Description,
		PostedOn:    time.Now(),
		CompanyName: newJob.CompanyName,
		PostedBy:    userId,
	}
	newJob.PostedOn = time.Now()

	// Insert the job into the database
	result, err := js.db.Exec("INSERT INTO jobs (title, description, posted_on, total_applications, company_name, posted_by) VALUES (?, ?, ?, ?, ?, ?)",
		newJob.Title, newJob.Description, newJob.PostedOn, newJob.TotalApplications, newJob.CompanyName, userId)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Retrieve the ID of the inserted job
	id, err := result.LastInsertId()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Assign the ID to the job struct
	job.ID = int(id)

	return &job, nil
}

// GetJobs retrieves all job openings from the database
func (js *JobService) GetJobs() ([]*models.Job, error) {
	// Prepare the SQL query to fetch job openings
	query := `
		SELECT id, title, description, posted_on, total_applications, company_name, posted_by
		FROM jobs
	`

	// Execute the query
	rows, err := js.db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, fmt.Errorf("failed to fetch job openings: %v", err)
	}
	defer rows.Close()

	// Initialize a slice to store job openings
	var jobs []*models.Job

	// Iterate over the rows and scan each job opening into a struct
	for rows.Next() {
		var job models.Job
		var postedOnStr string // Create a variable of type string to store the posted_on column as a string
		err := rows.Scan(&job.ID, &job.Title, &job.Description, &postedOnStr, &job.TotalApplications, &job.CompanyName, &job.PostedBy)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("failed to scan job opening: %v", err)
		}

		// Parse the posted_on string into a time.Time object
		postedOn, err := time.Parse("2006-01-02 15:04:05", postedOnStr)
		if err != nil {
			log.Println(err)
			return nil, fmt.Errorf("failed to parse posted_on: %v", err)
		}
		job.PostedOn = postedOn

		jobs = append(jobs, &job)
	}

	// Check for any errors during rows iteration
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	return jobs, nil
}

// ApplyForJob applies for a job by creating a new application record in the database
func (js *JobService) ApplyForJob(userID string, jobID int) error {
	// Check if the user has already applied for the job
	hasApplied, err := js.HasApplied(userID, jobID)
	if err != nil {
		return fmt.Errorf("failed to check application status: %v", err)
	}
	if hasApplied {
		return errors.New("user has already applied for this job")
	}

	// Prepare the SQL query to insert a new application record
	userId, err := strconv.Atoi(userID)
	if err != nil {
		return fmt.Errorf("failed to convert user ID to int: %v", err)
	}
	query := "INSERT INTO applications (user_id, job_id) VALUES (?)"

	// Execute the query
	_, err = js.db.Exec(query, userId, jobID)
	if err != nil {
		return fmt.Errorf("failed to apply for job: %v", err)
	}

	// Update the total_applications count in the jobs table
	updateQuery := "UPDATE jobs SET total_applications = total_applications + 1 WHERE id = ?"
	_, err = js.db.Exec(updateQuery, jobID)
	if err != nil {
		return fmt.Errorf("failed to update total_applications count: %v", err)
	}

	return nil
}

// HasApplied checks if the user has already applied for the specified job
func (js *JobService) HasApplied(userID string, jobID int) (bool, error) {
	// Prepare the SQL query to check if the user has already applied for the job
	userId, err := strconv.Atoi(userID)
	if err != nil {
		return false, fmt.Errorf("failed to convert user ID to int: %v", err)
	}
	query := "SELECT COUNT(*) FROM applications WHERE user_id = ? AND job_id = ?"

	// Execute the query
	var count int
	err = js.db.QueryRow(query, userId, jobID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check application status: %v", err)
	}

	return count > 0, nil
}

// GetJobAndApplicants retrieves information about a specific job and its applicants.
func (js *JobService) GetJobAndApplicants(jobID int) (*models.Job, []*models.Applicant, error) {
	// Initialize variables to store job and applicants
	job := &models.Job{}
	var applicants []*models.Applicant

	// Fetch job details
	jobQuery := "SELECT id, title, description, posted_on, total_applications, company_name, posted_by FROM jobs WHERE id = ?"
	// Create a variable of type string to store the posted_on column as a string
	var postedOnStr string
	err := js.db.QueryRow(jobQuery, jobID).Scan(&job.ID, &job.Title, &job.Description, &postedOnStr, &job.TotalApplications, &job.CompanyName, &job.PostedBy)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch job details: %v", err)
	}

	// Parse posted_on string into time.Time
	postedOn, err := time.Parse("2006-01-02 15:04:05", postedOnStr)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse posted_on: %v", err)
	}
	job.PostedOn = postedOn

	// Fetch applicants for the job
	applicantsQuery := "SELECT u.name, u.email FROM users u JOIN applications a ON u.id = a.user_id WHERE a.job_id = ?"
	rows, err := js.db.Query(applicantsQuery, jobID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch applicants: %v", err)
	}
	defer rows.Close()

	// Iterate over the rows and scan each applicant into a struct
	for rows.Next() {
		var applicant models.Applicant
		err := rows.Scan(&applicant.Name, &applicant.Email)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan applicant: %v", err)
		}
		applicants = append(applicants, &applicant)
	}

	// Check for any errors during rows iteration
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	return job, applicants, nil
}
