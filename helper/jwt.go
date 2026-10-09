package helper

import (
	"fmt"
	"time"

	"siakad-mini/app/model"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJWTToken(user *model.User, studentID *int, secret string) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
		"iat":     time.Now().Unix(),
	}

	if studentID != nil {
		claims["student_id"] = *studentID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, nil
}

func ParseJWTToken(tokenString string, secret string) (model.JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return model.JWTClaims{}, fmt.Errorf("token invalid or expired")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return model.JWTClaims{}, fmt.Errorf("invalid token claims")
	}

	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return model.JWTClaims{}, fmt.Errorf("invalid user id in token")
	}

	email, _ := claims["email"].(string)
	role, _ := claims["role"].(string)

	jwtClaims := model.JWTClaims{
		UserID: int(userIDFloat),
		Email:  email,
		Role:   role,
	}

	if sIDVal, exists := claims["student_id"]; exists && sIDVal != nil {
		if sIDFloat, ok := sIDVal.(float64); ok {
			sID := int(sIDFloat)
			jwtClaims.StudentID = &sID
		}
	}

	return jwtClaims, nil
}
