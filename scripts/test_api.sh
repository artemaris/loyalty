#!/bin/bash

# Скрипт для тестирования API
# Использование: ./scripts/test_api.sh <BASE_URL>
# Пример: ./scripts/test_api.sh http://localhost:8080

if [ $# -eq 0 ]; then
    echo "Usage: $0 <BASE_URL>"
    echo "Example: $0 http://localhost:8080"
    exit 1
fi

BASE_URL=$1

echo "Testing GopherMart API at $BASE_URL"
echo "=================================="

# Тест корневого endpoint
echo "1. Testing root endpoint..."
curl -s "$BASE_URL/" | head -1
echo

# Тест регистрации пользователя
echo "2. Testing user registration..."
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/register" \
    -H "Content-Type: application/json" \
    -d '{"login":"testuser","password":"testpass"}')
echo "Register response: $REGISTER_RESPONSE"
echo

# Тест входа пользователя
echo "3. Testing user login..."
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/login" \
    -H "Content-Type: application/json" \
    -d '{"login":"testuser","password":"testpass"}')
echo "Login response: $LOGIN_RESPONSE"
echo

# Извлекаем токен из cookies (упрощенная версия)
TOKEN=$(curl -s -c cookies.txt -X POST "$BASE_URL/api/user/login" \
    -H "Content-Type: application/json" \
    -d '{"login":"testuser","password":"testpass"}' > /dev/null && \
    grep auth_token cookies.txt | cut -f7)

echo "4. Testing order upload..."
ORDER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/orders" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: text/plain" \
    -d "12345678903")
echo "Order upload response: $ORDER_RESPONSE"
echo

echo "5. Testing get orders..."
curl -s -X GET "$BASE_URL/api/user/orders" \
    -H "Authorization: Bearer $TOKEN" | head -3
echo

echo "6. Testing get balance..."
curl -s -X GET "$BASE_URL/api/user/balance" \
    -H "Authorization: Bearer $TOKEN"
echo

echo "7. Testing withdrawal..."
WITHDRAWAL_RESPONSE=$(curl -s -X POST "$BASE_URL/api/user/balance/withdraw" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d '{"order":"2377225624","sum":100}')
echo "Withdrawal response: $WITHDRAWAL_RESPONSE"
echo

echo "8. Testing get withdrawals..."
curl -s -X GET "$BASE_URL/api/user/withdrawals" \
    -H "Authorization: Bearer $TOKEN"
echo

# Очистка
rm -f cookies.txt

echo "API testing completed!" 