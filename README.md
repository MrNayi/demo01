# Demo01 - Go 订单管理系统

一个基于 Go 语言开发的订单管理系统，展示了 Go 并发编程、RESTful API 设计、数据库操作等核心特性。

## 项目特性

- **分层架构**：Handler → Service → Repository → Database
- **并发控制**：使用 Goroutine、WaitGroup、Channel 实现库存扣减
- **乐观锁**：数据库层面的并发控制
- **本地缓存**：使用 map + 读写锁实现高性能查询
- **RESTful API**：标准的 REST 接口设计
- **依赖注入**：便于测试和维护

## 技术栈

- **语言**：Go 1.21.4
- **Web 框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL
- **缓存**：Redis（可选）
- **消息队列**：Kafka（可选）

## 项目结构

```
demo01/
├── cmd/main.go                    # 主程序入口
├── config/config.go               # 配置管理
├── internal/
│   ├── handler/order_handler.go   # HTTP 处理器
│   ├── service/order_service.go   # 业务逻辑层
│   ├── repository/                # 数据访问层
│   │   ├── order_repo.go
│   │   └── inventory_repo.go
│   ├── model/order.go             # 数据模型
│   └── util/redis.go              # Redis 工具
├── test_order.json                # 测试数据
└── README.md
```

## API 接口

### 创建订单
```bash
POST /orders
Content-Type: application/json

{
  "user_id": "user123",
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "price": 99.99
    }
  ]
}
```

### 查询订单
```bash
GET /orders/:id
```

### 查询所有订单
```bash
GET /orders?page=1&page_size=10
```

## 并发特性

### 1. 库存扣减并发控制
- 使用 Goroutine 并发处理多个商品库存扣减
- WaitGroup 等待所有扣减完成
- Channel 传递错误信息
- 读写锁保护共享资源

### 2. 乐观锁机制
- 使用 version 字段防止并发更新冲突
- 事务保证原子性

### 3. 本地缓存
- 读写锁保护 map 的并发访问
- 读多写少的场景优化

## 运行项目

### 1. 安装依赖
```bash
go mod tidy
```

### 2. 配置环境变量
```bash
export MYSQL_DSN="root:password@tcp(localhost:3306)/demo01?charset=utf8mb4&parseTime=True&loc=Local"
export REDIS_ADDR="localhost:6379"
export KAFKA_ADDR="localhost:9092"
export PORT="8080"
```

### 3. 运行项目
```bash
go run cmd/main.go
```

### 4. 测试接口
```bash
# 创建订单
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d @test_order.json

# 查询订单
curl http://localhost:8080/orders/order_20250707114508_user123
```

## 压测结果

### 查询接口性能
- QPS: 4891
- 平均响应时间: 3.8ms
- 100% 成功率

### 创建订单接口
- 支持并发库存扣减
- 乐观锁保证数据一致性

## 开发说明

这是一个演示项目，展示了 Go 语言在并发编程、Web 开发、数据库操作等方面的最佳实践。项目代码结构清晰，注释详细，适合学习和参考。 