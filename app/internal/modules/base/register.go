package base

import (
	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/platform/database"
)

const (
	moduleOrder = 100
	serviceKey  = "base.service"
	handlerKey  = "base.handler"
)

func init() {
	framework.RegisterModule(framework.ModuleSpec{
		Key:   "base",
		Order: moduleOrder,
		Bind: func(container *framework.Container) error {
			service := NewService(container.Config, container.DB)
			container.Set(serviceKey, service)
			container.Set(handlerKey, NewHandler(service))
			return nil
		},
		Models: func() []interface{} {
			return []interface{}{
				&database.RateLimitBucket{},
			}
		},
		RouteGroups: func(container *framework.Container) []framework.RouteGroup {
			handler := framework.MustResolve[*Handler](container, handlerKey)
			service := framework.MustResolve[*Service](container, serviceKey)
			authMiddleware := GinAuthMiddleware(service)

			return []framework.RouteGroup{
				{
					BasePath: "/api/v1/base",
					Routes: []framework.Route{
						{Method: "POST", Path: "/access_token", Summary: "获取token", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Login)}},
						{Method: "POST", Path: "/refresh_token", Summary: "刷新访问令牌", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Refresh)}},
						{Method: "GET", Path: "/app_meta", Summary: "获取应用元信息", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.AppMeta)}},
					},
				},
				{
					BasePath:    "/api/v1/base",
					Middlewares: []gin.HandlerFunc{authMiddleware},
					Routes: []framework.Route{
						{Method: "GET", Path: "/userinfo", Summary: "查看用户信息", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UserInfo)}},
						{Method: "GET", Path: "/password_policy", Summary: "查看密码策略", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.PasswordPolicy)}},
						{Method: "POST", Path: "/update_password", Summary: "修改密码", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UpdatePassword)}},
						{Method: "POST", Path: "/update_profile", Summary: "更新个人信息", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UpdateProfile)}},
						{Method: "POST", Path: "/logout", Summary: "用户注销", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Logout)}},
						{Method: "GET", Path: "/usermenu", Summary: "查看用户菜单", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UserMenu)}},
						{Method: "GET", Path: "/userapi", Summary: "查看用户API", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UserAPI)}},
						{Method: "GET", Path: "/overview", Summary: "查看管理台概览", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Overview)}},
					},
				},
			}
		},
	})
}
