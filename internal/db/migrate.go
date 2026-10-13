package db

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"gorm.io/gorm"
)

func Migrate(ctx context.Context, gdb *gorm.DB, migrations fs.FS) error {
	db := gdb.WithContext(ctx)
	if err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`).Error; err != nil {
		return fmt.Errorf("创建 schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrations, ".")
	if err != nil {
		return fmt.Errorf("读取 migrations: %w", err)
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".sql") {
			continue
		}
		files = append(files, e.Name())
	}
	sort.Strings(files)

	for _, name := range files {
		var exists bool
		if err := db.Raw(
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)`, name,
		).Scan(&exists).Error; err != nil {
			return fmt.Errorf("查询迁移版本 %s: %w", name, err)
		}
		if exists {
			continue
		}

		sqlBytes, err := fs.ReadFile(migrations, name)
		if err != nil {
			return fmt.Errorf("读取 %s: %w", name, err)
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Exec(string(sqlBytes)).Error; err != nil {
				return fmt.Errorf("执行 %s: %w", name, err)
			}
			return tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, name).Error
		})
		if err != nil {
			return err
		}
	}
	return nil
}
