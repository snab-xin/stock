package repo

import (
	"context"
	"strings"
	"time"

	"github.com/stock/ashare/internal/domain"
	"gorm.io/gorm"
)

type NoteRepo struct {
	db *gorm.DB
}

func NewNoteRepo(db *gorm.DB) *NoteRepo {
	return &NoteRepo{db: db}
}

func (r *NoteRepo) List(ctx context.Context, userID int64, tsCode, targetType string, limit int) ([]domain.UserNote, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	q := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if tsCode = strings.TrimSpace(tsCode); tsCode != "" {
		q = q.Where("ts_code = ?", strings.ToUpper(tsCode))
	}
	if targetType = strings.TrimSpace(targetType); targetType != "" {
		q = q.Where("target_type = ?", targetType)
	}
	var items []domain.UserNote
	err := q.Select(noteColumns).Order("trade_date DESC, id DESC").Limit(limit).Find(&items).Error
	return items, err
}

// Marks 某标的在日线上的笔记标记，不含正文。
func (r *NoteRepo) Marks(ctx context.Context, userID int64, tsCode string) ([]domain.UserNote, error) {
	tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
	if tsCode == "" {
		return nil, nil
	}
	var items []domain.UserNote
	err := r.db.WithContext(ctx).
		Select("id", "trade_date", "summary", "ts_code", "target_type").
		Where("user_id = ? AND ts_code = ?", userID, tsCode).
		Order("trade_date, id").
		Find(&items).Error
	return items, err
}

func (r *NoteRepo) Get(ctx context.Context, userID, id int64) (domain.UserNote, error) {
	var items []domain.UserNote
	err := r.db.WithContext(ctx).Select(noteColumns).Where("user_id = ? AND id = ?", userID, id).Limit(1).Find(&items).Error
	if err != nil {
		return domain.UserNote{}, err
	}
	if len(items) == 0 {
		return domain.UserNote{}, gorm.ErrRecordNotFound
	}
	return items[0], nil
}

func (r *NoteRepo) Create(ctx context.Context, n *domain.UserNote) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *NoteRepo) Update(ctx context.Context, n *domain.UserNote) error {
	return r.db.WithContext(ctx).Model(&domain.UserNote{}).
		Where("id = ? AND user_id = ?", n.ID, n.UserID).
		Updates(map[string]any{
			"ts_code":     n.TSCode,
			"target_type": n.TargetType,
			"trade_date":  n.TradeDate,
			"summary":     n.Summary,
			"content":     n.Content,
		}).Error
}

func (r *NoteRepo) Delete(ctx context.Context, userID, id int64) error {
	res := r.db.WithContext(ctx).Where("user_id = ? AND id = ?", userID, id).Delete(&domain.UserNote{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// NormalizeNoteTarget 校验可选标的。无代码时类型必须为空；有代码时类型为 stock/cb/index。
func NormalizeNoteTarget(tsCode, targetType string) (string, string, error) {
	tsCode = strings.ToUpper(strings.TrimSpace(tsCode))
	targetType = strings.ToLower(strings.TrimSpace(targetType))
	if tsCode == "" {
		if targetType != "" {
			return "", "", errTargetNeedsCode
		}
		return "", "", nil
	}
	switch targetType {
	case "stock", "cb", "index":
		return tsCode, targetType, nil
	default:
		return "", "", errBadTargetType
	}
}

var (
	errTargetNeedsCode = errString("填写了类型但没有代码")
	errBadTargetType   = errString("类型须为 stock、cb 或 index")
)

type errString string

func (e errString) Error() string { return string(e) }

// TouchTradeDate 保证日期非零。
func TouchTradeDate(t time.Time) error {
	if t.IsZero() {
		return errString("交易日必填")
	}
	return nil
}
