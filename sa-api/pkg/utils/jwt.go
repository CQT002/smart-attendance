package utils

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hdbank/smart-attendance/internal/domain/entity"
)

// Claims dữ liệu được mã hóa trong JWT token
type Claims struct {
	UserID   uint             `json:"user_id"`
	BranchID *uint            `json:"branch_id"`
	Role     entity.UserRole  `json:"role"`
	Email    string           `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken tạo JWT access token
func GenerateToken(user *entity.User, secret string, expireHours int) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		BranchID: user.BranchID,
		Role:     user.Role,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(expireHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "smart-attendance",
			Subject:   fmt.Sprintf("%d", user.ID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// GenerateRefreshToken tạo refresh token (ít thông tin hơn, sống lâu hơn)
func GenerateRefreshToken(userID uint, secret string, expireDays int) (string, error) {
	claims := jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().AddDate(0, 0, expireDays)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		Issuer:    "smart-attendance",
		Subject:   fmt.Sprintf("%d", userID),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// ParseToken parse và validate JWT token
func ParseToken(tokenStr, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}
