package repo

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/stock/ashare/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Count(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.User{}).Count(&n).Error
	return n, err
}

func (r *UserRepo) Create(ctx context.Context, username, password string) (domain.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}
	u := domain.User{
		Username:     strings.TrimSpace(username),
		PasswordHash: string(hash),
		Role:         "user",
		Status:       "active",
	}
	if err := r.db.WithContext(ctx).Create(&u).Error; err != nil {
		return domain.User{}, err
	}
	return u, nil
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (domain.User, error) {
	var items []domain.User
	err := r.db.WithContext(ctx).Where("username = ?", strings.TrimSpace(username)).Limit(1).Find(&items).Error
	if err != nil {
		return domain.User{}, err
	}
	if len(items) == 0 {
		return domain.User{}, gorm.ErrRecordNotFound
	}
	return items[0], nil
}

func (r *UserRepo) GetByID(ctx context.Context, id int64) (domain.User, error) {
	var items []domain.User
	err := r.db.WithContext(ctx).Where("id = ?", id).Limit(1).Find(&items).Error
	if err != nil {
		return domain.User{}, err
	}
	if len(items) == 0 {
		return domain.User{}, gorm.ErrRecordNotFound
	}
	return items[0], nil
}

func CheckPassword(u domain.User, password string) bool {
	if u.Status == "locked" {
		return false
	}
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (r *UserRepo) CreateSession(ctx context.Context, userID int64, token string, ttl time.Duration) error {
	s := domain.UserSession{
		UserID:    userID,
		TokenHash: HashToken(token),
		ExpiresAt: time.Now().Add(ttl),
	}
	return r.db.WithContext(ctx).Create(&s).Error
}

func (r *UserRepo) UserByToken(ctx context.Context, token string) (domain.User, error) {
	var sessions []domain.UserSession
	err := r.db.WithContext(ctx).
		Where("token_hash = ? AND expires_at > ?", HashToken(token), time.Now()).
		Limit(1).
		Find(&sessions).Error
	if err != nil {
		return domain.User{}, err
	}
	if len(sessions) == 0 {
		return domain.User{}, gorm.ErrRecordNotFound
	}
	return r.GetByID(ctx, sessions[0].UserID)
}

func (r *UserRepo) DeleteSession(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", HashToken(token)).Delete(&domain.UserSession{}).Error
}
