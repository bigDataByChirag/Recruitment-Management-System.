package auth

import (
	"crypto/rsa"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

const (
	Admin     = "admin"
	Applicant = "applicant"
)

// Auth struct represents the authentication module with private and public keys
type Auth struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewAuth creates a new Auth instance with provided public and private keys
func NewAuth(pubKey *rsa.PublicKey, privateKey *rsa.PrivateKey) (*Auth, error) {
	if pubKey == nil || privateKey == nil {
		return nil, errors.New("private key, public key cannot be nil")
	}
	return &Auth{privateKey: privateKey, publicKey: pubKey}, nil
}

// Claims struct represents JWT claims with additional 'Roles' field
type Claims struct {
	jwt.RegisteredClaims
	Roles string `json:"roles"`
}

// GenerateToken generates a JWT token for a given user ID and role
func (a *Auth) GenerateToken(id int, role string) (string, error) {
	c := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "api project",
			Subject:   strconv.Itoa(id),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(50 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		Roles: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, c)

	encodedToken, err := token.SignedString(a.privateKey)
	if err != nil {
		return "", err
	}
	return encodedToken, nil
}

// VerifyToken verifies a JWT token and checks if the user has the required role
func (a *Auth) VerifyToken(tokenString string, requiredRole string) (*Claims, error) {
	var c Claims

	// Key function for parsing and verifying the token
	k := func(*jwt.Token) (interface{}, error) {
		return a.publicKey, nil
	}

	// Parse the token with custom claims and key function
	token, err := jwt.ParseWithClaims(tokenString, &c, k)
	if err != nil {
		// If error while parsing the token, return the error
		return nil, err
	}

	// Check if the token is valid
	if !token.Valid {
		// If the token is not valid, return an error
		return nil, err
	}

	// If the required role is 'User', check if the user has the required role
	if c.Roles == Applicant {
		if c.Roles != requiredRole {
			return nil, errors.New("you are not authorized to perform this")
		}
	}

	// Return the parsed claims
	return &c, nil
}
