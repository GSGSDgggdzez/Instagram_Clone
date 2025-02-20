package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var secretKey = []byte("your-secret-key-here") // In production, use environment variable

type Claims struct {
	UserID         uint   `json:"user_ID"`
	Email          string `json:"email"`
	Name           string `json:"name"`
	Avatar         string `json:"avatar"`
	JWTToken       string `json:"token"`
	Bio            string `json:"bio"`
	IsVerified     bool   `json:"is_verified"`
	FollowerCount  int    `json:"follower_count"`
	FollowingCount int    `json:"following_count"`
	Username       string `json:"username"`
	Phone          string `json:"phone"`
	Website        string `json:"website"`
	Privacy        bool   `json:"privacy"`
	EmailVerified  bool   `json:"email_verified"`
	Language       string `json:"language"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, email, name, username, avatar, token, bio, website, phone, language string, is_verified, privacy, email_verified bool, following_count, follower_count int) (string, error) {
	claims := Claims{
		UserID:         userID,
		Email:          email,
		Name:           name,
		Username:       username,
		Avatar:         avatar,
		JWTToken:       token,
		Bio:            bio,
		Website:        website,
		Phone:          phone,
		FollowerCount:  follower_count,
		FollowingCount: following_count,
		IsVerified:     is_verified,
		Privacy:        privacy,
		EmailVerified:  email_verified,
		Language:       language,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return jwtToken.SignedString(secretKey)
}
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func ExtractTokenFromHeader(header string) (*Claims, *fiber.Map) {
	if header == "" {
		return nil, &fiber.Map{
			"status": fiber.StatusUnauthorized,
			"error":  "Authorization header is required",
		}
	}

	parts := strings.Split(header, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		return nil, &fiber.Map{
			"status": fiber.StatusUnauthorized,
			"error":  "Invalid authorization header format",
		}
	}

	tokenString := parts[1]

	claims, err := ValidateToken(tokenString)
	if err != nil {
		return nil, &fiber.Map{
			"status": fiber.StatusUnauthorized,
			"error":  fmt.Sprintf("Invalid token: %v", err),
		}
	}

	return claims, nil
}
