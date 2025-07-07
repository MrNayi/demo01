#!/bin/bash

# API测试脚本
BASE_URL="http://localhost:8080"

echo "🚀 开始API测试..."

# 测试商品相关API
echo "📦 测试商品API..."

# 1. 获取所有商品
echo "1. 获取所有商品"
curl -s -X GET "$BASE_URL/products" | jq '.'

# 2. 获取特定商品
echo -e "\n2. 获取商品ID=1"
curl -s -X GET "$BASE_URL/products/1" | jq '.'

# 3. 获取商品库存
echo -e "\n3. 获取商品ID=1的库存"
curl -s -X GET "$BASE_URL/products/1/stock" | jq '.'

# 4. 创建新商品
echo -e "\n4. 创建新商品"
curl -s -X POST "$BASE_URL/products" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试商品",
    "description": "这是一个测试商品",
    "price": 99.99,
    "stock": 50,
    "category": "测试",
    "status": "active"
  }' | jq '.'

# 测试订单相关API
echo -e "\n📋 测试订单API..."

# 5. 创建订单
echo "5. 创建订单"
curl -s -X POST "$BASE_URL/orders" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user_001",
    "items": [
      {
        "product_id": 1,
        "quantity": 2,
        "price": 5999.00
      },
      {
        "product_id": 3,
        "quantity": 1,
        "price": 1999.00
      }
    ]
  }' | jq '.'

# 6. 获取所有订单
echo -e "\n6. 获取所有订单"
curl -s -X GET "$BASE_URL/orders" | jq '.'

# 7. 获取特定订单
echo -e "\n7. 获取最新订单"
ORDER_ID=$(curl -s -X GET "$BASE_URL/orders" | jq -r '.data[0].id')
if [ "$ORDER_ID" != "null" ]; then
    curl -s -X GET "$BASE_URL/orders/$ORDER_ID" | jq '.'
else
    echo "没有找到订单"
fi

echo -e "\n✅ API测试完成！" 