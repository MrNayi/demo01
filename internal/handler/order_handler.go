// internal/handler/order_handler.go
package handler

import (
	"demo01/internal/model"
	"demo01/internal/service"
	"demo01/internal/util"
	"fmt"
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
	// 1. 参数绑定和验证
	var req CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		// 打印详细错误信息用于调试
		fmt.Printf("绑定JSON失败: %v, Content-Type: %s\n", err, c.GetHeader("Content-Type"))
		util.ResponseUtil.InvalidParams(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 参数业务验证
	if len(req.Items) == 0 {
		util.ResponseUtil.InvalidParams(c, "订单商品不能为空")
		return
	}

	// 3. 调用 Service 层处理业务逻辑
	order, err := h.orderService.CreateOrder(c.Request.Context(), req.UserID, req.Items)
	if err != nil {
		util.ResponseUtil.ServerError(c, "创建订单失败: "+err.Error())
		return
	}

	// 4. 封装并返回成功响应
	util.ResponseUtil.Success(c, "订单创建成功", order)
}

// GetAllOrdersHandler 分页查询所有订单
func (h *OrderHandler) GetAllOrdersHandler(c *gin.Context) {
	// 1. 参数获取和转换
	pageStr, pageSizeStr := c.Query("page"), c.Query("page_size")
	page, _ := strconv.Atoi(pageStr)
	pageSize, _ := strconv.Atoi(pageSizeStr)

	// 2. 参数验证和默认值设置
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大分页大小
	}

	// 3. 调用 Service 层查询数据
	orders, err := h.orderService.GetAllOrders(c.Request.Context(), page, pageSize)
	if err != nil {
		util.ResponseUtil.ServerError(c, "查询订单列表失败: "+err.Error())
		return
	}

	// 4. 封装并返回响应
	util.ResponseUtil.Success(c, "查询订单列表成功", gin.H{
		"orders":    orders,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetOrderHandler 查询订单接口
func (h *OrderHandler) GetOrderHandler(c *gin.Context) {
	// 1. 参数获取和验证
	orderID := c.Param("id")
	if orderID == "" {
		util.ResponseUtil.InvalidParams(c, "订单ID不能为空")
		return
	}

	// 2. 调用 Service 层查询订单
	order, err := h.orderService.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		util.ResponseUtil.NotFound(c, "订单不存在")
		return
	}

	// 3. 封装并返回响应
	util.ResponseUtil.Success(c, "查询订单成功", order)
}
