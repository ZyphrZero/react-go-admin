package base

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"react-go-admin/app/internal/config"
	"react-go-admin/app/internal/platform/database"
)

func TestUpdateProfileRemovesOldAvatarFile(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:base-avatar-remove?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.SystemSetting{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}

	baseDir := t.TempDir()
	oldAvatarURL := writeAvatarFixture(t, baseDir, "old-remove.webp")
	user := &database.User{
		Username:       "admin",
		Password:       "hashed",
		IsActive:       true,
		IsSuperuser:    true,
		SessionVersion: 0,
		Avatar:         stringPtr(oldAvatarURL),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := NewService(&config.Config{
		BaseDir:      baseDir,
		DBConnection: "sqlite",
	}, db)

	if err := service.UpdateProfile(context.Background(), user, updateProfileRequest{
		AvatarMode: "remove",
	}, nil); err != nil {
		t.Fatalf("update profile remove avatar: %v", err)
	}

	var refreshed database.User
	if err := db.First(&refreshed, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if refreshed.Avatar != nil {
		t.Fatalf("expected avatar to be nil after removal, got %v", *refreshed.Avatar)
	}
	if _, err := os.Stat(localAvatarPath(baseDir, oldAvatarURL)); !os.IsNotExist(err) {
		t.Fatalf("expected old avatar file to be deleted")
	}
}

func TestUpdateProfileReplacesAvatarAndDeletesPreviousFile(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:base-avatar-replace?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&database.User{}, &database.SystemSetting{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}

	baseDir := t.TempDir()
	oldAvatarURL := writeAvatarFixture(t, baseDir, "old-replace.webp")
	user := &database.User{
		Username:       "admin",
		Password:       "hashed",
		IsActive:       true,
		IsSuperuser:    true,
		SessionVersion: 0,
		Avatar:         stringPtr(oldAvatarURL),
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := NewService(&config.Config{
		BaseDir:      baseDir,
		DBConnection: "sqlite",
	}, db)
	avatarFile := newMultipartFileHeader(t, "avatar.webp", []byte("new-avatar"))

	if err := service.UpdateProfile(context.Background(), user, updateProfileRequest{
		AvatarMode: "replace",
	}, avatarFile); err != nil {
		t.Fatalf("update profile replace avatar: %v", err)
	}

	var refreshed database.User
	if err := db.First(&refreshed, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if refreshed.Avatar == nil || strings.TrimSpace(*refreshed.Avatar) == "" {
		t.Fatalf("expected avatar url after replacement")
	}
	if *refreshed.Avatar == oldAvatarURL {
		t.Fatalf("expected avatar url to change after replacement")
	}
	if _, err := os.Stat(localAvatarPath(baseDir, oldAvatarURL)); !os.IsNotExist(err) {
		t.Fatalf("expected previous avatar file to be deleted")
	}
	if _, err := os.Stat(localAvatarPath(baseDir, *refreshed.Avatar)); err != nil {
		t.Fatalf("expected new avatar file to exist: %v", err)
	}
}

func writeAvatarFixture(t *testing.T, baseDir string, filename string) string {
	t.Helper()

	dir := filepath.Join(baseDir, "storage", "uploads", "avatar", "20260402")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir avatar fixture: %v", err)
	}
	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, []byte("fixture"), 0o644); err != nil {
		t.Fatalf("write avatar fixture: %v", err)
	}
	return "/static/uploads/avatar/20260402/" + filename
}

func localAvatarPath(baseDir string, avatarURL string) string {
	relative := strings.TrimPrefix(avatarURL, "/static/")
	return filepath.Join(baseDir, "storage", filepath.FromSlash(relative))
}

func newMultipartFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("avatar_file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close writer: %v", err)
	}

	reader := multipart.NewReader(bytes.NewReader(body.Bytes()), writer.Boundary())
	form, err := reader.ReadForm(int64(len(body.Bytes())))
	if err != nil {
		t.Fatalf("read form: %v", err)
	}
	files := form.File["avatar_file"]
	if len(files) == 0 {
		t.Fatalf("expected avatar_file in multipart form")
	}
	return files[0]
}
