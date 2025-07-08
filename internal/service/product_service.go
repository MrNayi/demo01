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

// 创建商品服务实例
func NewProductService(productRepo *repository.ProductRepo) *ProductService {
	return &ProductService{
		// 提供操作数据库的实例
		productRepo: productRepo,
	}
}

// GetProduct 获取商品信息
func (s *ProductService) GetProduct(ctx context.Context, productID int) (*model.Product, error) {
	// 1. 先从本地缓存查询
	// 本地缓存指的是当前进程的内存空间 不是用户浏览器的本地存储也不是单个协程的内存空间
	// 结合go的并发模型 我们其实不难理解 一个进程可能对应很多个协程 那么从这个意义上来说 是否可以认为在本地缓存中 也是有多个协程去共享内存的
	// 从而产生了竞态条件 需要加锁 但是加锁的性能损耗很大 所以需要使用sync.map
	// 本地缓存局限于一个进程内 如果想要实现全局的缓存 需要使用redis
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

	// TODO 4. 写入redis缓存
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

// RecommendProducts 批量推荐商品
// 使用wait group 并发查询每个商品的信息 统一聚合结果
func (s *ProductService) RecommendProducts(ctx context.Context, ids []int) ([]*model.Product, error) {
	if len(ids) == 0 {
		// 默认推荐前10个商品
		products, err := s.productRepo.GetAll(ctx, 1, 10)
		if err != nil {
			return nil, err
		}
		result := make([]*model.Product, 0, len(products))
		for i := range products {
			result = append(result, &products[i])
		}
		return result, nil
	}

	// 并发拉取每个商品详情
	result := make([]*model.Product, len(ids))
	errList := make([]error, len(ids))
	// 通过waitgroup并发查询每一个商品详情 最后汇聚成一个推荐页
	// 比如我们的推荐页默认为10个商品 那么并发查找就能够让原来的10个商品查询时间 变为最慢的一个
	// 这是我想到的适合waitgroup的场景
	var wg sync.WaitGroup

	for i, id := range ids {
		wg.Add(1)
		go func(idx, pid int) {
			defer wg.Done()
			// 先查本地缓存
			if cacheValue, exists := s.localCache.Load(pid); exists {
				if cache, ok := cacheValue.(*model.Product); ok {
					result[idx] = cache
					return
				}
			}
			// 查数据库
			product, err := s.productRepo.GetByID(ctx, pid)
			if err != nil {
				errList[idx] = err
				return
			}
			result[idx] = product
			// 写入本地缓存
			s.localCache.Store(pid, product)
		}(i, id)
	}
	wg.Wait()

	// 过滤掉未查到的商品（如有）
	finalResult := make([]*model.Product, 0, len(ids))
	for i, p := range result {
		if p != nil && errList[i] == nil {
			finalResult = append(finalResult, p)
		}
	}
	return finalResult, nil
}

// RecommendProductsSerial 串行方式批量查询商品详情
func (s *ProductService) RecommendProductsSerial(ctx context.Context, ids []int) ([]*model.Product, error) {
	var products []*model.Product
	if len(ids) == 0 {
		// 默认推荐前10个商品
		all, err := s.productRepo.GetAll(ctx, 1, 10)
		if err != nil {
			return nil, err
		}
		for i := range all {
			products = append(products, &all[i])
		}
		return products, nil
	}
	for _, id := range ids {
		product, err := s.GetProduct(ctx, id)
		if err == nil && product != nil {
			products = append(products, product)
		}
	}
	return products, nil
}
