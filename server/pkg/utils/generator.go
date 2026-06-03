package utils

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

func NewUUID() string {
	return uuid.New().String()
}

 
func GenerateUsername(name string) string {
	base := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	var clean []rune
	for _, r := range base {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			clean = append(clean, r)
		}
	}
	suffix := rand.New(rand.NewSource(time.Now().UnixNano())).Intn(9000) + 1000
	return fmt.Sprintf("%s_%d", string(clean), suffix)
}

 
func GenerateAvatarURL(username string) string {
	return fmt.Sprintf("https://api.dicebear.com/7.x/initials/svg?seed=%s", username)
}

 
func GenerateRandomString(n int) string {
	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = chars[r.Intn(len(chars))]
	}
	return string(b)
}