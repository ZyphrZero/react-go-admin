package users

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/modules/base"
	"react-go-admin/app/internal/platform/database"
	"react-go-admin/app/internal/platform/password"
)

const (
	moduleOrderUsers      = 400
	usersServiceKey       = "users.service"
	usersHandlerKey       = "users.handler"
	initOrderAdminUser    = 300
	initOrderAdminBinding = 400
)

func init() {
	framework.RegisterModule(framework.ModuleSpec{
		Key:   "users",
		Order: moduleOrderUsers,
		Bind: func(container *framework.Container) error {
			authz := framework.MustResolve[*base.Service](container, "base.service")
			service := NewService(container.DB, authz)
			container.Set(usersServiceKey, service)
			container.Set(usersHandlerKey, NewHandler(service))
			return nil
		},
		Models: func() []interface{} {
			return []interface{}{
				&database.User{},
				&database.UserRole{},
			}
		},
		RouteGroups: func(container *framework.Container) []framework.RouteGroup {
			handler := framework.MustResolve[*Handler](container, usersHandlerKey)
			baseService := framework.MustResolve[*base.Service](container, "base.service")

			return []framework.RouteGroup{
				{
					BasePath:    "/api/v1/user",
					Middlewares: []gin.HandlerFunc{base.GinAuthMiddleware(baseService), base.GinPermissionMiddleware(baseService)},
					Routes: []framework.Route{
						{Method: "GET", Path: "/list", Summary: "查看用户列表", Tags: "用户管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.List)}},
						{Method: "GET", Path: "/get", Summary: "查看用户", Tags: "用户管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Get)}},
						{Method: "POST", Path: "/create", Summary: "创建用户", Tags: "用户管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Create)}},
						{Method: "POST", Path: "/update", Summary: "更新用户", Tags: "用户管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Update)}},
						{Method: "DELETE", Path: "/delete", Summary: "删除用户", Tags: "用户管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Delete)}},
						{Method: "POST", Path: "/reset_password", Summary: "重置密码", Tags: "用户管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.ResetPassword)}},
					},
				},
			}
		},
		InitTasks: func(container *framework.Container) []framework.InitTask {
			return []framework.InitTask{
				{
					Name:  "admin_user",
					Order: initOrderAdminUser,
					Run:   ensureAdminUser,
				},
				{
					Name:  "admin_role_assignment",
					Order: initOrderAdminBinding,
					Run:   ensureAdminRoleAssignment,
				},
			}
		},
	})
}

func ensureAdminUser(ctx context.Context, container *framework.Container) error {
	cfg := container.Config
	db := container.DB

	var existing database.User
	result := db.WithContext(ctx).Where("username = ?", cfg.InitialAdminUsername).Limit(1).Find(&existing)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected > 0 {
		if existing.IsSuperuser {
			return nil
		}
		return fmt.Errorf("username %s already exists but is not superuser", cfg.InitialAdminUsername)
	}

	policy := password.NewPolicy(cfg)
	initialPassword := strings.TrimSpace(cfg.InitialAdminPassword)
	generatedPassword := ""
	if initialPassword == "" {
		var err error
		initialPassword, err = password.GenerateBootstrapPassword(12)
		if err != nil {
			return err
		}
		generatedPassword = initialPassword
	} else if err := policy.Validate(initialPassword); err != nil {
		return fmt.Errorf("INITIAL_ADMIN_PASSWORD 不满足密码策略: %w", err)
	}

	hashedPassword, err := password.Hash(initialPassword)
	if err != nil {
		return err
	}

	user := &database.User{
		Username:       cfg.InitialAdminUsername,
		Email:          stringPtr(strings.TrimSpace(cfg.InitialAdminEmail)),
		Nickname:       stringPtr(strings.TrimSpace(cfg.InitialAdminNickname)),
		Password:       hashedPassword,
		IsActive:       true,
		IsSuperuser:    true,
		SessionVersion: 0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	if err := db.WithContext(ctx).Create(user).Error; err != nil {
		return err
	}

	if generatedPassword != "" && container.Logger != nil {
		container.Logger.Warn("bootstrap admin password generated automatically", "username", cfg.InitialAdminUsername, "password", generatedPassword)
	}
	return nil
}

func ensureAdminRoleAssignment(ctx context.Context, container *framework.Container) error {
	cfg := container.Config
	db := container.DB

	var adminUser database.User
	userResult := db.WithContext(ctx).Where("username = ?", cfg.InitialAdminUsername).Limit(1).Find(&adminUser)
	if userResult.Error != nil {
		return userResult.Error
	}
	if userResult.RowsAffected == 0 {
		return nil
	}

	var adminRole database.Role
	if err := db.WithContext(ctx).Where("name = ?", "管理员").First(&adminRole).Error; err != nil {
		return err
	}

	var count int64
	if err := db.WithContext(ctx).Model(&database.UserRole{}).
		Where("user_id = ? AND role_id = ?", adminUser.ID, adminRole.ID).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	return db.WithContext(ctx).Create(&database.UserRole{
		UserID: adminUser.ID,
		RoleID: adminRole.ID,
	}).Error
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
