package util

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`    // 状态码
	Message string      `json:"message"` // 响应消息
	Data    interface{} `json:"data"`    // 响应数据
}

// 状态码常量
const (
	CodeSuccess  = 200
	CodeError    = 500 // 服务器错误
	CodeInvalid  = 400 // 参数错误
	CodeNotFound = 404 // 资源不存在
)

// ResponseHelper 响应助手，提供统一的响应方法
type ResponseHelper struct{}

// NewResponseHelper 创建响应助手实例
func NewResponseHelper() *ResponseHelper {
	return &ResponseHelper{}
}

// Success 成功响应
func (h *ResponseHelper) Success(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// Error 错误响应
func (h *ResponseHelper) Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// InvalidParams 参数错误响应
func (h *ResponseHelper) InvalidParams(c *gin.Context, message string) {
	h.Error(c, CodeInvalid, message)
}

// ServerError 服务器错误响应
func (h *ResponseHelper) ServerError(c *gin.Context, message string) {
	h.Error(c, CodeError, message)
}

// NotFound 资源不存在响应
func (h *ResponseHelper) NotFound(c *gin.Context, message string) {
	h.Error(c, CodeNotFound, message)
}

// 全局响应助手实例
var ResponseUtil = NewResponseHelper()
