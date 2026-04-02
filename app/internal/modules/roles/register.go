package roles

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/modules/base"
	"react-go-admin/app/internal/platform/database"
)

const (
	moduleOrderRoles          = 300
	rolesServiceKey           = "roles.service"
	rolesHandlerKey           = "roles.handler"
	initOrderDefaultRoles     = 200
	initOrderDefaultRolePerms = 500
)

func init() {
	framework.RegisterModule(framework.ModuleSpec{
		Key:   "roles",
		Order: moduleOrderRoles,
		Bind: func(container *framework.Container) error {
			authz := framework.MustResolve[*base.Service](container, "base.service")
			service := NewService(container.DB, authz)
			container.Set(rolesServiceKey, service)
			container.Set(rolesHandlerKey, NewHandler(service))
			return nil
		},
		Models: func() []interface{} {
			return []interface{}{
				&database.Role{},
			}
		},
		RouteGroups: func(container *framework.Container) []framework.RouteGroup {
			handler := framework.MustResolve[*Handler](container, rolesHandlerKey)
			baseService := framework.MustResolve[*base.Service](container, "base.service")

			return []framework.RouteGroup{
				{
					BasePath:    "/api/v1/role",
					Middlewares: []gin.HandlerFunc{base.GinAuthMiddleware(baseService), base.GinPermissionMiddleware(baseService)},
					Routes: []framework.Route{
						{Method: "GET", Path: "/list", Summary: "查看角色列表", Tags: "角色管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.List)}},
						{Method: "GET", Path: "/get", Summary: "查看角色", Tags: "角色管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Get)}},
						{Method: "GET", Path: "/permission_options", Summary: "获取角色权限选项", Tags: "角色管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.PermissionOptions)}},
						{Method: "POST", Path: "/create", Summary: "创建角色", Tags: "角色管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Create)}},
						{Method: "POST", Path: "/update", Summary: "更新角色", Tags: "角色管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Update)}},
						{Method: "DELETE", Path: "/delete", Summary: "删除角色", Tags: "角色管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Delete)}},
					},
				},
			}
		},
		InitTasks: func(container *framework.Container) []framework.InitTask {
			return []framework.InitTask{
				{
					Name:  "default_roles",
					Order: initOrderDefaultRoles,
					Run:   ensureDefaultRoles,
				},
				{
					Name:  "default_role_permissions",
					Order: initOrderDefaultRolePerms,
					Run:   syncDefaultRolePermissions,
				},
			}
		},
	})
}

func ensureDefaultRoles(ctx context.Context, container *framework.Container) error {
	db := container.DB
	defaultRoles := []struct {
		Name string
		Desc string
	}{
		{Name: "管理员", Desc: "管理员角色"},
		{Name: "普通用户", Desc: "普通用户角色"},
	}

	for _, roleDef := range defaultRoles {
		var existing database.Role
		result := db.WithContext(ctx).Where("name = ?", roleDef.Name).Limit(1).Find(&existing)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected > 0 {
			continue
		}

		desc := roleDef.Desc
		role := database.Role{
			Name:      roleDef.Name,
			Desc:      &desc,
			MenuPaths: database.JSONStringSlice(base.DefaultRoleMenuPaths(roleDef.Name)),
			APIIDs:    database.JSONInt64Slice{},
		}
		if err := db.WithContext(ctx).Create(&role).Error; err != nil {
			return err
		}
	}
	return nil
}

func syncDefaultRolePermissions(ctx context.Context, container *framework.Container) error {
	db := container.DB

	var allAPIs []database.APIRecord
	if err := db.WithContext(ctx).Order("id ASC").Find(&allAPIs).Error; err != nil {
		return err
	}
	allAPIIDs := make([]int64, 0, len(allAPIs))
	for _, api := range allAPIs {
		allAPIIDs = append(allAPIIDs, api.ID)
	}

	for _, roleName := range []string{"管理员", "普通用户"} {
		var role database.Role
		if err := db.WithContext(ctx).Where("name = ?", roleName).First(&role).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{
			"updated_at": time.Now(),
		}
		changed := false
		if len(role.MenuPaths) == 0 {
			updates["menu_paths"] = database.JSONStringSlice(base.DefaultRoleMenuPaths(roleName))
			changed = true
		}
		if roleName == "管理员" && len(role.APIIDs) == 0 {
			updates["api_ids"] = database.JSONInt64Slice(allAPIIDs)
			changed = true
		}
		if roleName == "普通用户" && role.APIIDs == nil {
			updates["api_ids"] = database.JSONInt64Slice{}
			changed = true
		}
		if changed {
			if err := db.WithContext(ctx).Model(&role).Updates(updates).Error; err != nil {
				return err
			}
		}
	}

	return nil
}
