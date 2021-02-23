package utils

import (
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	viper "github.com/spf13/viper"
	databases "gitlab.com/fibocloud/aws-billing/api_v2/databases"
	bcrypt "golang.org/x/crypto/bcrypt"
)

// access secret key
var accessKey = []byte(viper.GetString("JWT_ACCESS_KEY"))

// refresh secret key
var refreshKey = []byte(viper.GetString("JWT_REFRESH_KEY"))

// Claims ...
type Claims struct {
	Email         string                    `json:"email"`
	IsActive      bool                      `json:"is_active"`
	AwsCredential *databases.AwsCredentials `json:"aws_credential"`
	jwt.StandardClaims
}

// ExtractJWTString Get claim from token string
func ExtractJWTString(tokenString string) (*Claims, error) {
	retClaim := &Claims{}
	JwtToken, err := jwt.ParseWithClaims(tokenString, retClaim, func(t *jwt.Token) (interface{}, error) {
		return []byte(accessKey), nil
	})
	if err == nil {
		if !JwtToken.Valid {
			return retClaim, nil
		}
	}
	return retClaim, err
}

// GenerateToken ...
func GenerateToken(user databases.SystemUser) (string, string) {
	fmt.Println("user", user)
	accessExpTime := time.Now().Add(720 * time.Hour)
	refreshExpTime := time.Now().Add(168 * time.Hour)
	accessToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		Email:         user.Email,
		IsActive:      user.IsActive,
		AwsCredential: user.AwsCredentials,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: accessExpTime.Unix(),
		},
	}).SignedString(accessKey)
	refreshToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, &Claims{
		Email:         user.Email,
		IsActive:      user.IsActive,
		AwsCredential: user.AwsCredentials,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: refreshExpTime.Unix(),
		},
	}).SignedString(accessKey)
	return accessToken, refreshToken
}

// GenerateHash password hash generate
func GenerateHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// ComparePassword compare password and hash
func ComparePassword(password, hash string) (bool, error) {
	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(hash)); err != nil {
		return false, err
	}
	return true, nil
}
