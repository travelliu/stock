// Package bootstrap performs one-time tasks at server startup.
package bootstrap

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"stock/pkg/stockd/auth"
	"stock/pkg/stockd/models"
)

// EnsureAdmin seeds admin/<random-password> when the users table is empty.
// On seed, the plain password is returned AND logged at WARN level so the
// operator can capture it. Returns "" if no seeding occurred.
func EnsureAdmin(gdb *gorm.DB, logger *logrus.Logger) (string, error) {
	var n int64
	if err := gdb.Model(&models.User{}).Count(&n).Error; err != nil {
		return "", fmt.Errorf("count users: %w", err)
	}
	if n > 0 {
		return "", nil
	}
	plain, err := generatePassword(24)
	if err != nil {
		return "", err
	}
	hash, err := auth.HashPassword(plain)
	if err != nil {
		return "", err
	}
	admin := &models.User{Username: "admin", PasswordHash: hash, Role: "admin"}
	if err := gdb.Create(admin).Error; err != nil {
		return "", fmt.Errorf("create admin: %w", err)
	}
	logger.WithFields(logrus.Fields{
		"username": "admin",
		"password": plain,
	}).Warn("seeded initial admin user — change this password immediately")
	return plain, nil
}

func generatePassword(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf)[:n], nil
}
