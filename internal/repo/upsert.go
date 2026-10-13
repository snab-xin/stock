package repo

import (
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const upsertBatchSize = 500

func upsert(ctx context.Context, db *gorm.DB, slicePtr any, conflictCols []string, updateCols []string) error {
	v := reflect.ValueOf(slicePtr)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("upsert 需要切片指针")
	}
	if v.Elem().Len() == 0 {
		return nil
	}

	cols := make([]clause.Column, len(conflictCols))
	for i, name := range conflictCols {
		cols[i] = clause.Column{Name: name}
	}
	updates := append(append([]string{}, updateCols...), "updated_at")

	return db.WithContext(ctx).
		Session(&gorm.Session{SkipDefaultTransaction: true}).
		Clauses(clause.OnConflict{
			Columns:   cols,
			DoUpdates: clause.AssignmentColumns(updates),
		}).
		CreateInBatches(slicePtr, upsertBatchSize).Error
}
