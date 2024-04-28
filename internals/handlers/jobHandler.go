package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"rms/internals/auth"
	"rms/internals/models"
	"rms/internals/services"
	"strconv"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	jobService  *services.JobService
	authService *auth.Auth
}

func NewJobHandler(js *services.JobService, authService *auth.Auth) *JobHandler {
	return &JobHandler{
		jobService:  js,
		authService: authService,
	}
}

func (jh *JobHandler) CreateJob(c *gin.Context) {
	// Extract user details from the authentication token
	cl, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "StatusUnauthorized"})
		return
	}
	claims, ok := cl.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "StatusUnauthorized"})
		return
	}

	// Parse JSON request body into job struct
	var newJob models.NewJob
	if err := c.BindJSON(&newJob); err != nil {
		slog.Error("Bind Jason:", slog.String("Error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		fmt.Println()
		return
	}

	// Create the job using the job service
	createdJob, err := jh.jobService.CreateJob(&newJob, claims.Subject)
	if err != nil {
		slog.Error("create job:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create job"})
		fmt.Println()
		return
	}

	// Return the created job in the response
	c.JSON(http.StatusCreated, createdJob)
}

// GetJobs handles the GET /jobs request to fetch job openings
func (jh *JobHandler) GetJobs(c *gin.Context) {
	// Call the service method to fetch job openings
	jobs, err := jh.jobService.GetJobs()
	if err != nil {
		slog.Error("get jobs:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch job openings"})
		fmt.Println()
		return
	}

	c.JSON(http.StatusOK, gin.H{"jobs": jobs})
}

// JobApplyHandler handles the request to apply for a job
func (jh *JobHandler) JobApplyHandler(c *gin.Context) {

	// Extract the authenticated user from the context
	cl, ok := c.Get("claims")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "StatusUnauthorized"})
		return
	}
	claims, ok := cl.(*auth.Claims)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "StatusUnauthorized"})
		return
	}

	// Check if the user has the required role
	if claims.Roles != "Applicant" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Only Applicant users are allowed to apply for jobs"})
		return
	}

	// Parse the job ID from the query parameters
	jobIDStr := c.Param("job_id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		slog.Error("string conversion:", slog.String("Error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		fmt.Println()
		return
	}

	// Apply for the job
	err = jh.jobService.ApplyForJob(claims.Subject, jobID)
	if err != nil {
		slog.Error("apply for job:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to apply for job")})
		fmt.Println()
		return
	}

	// Return success response
	c.JSON(http.StatusAccepted, gin.H{"message": "applied succesfully"})
}

// GetJobAndApplicantsHandler retrieves information about a specific job and its applicants.
func (jh *JobHandler) GetJobAndApplicantsHandler(c *gin.Context) {
	// Extract the job ID from the URL parameters
	jobIDStr := c.Param("job_id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		slog.Error("string conversion:", slog.String("Error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		fmt.Println()
		return
	}

	// Get job and applicants
	job, applicants, err := jh.jobService.GetJobAndApplicants(jobID)
	if err != nil {
		slog.Error("get job and applicants:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to fetch job and applicants: %v", err)})
		fmt.Println()
		return
	}

	// Return job and applicants in the response
	c.JSON(http.StatusOK, gin.H{"job": job, "applicants": applicants})
}
