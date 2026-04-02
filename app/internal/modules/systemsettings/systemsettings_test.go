package systemsettings

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

func TestUpdateApplicationRemovesPreviousLoginImage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:systemsettings-remove?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&database.SystemSetting{}); err != nil {
		t.Fatalf("migrate system settings: %v", err)
	}

	baseDir := t.TempDir()
	oldURL := writeLoginImageFixture(t, baseDir, "old-remove.webp")
	service := NewService(&config.Config{BaseDir: baseDir}, db)
	if err := service.saveJSON(context.Background(), applicationKey, applicationSettings{
		AppTitle:          "Test",
		ProjectName:       "Test",
		AppDescription:    "Test",
		LoginPageImageURL: oldURL,
	}, "应用配置"); err != nil {
		t.Fatalf("seed application settings: %v", err)
	}

	updated, err := service.UpdateApplication(context.Background(), updateApplicationRequest{
		settings: applicationSettings{
			AppTitle:       "Test",
			ProjectName:    "Test",
			AppDescription: "Test",
		},
		imageAction: "remove",
	}, nil)
	if err != nil {
		t.Fatalf("update application remove image: %v", err)
	}

	if updated.LoginPageImageURL != "" {
		t.Fatalf("expected empty login image url after removal, got %q", updated.LoginPageImageURL)
	}
	if _, err := os.Stat(localLoginImagePath(baseDir, oldURL)); !os.IsNotExist(err) {
		t.Fatalf("expected previous login image to be deleted")
	}
}

func TestUpdateApplicationReplacesPreviousLoginImage(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:systemsettings-replace?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&database.SystemSetting{}); err != nil {
		t.Fatalf("migrate system settings: %v", err)
	}

	baseDir := t.TempDir()
	oldURL := writeLoginImageFixture(t, baseDir, "old-replace.webp")
	service := NewService(&config.Config{BaseDir: baseDir}, db)
	if err := service.saveJSON(context.Background(), applicationKey, applicationSettings{
		AppTitle:          "Test",
		ProjectName:       "Test",
		AppDescription:    "Test",
		LoginPageImageURL: oldURL,
	}, "应用配置"); err != nil {
		t.Fatalf("seed application settings: %v", err)
	}

	imageFile := newLoginImageFileHeader(t, "login.webp", []byte("new-image"))
	updated, err := service.UpdateApplication(context.Background(), updateApplicationRequest{
		settings: applicationSettings{
			AppTitle:       "Test",
			ProjectName:    "Test",
			AppDescription: "Test",
		},
		imageAction: "replace",
	}, imageFile)
	if err != nil {
		t.Fatalf("update application replace image: %v", err)
	}

	if strings.TrimSpace(updated.LoginPageImageURL) == "" {
		t.Fatalf("expected login image url after replacement")
	}
	if updated.LoginPageImageURL == oldURL {
		t.Fatalf("expected login image url to change after replacement")
	}
	if _, err := os.Stat(localLoginImagePath(baseDir, oldURL)); !os.IsNotExist(err) {
		t.Fatalf("expected previous login image to be deleted")
	}
	if _, err := os.Stat(localLoginImagePath(baseDir, updated.LoginPageImageURL)); err != nil {
		t.Fatalf("expected new login image file to exist: %v", err)
	}
}

func writeLoginImageFixture(t *testing.T, baseDir string, filename string) string {
	t.Helper()

	dir := filepath.Join(baseDir, "storage", "uploads", "image", "20260402")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir login image fixture: %v", err)
	}
	fullPath := filepath.Join(dir, filename)
	if err := os.WriteFile(fullPath, []byte("fixture"), 0o644); err != nil {
		t.Fatalf("write login image fixture: %v", err)
	}
	return "/static/uploads/image/20260402/" + filename
}

func localLoginImagePath(baseDir string, imageURL string) string {
	relative := strings.TrimPrefix(imageURL, "/static/")
	return filepath.Join(baseDir, "storage", filepath.FromSlash(relative))
}

func newLoginImageFileHeader(t *testing.T, filename string, content []byte) *multipart.FileHeader {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("login_page_image_file", filename)
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
	files := form.File["login_page_image_file"]
	if len(files) == 0 {
		t.Fatalf("expected login_page_image_file in multipart form")
	}
	return files[0]
}
