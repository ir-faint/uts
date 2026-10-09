package helper

import (
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 10

func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func VerifyPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

// In-Memory Login Rate Limiter (5 attempts per minute per IP)
type attemptRecord struct {
	count     int
	resetTime time.Time
}

type loginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string]*attemptRecord
}

var limiter = &loginRateLimiter{
	attempts: make(map[string]*attemptRecord),
}

func IsLoginRateLimited(ip string) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	rec, exists := limiter.attempts[ip]
	if !exists {
		return false
	}
	if time.Now().After(rec.resetTime) {
		delete(limiter.attempts, ip)
		return false
	}
	return rec.count >= 5
}

func RecordFailedLoginAttempt(ip string) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := time.Now()
	rec, exists := limiter.attempts[ip]
	if !exists || now.After(rec.resetTime) {
		limiter.attempts[ip] = &attemptRecord{
			count:     1,
			resetTime: now.Add(1 * time.Minute),
		}
	} else {
		rec.count++
	}
}

func ResetFailedLoginAttempts(ip string) {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()
	delete(limiter.attempts, ip)
}
