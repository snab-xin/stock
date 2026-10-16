package domain

import "time"

// User 登录用户。密码只存 bcrypt 哈希。
type User struct {
	Audit
	Username     string `gorm:"uniqueIndex;size:64"`
	PasswordHash string `gorm:"column:password_hash;size:256"`
	Email        string `gorm:"size:128"`
	Phone        string `gorm:"size:16"`
	Role         string `gorm:"size:16"`
	Status       string `gorm:"size:16"` // active / locked
}

func (User) TableName() string { return "users" }

// UserNote 用户笔记。正文是 Markdown。
// TradeDate 必填；TSCode 可空，不建外键。有代码时 TargetType 为 stock / cb / index。
type UserNote struct {
	Audit
	UserID     int64     `gorm:"index"`
	TSCode     string    `gorm:"column:ts_code;size:16"`
	TargetType string    `gorm:"column:target_type;size:16"`
	TradeDate  time.Time `gorm:"column:trade_date;index"`
	Summary    string    `gorm:"size:256"`
	Content    string    `gorm:"type:text"`
}

func (UserNote) TableName() string { return "user_notes" }

// UserSession 登录会话。只存 token 的 SHA-256。
type UserSession struct {
	ID        int64     `gorm:"primaryKey"`
	UserID    int64     `gorm:"index"`
	TokenHash string    `gorm:"column:token_hash;size:64;uniqueIndex"`
	ExpiresAt time.Time `gorm:"column:expires_at"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (UserSession) TableName() string { return "user_sessions" }
