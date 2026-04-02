package base

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/platform/response"
)

func GinAuthMiddleware(baseService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, err := baseService.Authenticate(c.Request.Context(), c.Request.Header.Get("Authorization"))
		if err != nil {
			response.Error(c.Writer, http.StatusUnauthorized, err.Error(), nil)
			c.Abort()
			return
		}
		c.Request = c.Request.WithContext(withCurrentUser(c.Request.Context(), user))
		c.Next()
	}
}

func GinPermissionMiddleware(baseService *Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUserFromContext(c.Request.Context())
		if !ok {
			response.Error(c.Writer, http.StatusUnauthorized, "未授权访问", nil)
			c.Abort()
			return
		}
		routePath := c.FullPath()
		if routePath == "" {
			routePath = c.Request.URL.Path
		}
		if err := baseService.CheckPermission(c.Request.Context(), user, c.Request.Method, normalizePermissionPath(routePath)); err != nil {
			response.Error(c.Writer, http.StatusForbidden, err.Error(), nil)
			c.Abort()
			return
		}
		c.Next()
	}
}

var ginParamPattern = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

func normalizePermissionPath(path string) string {
	return ginParamPattern.ReplaceAllString(path, `{$1}`)
}
