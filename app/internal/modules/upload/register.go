package upload

import (
	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/modules/base"
)

const (
	moduleOrderUpload = 700
	uploadServiceKey  = "upload.service"
	uploadHandlerKey  = "upload.handler"
)

func init() {
	framework.RegisterModule(framework.ModuleSpec{
		Key:   "upload",
		Order: moduleOrderUpload,
		Bind: func(container *framework.Container) error {
			service := NewService(container.Config, container.DB)
			container.Set(uploadServiceKey, service)
			container.Set(uploadHandlerKey, NewHandler(service))
			return nil
		},
		RouteGroups: func(container *framework.Container) []framework.RouteGroup {
			handler := framework.MustResolve[*Handler](container, uploadHandlerKey)
			baseService := framework.MustResolve[*base.Service](container, "base.service")

			return []framework.RouteGroup{
				{
					BasePath:    "/api/v1/base",
					Middlewares: []gin.HandlerFunc{base.GinAuthMiddleware(baseService)},
					Routes: []framework.Route{
						{Method: "POST", Path: "/upload_avatar", Summary: "上传头像", Tags: "基础模块", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UploadAvatar)}},
					},
				},
				{
					BasePath:    "/api/v1/upload",
					Middlewares: []gin.HandlerFunc{base.GinAuthMiddleware(baseService), base.GinPermissionMiddleware(baseService)},
					Routes: []framework.Route{
						{Method: "POST", Path: "/image", Summary: "上传图片", Tags: "文件上传", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UploadImage)}},
						{Method: "POST", Path: "/files", Summary: "批量上传文件", Tags: "文件上传", Handlers: []gin.HandlerFunc{gin.WrapF(handler.UploadFiles)}},
						{Method: "GET", Path: "/list", Summary: "获取对象存储文件列表", Tags: "文件上传", Handlers: []gin.HandlerFunc{gin.WrapF(handler.ListFiles)}},
						{Method: "DELETE", Path: "/delete", Summary: "删除文件", Tags: "文件上传", Handlers: []gin.HandlerFunc{gin.WrapF(handler.DeleteFile)}},
						{Method: "POST", Path: "/set-public-acl", Summary: "批量设置文件ACL为公共读", Tags: "文件上传", Handlers: []gin.HandlerFunc{gin.WrapF(handler.SetPublicACL)}},
					},
				},
			}
		},
	})
}
