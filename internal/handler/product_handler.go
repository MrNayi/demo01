package handler

import (
	"demo01/internal/model"
	"demo01/internal/service"
	"demo01/internal/util"
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

// 我理解的handler 只做请求的接收 将request进行转换，交给service去进行具体业务逻辑的处理
// 所以在handler主要完成的工作就是数据的组装和响应封装 然后对前端的请求有一个响应

// CreateProductHandler 创建商品
func (h *ProductHandler) CreateProductHandler(c *gin.Context) {
	// 1. 参数绑定和验证
	var product model.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		util.ResponseUtil.InvalidParams(c, "请求参数错误: "+err.Error())
		return
	}

	// 2. 调用 Service 层处理业务逻辑
	ctx := c.Request.Context()
	if err := h.productService.CreateProduct(ctx, &product); err != nil {
		util.ResponseUtil.ServerError(c, "创建商品失败: "+err.Error())
		return
	}

	// 3. 封装并返回成功响应
	util.ResponseUtil.Success(c, "商品创建成功", product)
}

// GetProductHandler 获取商品信息
func (h *ProductHandler) GetProductHandler(c *gin.Context) {
	// 1. 参数获取和验证
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		util.ResponseUtil.InvalidParams(c, "商品ID格式错误")
		return
	}

	// 2. 调用 Service 层查询商品
	ctx := c.Request.Context()
	product, err := h.productService.GetProduct(ctx, productID)
	if err != nil {
		util.ResponseUtil.NotFound(c, "商品不存在")
		return
	}

	// 3. 封装并返回响应
	util.ResponseUtil.Success(c, "查询商品成功", product)
}

// GetAllProductsHandler 获取所有商品（分页）
func (h *ProductHandler) GetAllProductsHandler(c *gin.Context) {
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
	ctx := c.Request.Context()
	products, err := h.productService.GetAllProducts(ctx, page, pageSize)
	if err != nil {
		util.ResponseUtil.ServerError(c, "获取商品列表失败: "+err.Error())
		return
	}

	// 4. 封装并返回响应
	util.ResponseUtil.Success(c, "查询商品列表成功", gin.H{
		"products":  products,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetStockHandler 获取库存
func (h *ProductHandler) GetStockHandler(c *gin.Context) {
	// 1. 参数获取和验证
	productIDStr := c.Param("id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		util.ResponseUtil.InvalidParams(c, "商品ID格式错误")
		return
	}

	// 2. 调用 Service 层查询库存
	ctx := c.Request.Context()
	stock, err := h.productService.GetStock(ctx, productID)
	if err != nil {
		util.ResponseUtil.NotFound(c, "获取库存失败")
		return
	}

	// 3. 封装并返回响应
	util.ResponseUtil.Success(c, "查询库存成功", gin.H{
		"product_id": productID,
		"stock":      stock,
	})
}

// RecommendProductsHandler 猜你喜欢商品推荐接口
func (h *ProductHandler) RecommendProductsHandler(c *gin.Context) {
	idsStr := c.Query("ids")
	var ids []int
	if idsStr != "" {
		idStrs := util.SplitAndTrim(idsStr, ",")
		for _, s := range idStrs {
			id, err := strconv.Atoi(s)
			if err != nil {
				util.ResponseUtil.InvalidParams(c, "商品ID格式错误: "+s)
				return
			}
			ids = append(ids, id)
		}
	}

	ctx := c.Request.Context()
	products, err := h.productService.RecommendProducts(ctx, ids)
	if err != nil {
		util.ResponseUtil.ServerError(c, "推荐商品失败: "+err.Error())
		return
	}

	util.ResponseUtil.Success(c, "推荐商品成功 (WaitGroup并发)", gin.H{
		"mode":     "waitgroup",
		"products": products,
	})
}

// RecommendProductsSerialHandler 串行方式批量查询商品详情（仅测试用）
func (h *ProductHandler) RecommendProductsSerialHandler(c *gin.Context) {
	idsStr := c.Query("ids")
	var ids []int
	if idsStr != "" {
		idStrs := util.SplitAndTrim(idsStr, ",")
		for _, s := range idStrs {
			id, err := strconv.Atoi(s)
			if err != nil {
				util.ResponseUtil.InvalidParams(c, "商品ID格式错误: "+s)
				return
			}
			ids = append(ids, id)
		}
	}

	ctx := c.Request.Context()
	products, err := h.productService.RecommendProductsSerial(ctx, ids)
	if err != nil {
		util.ResponseUtil.ServerError(c, "串行推荐商品失败: "+err.Error())
		return
	}
	util.ResponseUtil.Success(c, "串行推荐商品成功 (Serial)", gin.H{
		"mode":     "serial",
		"products": products,
	})
}
