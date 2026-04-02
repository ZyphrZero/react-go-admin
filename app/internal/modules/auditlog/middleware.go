package auditlog

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"react-go-admin/app/internal/framework"
	"react-go-admin/app/internal/modules/base"
	"react-go-admin/app/internal/platform/database"
)

const maxAuditPayloadBytes = 8 * 1024

type responseBodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (r *responseBodyWriter) Write(data []byte) (int, error) {
	if r.body != nil && r.body.Len() < maxAuditPayloadBytes {
		remaining := maxAuditPayloadBytes - r.body.Len()
		if len(data) > remaining {
			r.body.Write(data[:remaining])
		} else {
			r.body.Write(data)
		}
	}
	return r.ResponseWriter.Write(data)
}

var auditParamPattern = regexp.MustCompile(`:([A-Za-z0-9_]+)`)

// GinMiddleware persists API request/response snapshots into audit_log.
func GinMiddleware(container *framework.Container) gin.HandlerFunc {
	metaByRoute := buildRouteMetaMap(container)

	return func(c *gin.Context) {
		if container == nil || container.DB == nil {
			c.Next()
			return
		}
		if c.Request.Method == http.MethodOptions || !strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Next()
			return
		}

		requestArgs := captureRequestArgs(c.Request)
		writer := &responseBodyWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := c.Writer.Status()
		if status == 0 {
			status = http.StatusOK
		}

		user, ok := base.CurrentUserFromContext(c.Request.Context())
		userID := int64(0)
		username := "anonymous"
		if ok && user != nil {
			userID = user.ID
			if strings.TrimSpace(user.Username) != "" {
				username = user.Username
			}
		}

		routePath := c.FullPath()
		if routePath == "" {
			routePath = c.Request.URL.Path
		}
		meta := metaByRoute[strings.ToUpper(c.Request.Method)+" "+normalizeAuditPath(routePath)]

		record := &database.AuditLog{
			UserID:        userID,
			Username:      username,
			Module:        fallbackString(meta.Tags, inferModule(c.Request.URL.Path)),
			Summary:       fallbackString(meta.Summary, strings.TrimSpace(c.Request.Method+" "+c.Request.URL.Path)),
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			Status:        status,
			ResponseTime:  int(duration.Milliseconds()),
			RequestArgs:   database.JSONBytes(requestArgs),
			ResponseBody:  database.JSONBytes(normalizeAuditPayload(writer.body.Bytes())),
			IPAddress:     c.ClientIP(),
			UserAgent:     c.Request.UserAgent(),
			OperationType: operationTypeForMethod(c.Request.Method),
			LogLevel:      logLevelForStatus(status),
			IsDeleted:     false,
		}

		// Persist audit data even when the client disconnects immediately after the response.
		auditCtx, cancel := context.WithTimeout(context.WithoutCancel(c.Request.Context()), 3*time.Second)
		defer cancel()
		if err := container.DB.WithContext(auditCtx).Create(record).Error; err != nil && container.Logger != nil {
			container.Logger.Error("create audit log failed", "error", err, "path", c.Request.URL.Path, "method", c.Request.Method)
		}
	}
}

func buildRouteMetaMap(container *framework.Container) map[string]framework.RouteMeta {
	result := make(map[string]framework.RouteMeta)
	for _, meta := range framework.RegisteredRouteMetadata(container) {
		key := strings.ToUpper(meta.Method) + " " + normalizeAuditPath(meta.Path)
		result[key] = meta
	}
	return result
}

func captureRequestArgs(r *http.Request) []byte {
	if r == nil {
		return nil
	}

	if strings.Contains(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
		return []byte(`"[multipart/form-data]"`)
	}

	if r.Method != http.MethodGet && r.Body != nil {
		body, err := ioReadAllAndRestore(r)
		if err == nil && len(bytes.TrimSpace(body)) > 0 {
			return normalizeAuditPayload(body)
		}
	}

	if strings.TrimSpace(r.URL.RawQuery) == "" {
		return nil
	}

	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return normalizeAuditPayload([]byte(r.URL.RawQuery))
	}

	payload := make(map[string]interface{}, len(values))
	for key, items := range values {
		if len(items) == 1 {
			payload[key] = items[0]
			continue
		}
		payload[key] = items
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return normalizeAuditPayload([]byte(r.URL.RawQuery))
	}
	return encoded
}

func ioReadAllAndRestore(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return body, nil
}

func normalizeAuditPayload(raw []byte) []byte {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 {
		return nil
	}
	if len(trimmed) > maxAuditPayloadBytes {
		trimmed = append([]byte(nil), trimmed[:maxAuditPayloadBytes]...)
	}
	if json.Valid(trimmed) {
		return trimmed
	}
	encoded, err := json.Marshal(string(trimmed))
	if err != nil {
		return nil
	}
	return encoded
}

func normalizeAuditPath(path string) string {
	return auditParamPattern.ReplaceAllString(path, `{$1}`)
}

func operationTypeForMethod(method string) string {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet:
		return "查询"
	case http.MethodPost:
		return "提交"
	case http.MethodPut, http.MethodPatch:
		return "更新"
	case http.MethodDelete:
		return "删除"
	default:
		return "操作"
	}
}

func logLevelForStatus(status int) string {
	switch {
	case status >= 500:
		return "error"
	case status >= 400:
		return "warning"
	default:
		return "info"
	}
}

func inferModule(path string) string {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return "系统接口"
	}
	parts := strings.Split(trimmed, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return parts[len(parts)-1]
}

func fallbackString(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value != "" {
		return value
	}
	return fallback
}
