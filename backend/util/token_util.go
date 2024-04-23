package util

import (
	"chulbong-kr/config"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	mrand "math/rand/v2"
	"time"

	"github.com/gofiber/fiber/v2"
)

var LOGIN_TOKEN_COOKIE string

// GenerateOpaqueToken creates a random token
func GenerateOpaqueToken() (string, error) {
	bytes := make([]byte, 16) // Generates a 128-bit token
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GenerateRandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	s := make([]rune, n)
	for i := range s {
		s[i] = rune(letters[mrand.IntN(len(letters))])
	}
	return string(s)
}

func GenerateState() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func GenerateLoginCookie(value string) fiber.Cookie {
	cookie := fiber.Cookie{
		Name:     LOGIN_TOKEN_COOKIE,
		Value:    value,                              // The token generated for the user
		Expires:  time.Now().Add(24 * 7 * time.Hour), // Set the cookie to expire in 7 days
		HTTPOnly: true,                               // Ensure the cookie is not accessible through client-side scripts
		Secure:   true,                               // Ensure the cookie is sent over HTTPS
		SameSite: "Lax",                              // Lax, None, or Strict. Lax is a reasonable default
		Path:     "/",                                // Scope of the cookie
	}

	if config.IS_PRODUCTION == "production" {
		cookie.Domain = ".k-pullup.com" // Allow cookie to be shared across all subdomains
	}
	return cookie
}

func ClearLoginCookie() fiber.Cookie {
	cookie := fiber.Cookie{
		Name:     LOGIN_TOKEN_COOKIE,
		Value:    "",                         // The token generated for the user
		Expires:  time.Now().Add(-time.Hour), // Set the cookie to be expired
		HTTPOnly: true,                       // Ensure the cookie is not accessible through client-side scripts
		Secure:   true,                       // Ensure the cookie is sent over HTTPS
		SameSite: "Lax",                      // Lax, None, or Strict. Lax is a reasonable default
		Path:     "/",                        // Scope of the cookie
	}

	if config.IS_PRODUCTION == "production" {
		cookie.Domain = ".k-pullup.com" // Allow cookie to be shared across all subdomains
	}
	return cookie
}