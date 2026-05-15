package analysis

import (
	"context"
	"stock/pkg/stockd/config"

	"gorm.io/gorm"

	"stock/pkg/models"
)

type Service struct {
	db  *gorm.DB
	cfg *config.Config
}

func New(db *gorm.DB, cfg *config.Config) *Service { return &Service{db: db, cfg: cfg} }

type Input struct {
	UserID      uint
	TsCode      string
	OpenPrice   *float64
	ActualHigh  *float64
	ActualLow   *float64
	ActualClose *float64
}

func (s *Service) Run(ctx context.Context, in Input) (*models.AnalysisResult, error) {
	var bars []*models.DailyBar
	err := s.db.WithContext(ctx).Where("ts_code = ?", in.TsCode).Order("trade_date ASC").Find(&bars).Error
	if err != nil {
		return &models.AnalysisResult{}, err
	}

	var name string
	var st models.Stock
	if s.db.WithContext(ctx).First(&st, "ts_code = ? or code = ?", in.TsCode, in.TsCode).Error == nil {
		name = st.Name
	}

	res := Build(models.Input{
		TsCode: in.TsCode, StockName: name,
		Rows:        bars,
		OpenPrice:   in.OpenPrice,
		ActualHigh:  in.ActualHigh,
		ActualLow:   in.ActualLow,
		ActualClose: in.ActualClose,
	})
	return res, nil
}
