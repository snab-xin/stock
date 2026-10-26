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

// 开始执行
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

// 获取最近的数据
func (r *IngestRunRepo) ListRecent(ctx context.Context, limit int) ([]domain.IngestRun, error) {
	if limit <= 0 || limit > 100 {
		limit = 30
	}
	var rows []domain.IngestRun
	err := r.db.WithContext(ctx).Select(ingestRunColumns).Order("id DESC").Limit(limit).Find(&rows).Error
	return rows, err
}

// 完成
func (r *IngestRunRepo) Finish(ctx context.Context, run *domain.IngestRun, status, message string, rows int64) error {
	now := time.Now()
	run.Status = status
	run.Message = message
	run.RowsAffected = rows
	run.FinishedAt = &now
	return r.db.WithContext(ctx).Save(run).Error
}

// 获取最新的一条数据
func (r *IngestRunRepo) Latest(ctx context.Context, jobName string) (*domain.IngestRun, error) {
	var run domain.IngestRun
	err := r.db.WithContext(ctx).Select(ingestRunColumns).Where("job_name = ? AND status = ?", jobName, "finished").Order("id DESC").First(&run).Error
	if err != nil {
		return nil, err
	}
	return &run, nil
}
