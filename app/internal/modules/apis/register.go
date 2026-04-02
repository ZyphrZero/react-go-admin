package apis

import (
	"context"

	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/modules/base"
	"react-go-admin/app/internal/platform/database"
)

const (
	moduleOrderAPIs     = 200
	apisServiceKey      = "apis.service"
	apisHandlerKey      = "apis.handler"
	initOrderAPICatalog = 100
)

func init() {
	framework.RegisterModule(framework.ModuleSpec{
		Key:   "apis",
		Order: moduleOrderAPIs,
		Bind: func(container *framework.Container) error {
			service := NewService(container.DB, container)
			container.Set(apisServiceKey, service)
			container.Set(apisHandlerKey, NewHandler(service))
			return nil
		},
		Models: func() []interface{} {
			return []interface{}{
				&database.APIRecord{},
			}
		},
		RouteGroups: func(container *framework.Container) []framework.RouteGroup {
			handler := framework.MustResolve[*Handler](container, apisHandlerKey)
			baseService := framework.MustResolve[*base.Service](container, "base.service")

			return []framework.RouteGroup{
				{
					BasePath:    "/api/v1/api",
					Middlewares: []gin.HandlerFunc{base.GinAuthMiddleware(baseService), base.GinPermissionMiddleware(baseService)},
					Routes: []framework.Route{
						{Method: "GET", Path: "/list", Summary: "查看API列表", Tags: "API管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.List)}},
						{Method: "POST", Path: "/refresh", Summary: "刷新API列表", Tags: "API管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Refresh)}},
						{Method: "GET", Path: "/tags", Summary: "获取所有API标签", Tags: "API管理", Handlers: []gin.HandlerFunc{gin.WrapF(handler.Tags)}},
					},
				},
			}
		},
		InitTasks: func(container *framework.Container) []framework.InitTask {
			return []framework.InitTask{
				{
					Name:  "api_catalog_sync",
					Order: initOrderAPICatalog,
					Run: func(ctx context.Context, container *framework.Container) error {
						return framework.MustResolve[*Service](container, apisServiceKey).Refresh()
					},
				},
			}
		},
	})
}
