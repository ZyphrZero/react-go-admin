package seed_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"react-go-admin/app/internal/config"
	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/migrate"
	moduleregistry "react-go-admin/app/internal/modules"
	"react-go-admin/app/internal/platform/database"
	"react-go-admin/app/internal/seed"
)

func TestRunBootstrapsBaselineData(t *testing.T) {
	moduleregistry.Register()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	ctx := context.Background()
	if err := migrate.Up(ctx, db); err != nil {
		t.Fatalf("migrate up: %v", err)
	}

	cfg := &config.Config{
		InitialAdminUsername:     "admin",
		InitialAdminEmail:        "admin@example.com",
		InitialAdminNickname:     "Administrator",
		InitialAdminPassword:     "Admin@12345",
		PasswordMinLength:        8,
		PasswordRequireUppercase: true,
		PasswordRequireLowercase: true,
		PasswordRequireDigits:    true,
		PasswordRequireSpecial:   true,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	container := framework.NewContainer(cfg, logger, db)
	if err := framework.BuildModules(container); err != nil {
		t.Fatalf("build modules: %v", err)
	}

	if err := seed.Run(ctx, container); err != nil {
		t.Fatalf("seed run: %v", err)
	}
	if err := seed.Run(ctx, container); err != nil {
		t.Fatalf("seed rerun: %v", err)
	}

	var roles []database.Role
	if err := db.Order("id ASC").Find(&roles).Error; err != nil {
		t.Fatalf("query roles: %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %d", len(roles))
	}

	var admin database.User
	if err := db.Where("username = ?", cfg.InitialAdminUsername).First(&admin).Error; err != nil {
		t.Fatalf("query admin: %v", err)
	}
	if !admin.IsSuperuser {
		t.Fatalf("expected admin to be superuser")
	}

	var adminAssignments int64
	if err := db.Model(&database.UserRole{}).Where("user_id = ?", admin.ID).Count(&adminAssignments).Error; err != nil {
		t.Fatalf("count user roles: %v", err)
	}
	if adminAssignments != 1 {
		t.Fatalf("expected 1 admin role assignment, got %d", adminAssignments)
	}

	var apiCount int64
	if err := db.Model(&database.APIRecord{}).Count(&apiCount).Error; err != nil {
		t.Fatalf("count apis: %v", err)
	}
	if apiCount == 0 {
		t.Fatalf("expected api catalog to be populated")
	}

	var adminRole database.Role
	if err := db.Where("name = ?", "管理员").First(&adminRole).Error; err != nil {
		t.Fatalf("query admin role: %v", err)
	}
	if len(adminRole.APIIDs) == 0 {
		t.Fatalf("expected admin role API permissions to be populated")
	}
}
