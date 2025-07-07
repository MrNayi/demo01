// internal/handler/order_handler.go
package handler

import (
	"demo01/internal/model"
	"demo01/internal/service"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	orderService *service.OrderService
}

// CreateOrderReq 创建订单请求
type CreateOrderReq struct {
	UserID string            `json:"user_id" binding:"required"`
	Items  []model.OrderItem `json:"items" binding:"required"` // 可以一次传入多个商品（我想的是购物车可以批量下单
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// CreateOrderHandler 创建订单接口
func (h *OrderHandler) CreateOrderHandler(c *gin.Context) {
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// 打印详细错误信息
		fmt.Printf("绑定JSON失败: %v, Content-Type: %s\n", err, c.GetHeader("Content-Type"))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.CreateOrder(c.Request.Context(), req.UserID, req.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}

// GetAllOrdersHandler 分页查询所有订单
func (h *OrderHandler) GetAllOrdersHandler(c *gin.Context) {
	// 获取分页后 还需要将分页的string转换为int
	pageStr, pageSizeStr := c.Query("page"), c.Query("page_size")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)
	orders, err := h.orderService.GetAllOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": orders})
}

// GetOrderHandler 查询订单接口
func (h *OrderHandler) GetOrderHandler(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "订单ID不能为空"})
		return
	}

	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": order})
}
