package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"log/slog"
	"net/http"
	"rms/internals/auth"
	"rms/internals/services"
	"strings"
)

type Resume struct {
	resumeService *services.ResumeService
	a             *auth.Auth
}

// NewJob creates a new Job handler with the provided services and authentication
func NewResume(rs *services.ResumeService, a *auth.Auth) (*Resume, error) {
	if rs == nil || a == nil {
		return nil, errors.New("please provide all the values")
	}
	return &Resume{
		resumeService: rs,
		a:             a,
	}, nil
}

func (r Resume) UploadResume(c *gin.Context) {
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

	// Get the file from the form with key "resume"
	file, fileHeader, err := c.Request.FormFile("resume")
	if err != nil {
		slog.Error("form file:", slog.String("Error", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to parse form data"})
		fmt.Println()
		return
	}
	defer file.Close()

	// Extract file extension
	fileName := fileHeader.Filename
	ext := strings.ToLower(fileName[strings.LastIndex(fileName, ".")+1:])
	if ext != "pdf" && ext != "docx" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported file format. Only PDF and DOCX files are allowed."})
		return
	}

	// Read the file content
	fileContent, err := io.ReadAll(file)
	if err != nil {
		slog.Error("ReadAll:", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		fmt.Println()
		return
	}

	// Call the third-party API to upload resume and get parsed data
	parsedData, err := callResumeParserAPI(fileContent)
	if err != nil {
		slog.Error("call resume parser", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse resume data"})
		fmt.Println()
		return
	}

	// Store parsed data into the database
	err = r.resumeService.SaveResume(claims.Subject, parsedData, fileName)
	if err != nil {
		slog.Error("save resume", slog.String("Error", err.Error()))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save resume data"})
		fmt.Println()
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Resume data saved successfully"})
}

func callResumeParserAPI(data []byte) (map[string]interface{}, error) {
	// API Endpoint
	url := "https://api.apilayer.com/resume_parser/upload"
	// API Key
	apiKey := "dHIISlwHwDqbAlJxtXjLVthJXfvzs0C1"

	// Create request body
	reqBody := bytes.NewBuffer(data)

	// Create HTTP request
	req, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		return nil, err
	}

	// Set request headers
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("apikey", apiKey)

	// Send HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Parse response JSON
	var parsedData map[string]interface{}
	err = json.Unmarshal(respBody, &parsedData)
	if err != nil {
		return nil, err
	}

	return parsedData, nil
}
