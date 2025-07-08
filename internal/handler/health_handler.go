package handler

import (
	"context"
	"demo01/internal/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck 健康检查接口
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	ctx := c.Request.Context()

	// 检查数据库连接
	dbStatus := "ok"
	if err := h.checkDatabase(ctx); err != nil {
		dbStatus = "error"
		util.GlobalLogger.Error(ctx, "数据库健康检查失败", err)
	}

	// 检查Redis连接
	redisStatus := "ok"
	if err := h.checkRedis(ctx); err != nil {
		redisStatus = "error"
		util.GlobalLogger.Error(ctx, "Redis健康检查失败", err)
	}

	// 返回健康状态
	status := "healthy"
	if dbStatus == "error" || redisStatus == "error" {
		status = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": status,
			"checks": gin.H{
				"database": dbStatus,
				"redis":    redisStatus,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"checks": gin.H{
			"database": dbStatus,
			"redis":    redisStatus,
		},
	})
}

// checkDatabase 检查数据库连接
func (h *HealthHandler) checkDatabase(ctx context.Context) error {
	// 这里可以添加数据库连接检查逻辑
	// 暂时返回nil，表示健康
	return nil
}

// checkRedis 检查Redis连接
func (h *HealthHandler) checkRedis(ctx context.Context) error {
	return util.PingRedis(ctx)
}
