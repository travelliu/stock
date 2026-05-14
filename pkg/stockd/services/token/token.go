// Package token issues and revokes API tokens.
package token

import (
	"context"
	"time"

	"gorm.io/gorm"

	"stock/pkg/stockd/auth"
	"stock/pkg/models"
)

type Service struct{ db *gorm.DB }

func New(db *gorm.DB) *Service { return &Service{db: db} }

type IssueInput struct {
	UserID    uint
	Name      string
	ExpiresAt *time.Time
}

type Token struct {
	ID         uint
	UserID     uint
	Name       string
	LastUsedAt *time.Time
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	PlainOnce  string // set only by Issue's return; never persisted/returned from List
}

// Issue creates a new API token and returns (plain, dto). The plain text is
// shown to the user once; subsequent reads only return the metadata.
func (s *Service) Issue(ctx context.Context, in IssueInput) (string, *Token, error) {
	plain, hash, err := auth.GenerateAPIToken()
	if err != nil {
		return "", nil, err
	}
	row := &models.APIToken{
		UserID: in.UserID, Name: in.Name, TokenHash: hash, ExpiresAt: in.ExpiresAt,
	}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return "", nil, err
	}
	return plain, &Token{
		ID: row.ID, UserID: row.UserID, Name: row.Name,
		LastUsedAt: row.LastUsedAt, ExpiresAt: row.ExpiresAt, CreatedAt: row.CreatedAt,
	}, nil
}

func (s *Service) List(ctx context.Context, userID uint) ([]Token, error) {
	var rows []models.APIToken
	if err := s.db.WithContext(ctx).Where("user_id = ?", userID).Order("id ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]Token, len(rows))
	for i, r := range rows {
		out[i] = Token{ID: r.ID, UserID: r.UserID, Name: r.Name, LastUsedAt: r.LastUsedAt, ExpiresAt: r.ExpiresAt, CreatedAt: r.CreatedAt}
	}
	return out, nil
}

func (s *Service) Revoke(ctx context.Context, userID, id uint) error {
	return s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.APIToken{}).Error
}
