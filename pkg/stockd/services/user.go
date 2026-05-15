// Package services implements user-management business logic.
package services

import (
	"context"
	"errors"
	"fmt"

	"stock/pkg/models"
	"stock/pkg/stockd/auth"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrDisabled     = errors.New("account disabled")
	ErrInvalidCred  = errors.New("invalid credentials")
)

// DTO returned to callers. Never carries the password hash.
type User struct {
	ID           uint   `json:"ID,omitempty"`
	Username     string `json:"username,omitempty"`
	Role         string `json:"role,omitempty"`
	Disabled     bool   `json:"disabled,omitempty"`
	TushareToken string `json:"tushareToken,omitempty"`
}

type CreateUserInput struct {
	Username     string `json:"username,omitempty"`
	Password     string `json:"password,omitempty"`
	Role         string `json:"role,omitempty"` // "user" | "admin"
	TushareToken string `json:"tushareToken,omitempty"`
}

func toDTO(u *models.User) User {
	return User{
		ID: u.ID, Username: u.Username, Role: u.Role,
		Disabled: u.Disabled, TushareToken: u.TushareToken,
	}
}

func (s *Service) CreateUser(ctx context.Context, in CreateUserInput) (*User, error) {
	if in.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if in.Role != "user" && in.Role != "admin" {
		return nil, fmt.Errorf("role must be user|admin")
	}
	hash, err := auth.HashPassword(in.Password)
	if err != nil {
		return nil, err
	}
	u := &models.User{
		Username: in.Username, PasswordHash: hash,
		Role: in.Role, TushareToken: in.TushareToken,
	}
	if err := s.db.WithContext(ctx).Create(u).Error; err != nil {
		return nil, err
	}
	dto := toDTO(u)
	return &dto, nil
}

func (s *Service) Authenticate(ctx context.Context, username, password string) (*User, error) {
	var u models.User
	if err := s.db.WithContext(ctx).Where("username = ?", username).First(&u).Error; err != nil {
		return nil, ErrInvalidCred
	}
	if u.Disabled {
		return nil, ErrDisabled
	}
	if err := auth.CheckPassword(u.PasswordHash, password); err != nil {
		return nil, ErrInvalidCred
	}
	dto := toDTO(&u)
	return &dto, nil
}

func (s *Service) ChangePassword(ctx context.Context, id uint, old, new string) error {
	var u models.User
	if err := s.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return ErrUserNotFound
	}
	if err := auth.CheckPassword(u.PasswordHash, old); err != nil {
		return ErrInvalidCred
	}
	hash, err := auth.HashPassword(new)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&u).Update("password_hash", hash).Error
}

func (s *Service) SetUserDisabled(ctx context.Context, id uint, disabled bool) error {
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("disabled", disabled).Error
}

func (s *Service) SetUserTushareToken(ctx context.Context, id uint, token string) error {
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("tushare_token", token).Error
}

func (s *Service) GetUser(ctx context.Context, id uint) (*User, error) {
	var u models.User
	if err := s.db.WithContext(ctx).First(&u, id).Error; err != nil {
		return nil, ErrUserNotFound
	}
	dto := toDTO(&u)
	return &dto, nil
}

func (s *Service) ListUser(ctx context.Context) ([]User, error) {
	var rows []models.User
	if err := s.db.WithContext(ctx).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]User, len(rows))
	for i := range rows {
		out[i] = toDTO(&rows[i])
	}
	return out, nil
}

func (s *Service) ResetUserPassword(ctx context.Context, id uint, new string) error {
	hash, err := auth.HashPassword(new)
	if err != nil {
		return err
	}
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("password_hash", hash).Error
}

func (s *Service) DeleteUser(ctx context.Context, id uint) error {
	return s.db.WithContext(ctx).Delete(&models.User{}, id).Error
}

func (s *Service) SetUserRole(ctx context.Context, id uint, role string) error {
	if role != "user" && role != "admin" {
		return fmt.Errorf("role must be user|admin")
	}
	return s.db.WithContext(ctx).Model(&models.User{}).Where("id = ?", id).Update("role", role).Error
}
