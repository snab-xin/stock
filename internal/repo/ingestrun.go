package repo

import (
	"context"
	"time"

	"github.com/stock/ashare/internal/domain"
	"gorm.io/gorm"
)

type IngestRunRepo struct {
	db *gorm.DB
}

func NewIngestRunRepo(db *gorm.DB) *IngestRunRepo {
	return &IngestRunRepo{db: db}
}

func (r *IngestRunRepo) Start(ctx context.Context, jobName string, params domain.JSONObject) (*domain.IngestRun, error) {
	run := &domain.IngestRun{
		JobName:   jobName,
		StartedAt: time.Now(),
		Status:    "running",
		Params:    params,
	}
	if err := r.db.WithContext(ctx).Create(run).Error; err != nil {
		return nil, err
	}
	return run, nil
}

func (r *IngestRunRepo) ListRecent(ctx context.Context, limit int) ([]domain.IngestRun, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var rows []domain.IngestRun
	err := r.db.WithContext(ctx).Select(ingestRunColumns).Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *IngestRunRepo) Finish(ctx context.Context, run *domain.IngestRun, status, message string, rows int64) error {
	now := time.Now()
	run.Status = status
	run.Message = message
	run.RowsAffected = rows
	run.FinishedAt = &now
	return r.db.WithContext(ctx).Save(run).Error
}
