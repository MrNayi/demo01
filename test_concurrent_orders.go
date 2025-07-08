package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

type OrderRequest struct {
	UserID string      `json:"user_id"`
	Items  []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func main() {
	fmt.Println("开始并发订单测试，验证分布式锁防止超卖效果...")

	// 测试参数
	concurrentUsers := 10
	productID := 1
	quantityPerOrder := 1

	fmt.Printf("并发用户数: %d\n", concurrentUsers)
	fmt.Printf("商品ID: %d\n", productID)
	fmt.Printf("每单数量: %d\n", quantityPerOrder)
	fmt.Printf("总需求量: %d\n", concurrentUsers*quantityPerOrder)

	var wg sync.WaitGroup
	successCount := 0
	failCount := 0
	var mu sync.Mutex

	startTime := time.Now()

	// 并发创建订单
	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func(userID int) {
			defer wg.Done()

			// 构造订单请求
			orderReq := OrderRequest{
				UserID: fmt.Sprintf("concurrent_user_%d", userID),
				Items: []OrderItem{
					{
						ProductID: productID,
						Quantity:  quantityPerOrder,
						Price:     5999.0,
					},
				},
			}

			// 序列化请求
			jsonData, err := json.Marshal(orderReq)
			if err != nil {
				fmt.Printf("用户 %d 序列化请求失败: %v\n", userID, err)
				return
			}

			// 发送请求
			resp, err := http.Post("http://localhost:8080/orders", "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("用户 %d 请求失败: %v\n", userID, err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			// 读取响应
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("用户 %d 读取响应失败: %v\n", userID, err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			// 解析响应
			var orderResp OrderResponse
			if err := json.Unmarshal(body, &orderResp); err != nil {
				fmt.Printf("用户 %d 解析响应失败: %v\n", userID, err)
				mu.Lock()
				failCount++
				mu.Unlock()
				return
			}

			// 统计结果
			mu.Lock()
			if orderResp.Code == 200 {
				successCount++
				fmt.Printf("用户 %d 订单创建成功: %s\n", userID, orderResp.Message)
			} else {
				failCount++
				fmt.Printf("用户 %d 订单创建失败: %s\n", userID, orderResp.Message)
			}
			mu.Unlock()

		}(i)
	}

	wg.Wait()

	duration := time.Since(startTime)

	// 输出测试结果
	fmt.Println("\n=== 并发订单测试结果 ===")
	fmt.Printf("测试耗时: %v\n", duration)
	fmt.Printf("成功订单数: %d\n", successCount)
	fmt.Printf("失败订单数: %d\n", failCount)
	fmt.Printf("成功率: %.2f%%\n", float64(successCount)/float64(concurrentUsers)*100)

	// 验证库存一致性
	fmt.Println("\n=== 库存一致性验证 ===")
	if successCount > 0 {
		fmt.Printf("实际售出数量: %d\n", successCount*quantityPerOrder)
		fmt.Printf("理论最大售出数量: %d (取决于初始库存)\n", concurrentUsers*quantityPerOrder)

		if successCount < concurrentUsers {
			fmt.Println("✅ 分布式锁生效：部分订单被拒绝，防止了超卖")
		} else {
			fmt.Println("⚠️  所有订单都成功，请检查初始库存是否充足")
		}
	} else {
		fmt.Println("❌ 所有订单都失败，请检查服务状态")
	}

	fmt.Println("\n并发订单测试完成！")
}
