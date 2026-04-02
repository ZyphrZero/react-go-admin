package migrate

import (
	"context"

	"gorm.io/gorm"

	"react-go-admin/app/internal/framework"
)

func Up(ctx context.Context, db *gorm.DB) error {
	models := framework.RegisteredModels()
	if len(models) == 0 {
		return nil
	}
	return db.WithContext(ctx).AutoMigrate(models...)
}

func NeedsMigration(ctx context.Context, db *gorm.DB) (bool, error) {
	for _, model := range framework.RegisteredModels() {
		if !db.WithContext(ctx).Migrator().HasTable(model) {
			return true, nil
		}
	}
	return false, nil
}
