package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"slg_game_gm/server/internal/model"
	"slg_game_gm/server/internal/service"
)

const maxLoggedBody = 4 << 10

var sensitiveFields = []string{"password", "old_password", "new_password"}

// OperationLog 记录所有写操作（非 GET）的审计日志，需挂在 Auth 之后。
func OperationLog(logs *service.OpLogService) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}
		var body []byte
		if c.Request.Body != nil {
			body, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		start := time.Now()
		c.Next()

		entry := &model.OperationLog{
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			Query:     c.Request.URL.RawQuery,
			Body:      maskBody(body),
			IP:        c.ClientIP(),
			Status:    c.Writer.Status(),
			LatencyMs: time.Since(start).Milliseconds(),
			CreatedAt: start,
		}
		if u := CurrentUser(c); u != nil {
			entry.AdminID, entry.Username = u.ID.Hex(), u.Username
		}
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			if err := logs.Record(ctx, entry); err != nil {
				slog.Error("record operation log failed", "err", err)
			}
		}()
	}
}

func maskBody(body []byte) string {
	if len(body) == 0 {
		return ""
	}
	var m map[string]any
	if json.Unmarshal(body, &m) == nil {
		for _, k := range sensitiveFields {
			if _, ok := m[k]; ok {
				m[k] = "******"
			}
		}
		body, _ = json.Marshal(m)
	}
	if len(body) > maxLoggedBody {
		body = body[:maxLoggedBody]
	}
	return string(body)
}
