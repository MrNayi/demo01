## 项目结构
```
demo01/
├── cmd/main.go                    # 主程序入口
├── config/config.go               # 配置管理
├── internal/
│   ├── handler                    # HTTP 处理器
│   ├── service                    # 业务逻辑层
│   ├── repository                 # 数据访问层
│   ├── model                      # 数据模型
│   |── util                       # Redis 工具
|   |—— database                   # 数据库
├── /test                          # 测试数据
```
## 思考总结
1. 关于并发
并发的实际场景很复杂 为了用而去套一些并发知识是不合理的 比如waitgroup 我将他从扣减库存的逻辑取消 因为他将库存扣减从并发变成了串行，即使数据是安全的 实际也不会这么做性能太差
并发需要进行长期的学习和实践
2. 关于数据结构的设计

3. 关于分层
service只写业务逻辑 不应该涉及数据库相关的操作 包括redis、mysql甚至是本地缓存，这些数据交互应该都在repo实现

## 记录
### 1. 并发
语言
1. goroutine
2. channel 协程通信
3. select 多路复用模型
----
sync包
1. sync.Mutex 
2. sync.Map
3. sync.RWMutex  -- sync.RWLock
4. 
-------
原子操作
go也提供了atomic包 包下的数据类型都是并发安全的

---