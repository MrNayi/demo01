package service

import (
	"context"
	"demo01/internal/model"
	"demo01/internal/repository"
	"sync"
)

type ProductService struct {
	productRepo *repository.ProductRepo
	localCache  sync.Map // 本地缓存，存储热点商品信息
}

func NewProductService(productRepo *repository.ProductRepo) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}

// GetProduct 获取商品信息（带缓存）
func (s *ProductService) GetProduct(ctx context.Context, productID int) (*model.Product, error) {
	// 1. 先从本地缓存查询
	if cacheValue, exists := s.localCache.Load(productID); exists {
		if cache, ok := cacheValue.(*model.Product); ok {
			return cache, nil
		}
	}

	// 2. 从数据库查询
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return nil, err
	}

	// 3. 写入本地缓存
	s.localCache.Store(productID, product)

	return product, nil
}

// GetAllProducts 获取所有商品（分页）
func (s *ProductService) GetAllProducts(ctx context.Context, page, pageSize int) ([]model.Product, error) {
	return s.productRepo.GetAll(ctx, page, pageSize)
}

// CreateProduct 创建商品
func (s *ProductService) CreateProduct(ctx context.Context, product *model.Product) error {
	return s.productRepo.Create(ctx, product)
}

// DecreaseStock 扣减库存（乐观锁）
func (s *ProductService) DecreaseStock(ctx context.Context, productID, quantity int) error {
	// 1. 获取商品信息
	product, err := s.productRepo.GetByID(ctx, productID)
	if err != nil {
		return err
	}

	// 2. 检查库存
	if product.Stock < quantity {
		return repository.ErrInsufficientStock
	}

	// 3. 乐观锁更新库存
	err = s.productRepo.UpdateStock(ctx, productID, quantity, product.Version)
	if err != nil {
		return err
	}

	// 4. 清除本地缓存
	s.localCache.Delete(productID)

	return nil
}

// GetStock 获取库存
func (s *ProductService) GetStock(ctx context.Context, productID int) (int, error) {
	return s.productRepo.GetStock(ctx, productID)
}
