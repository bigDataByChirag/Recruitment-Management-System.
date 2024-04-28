package models

// User represents a user in the system
type User struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	Address      string `json:"address"`
	UserType     string `json:"user_type"`
	PasswordHash string `json:"-"`
	ProfileHead  string `json:"profile_headline"`
	Profile      Profile
}

// Struct to hold the new user data
type NewUser struct {
	Name            string `json:"name"`
	Email           string `json:"email"`
	Address         string `json:"address"`
	UserType        string `json:"user_type"`
	Password        string `json:"password"`
	ProfileHeadline string `json:"profile_headline"`
}

// Profile represents the profile of a user
type Profile struct {
	Applicant      *User          `json:"applicant"`
	ResumeFileAddr string         `json:"resume_file_address"`
	Skills         map[string]any `json:"skills"`
	Education      map[string]any `json:"education"`
	Experience     map[string]any `json:"experience"`
	Phone          string         `json:"phone"`
}

type NewProfile struct {
	ID             int    `json:"id"`
	Applicant      *User  `json:"applicant"`
	ResumeFileAddr string `json:"resume_file_address"`
	Skills         string `json:"skills"`
	Education      string `json:"education"`
	Experience     string `json:"experience"`
	Phone          string `json:"phone"`
}

type UserWithProfile struct {
	ID          int
	Name        string
	Email       string
	Address     string
	UserType    string
	ProfileHead string
	Skills      []string
	Education   []map[string]any
	Experience  []map[string]any
	Phone       string
}
