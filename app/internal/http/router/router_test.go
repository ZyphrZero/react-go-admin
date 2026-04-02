package router_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"react-go-admin/app/internal/config"
	"react-go-admin/app/internal/framework"
	httprouter "react-go-admin/app/internal/http/router"
	"react-go-admin/app/internal/migrate"
	moduleregistry "react-go-admin/app/internal/modules"
	"react-go-admin/app/internal/platform/database"
	"react-go-admin/app/internal/seed"
)

func TestProtectedRequestCreatesAuditLog(t *testing.T) {
	moduleregistry.Register()

	db, err := gorm.Open(sqlite.Open("file:router-audit-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	cfg := &config.Config{
		AppEnv:                      "test",
		Version:                     "test",
		AppTitle:                    "Test",
		ProjectName:                 "Test",
		AppDescription:              "Test",
		BaseDir:                     t.TempDir(),
		SecretKey:                   strings.Repeat("a", 64),
		JWTAlgorithm:                "HS256",
		JWTAccessTokenExpireMinutes: 15,
		JWTRefreshTokenExpireDays:   7,
		RefreshTokenCookieName:      "refresh_token",
		RefreshTokenCookieSameSite:  "lax",
		CORSOrigins:                 []string{"*"},
		CORSAllowCredentials:        true,
		CORSAllowMethods:            []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		CORSAllowHeaders:            []string{"*"},
		InitialAdminUsername:        "admin",
		InitialAdminEmail:           "admin@example.com",
		InitialAdminNickname:        "admin",
		InitialAdminPassword:        "Admin@123!@#",
		PasswordMinLength:           8,
		PasswordRequireUppercase:    true,
		PasswordRequireLowercase:    true,
		PasswordRequireDigits:       true,
		PasswordRequireSpecial:      true,
		NotificationPosition:        "top-right",
		NotificationDuration:        4000,
		NotificationVisibleToasts:   3,
		EnableInitAPI:               true,
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	container := framework.NewContainer(cfg, logger, db)
	if err := framework.BuildModules(container); err != nil {
		t.Fatalf("build modules: %v", err)
	}

	ctx := context.Background()
	if err := migrate.Up(ctx, db); err != nil {
		t.Fatalf("migrate up: %v", err)
	}
	if err := seed.Run(ctx, container); err != nil {
		t.Fatalf("seed run: %v", err)
	}

	engine := httprouter.New(container)

	loginBody := bytes.NewBufferString(`{"username":"admin","password":"Admin@123!@#"}`)
	loginReq := httptest.NewRequest(http.MethodPost, "/api/v1/base/access_token", loginBody)
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp := httptest.NewRecorder()
	engine.ServeHTTP(loginResp, loginReq)
	if loginResp.Code != http.StatusOK {
		t.Fatalf("login status = %d, body = %s", loginResp.Code, loginResp.Body.String())
	}

	var loginPayload struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginResp.Body.Bytes(), &loginPayload); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginPayload.Data.AccessToken == "" {
		t.Fatalf("expected access token in login response")
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/v1/user/list?page=1&page_size=10", nil)
	listReq.Header.Set("Authorization", "Bearer "+loginPayload.Data.AccessToken)
	listResp := httptest.NewRecorder()
	engine.ServeHTTP(listResp, listReq)
	if listResp.Code != http.StatusOK {
		t.Fatalf("user list status = %d, body = %s", listResp.Code, listResp.Body.String())
	}

	var log database.AuditLog
	if err := db.WithContext(ctx).Where("path = ?", "/api/v1/user/list").Order("id DESC").First(&log).Error; err != nil {
		t.Fatalf("expected audit log for user list request: %v", err)
	}
	if log.Username != "admin" {
		t.Fatalf("expected audit username admin, got %q", log.Username)
	}
	if log.Method != http.MethodGet {
		t.Fatalf("expected audit method GET, got %q", log.Method)
	}
	if log.Status != http.StatusOK {
		t.Fatalf("expected audit status 200, got %d", log.Status)
	}
}
