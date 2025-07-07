package handler

import (
	"demo01/internal/model"
	"demo01/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productService *service.ProductService
}

func NewProductHandler(productService *service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// CreateProductHandler 创建商品
func (h *ProductHandler) CreateProductHandler(c *gin.Context) {
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误: " + err.Error()})
		return
	}

	ctx := c.Request.Context()
	if err := h.productService.CreateProduct(ctx, &product); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建商品失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "商品创建成功",
		"data":    product,
	})
}

// GetProductHandler 获取商品信息
func (h *ProductHandler) GetProductHandler(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品ID格式错误"})
		return
	}

	ctx := c.Request.Context()
	product, err := h.productService.GetProduct(ctx, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "商品不存在: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": product,
	})
}

// GetAllProductsHandler 获取所有商品（分页）
func (h *ProductHandler) GetAllProductsHandler(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	ctx := c.Request.Context()
	products, err := h.productService.GetAllProducts(ctx, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取商品列表失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": products,
		"page": page,
		"size": pageSize,
	})
}

// GetStockHandler 获取库存
func (h *ProductHandler) GetStockHandler(c *gin.Context) {
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "商品ID格式错误"})
		return
	}

	ctx := c.Request.Context()
	stock, err := h.productService.GetStock(ctx, productID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "获取库存失败: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id": productID,
		"stock":      stock,
	})
}
