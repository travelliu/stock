// Package services issues and revokes API tokens.
package services

import (
	"context"
	"time"

	"stock/pkg/models"
	"stock/pkg/stockd/auth"
)

type IssueInput struct {
	UserID    uint       `json:"userId"`
	Name      string     `json:"name,omitempty"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

type Token struct {
	ID         uint       `json:"id"`
	UserID     uint       `json:"userId"`
	Name       string     `json:"name,omitempty"`
	LastUsedAt *time.Time `json:"lastUsedAt"`
	ExpiresAt  *time.Time `json:"expiresAt"`
	CreatedAt  time.Time  `json:"createdAt"`
	PlainOnce  string     `json:"plainOnce,omitempty"` // set only by Issue's return; never persisted/returned from List
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

func (s *Service) ListTokens(ctx context.Context, userID uint) ([]Token, error) {
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

func (s *Service) RevokeToken(ctx context.Context, userID, id uint) error {
	return s.db.WithContext(ctx).Where("id = ? AND user_id = ?", id, userID).Delete(&models.APIToken{}).Error
}
