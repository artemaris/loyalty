package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type UserCredentials struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

func main() {
	var baseURL = flag.String("url", "http://localhost:8080", "Base URL for API testing")
	flag.Parse()

	fmt.Printf("Testing GopherMart API at %s\n", *baseURL)
	fmt.Println("==================================")

	// Создаем HTTP клиент с поддержкой cookies
	jar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Printf("Failed to create cookie jar: %v\n", err)
		return
	}

	client := &http.Client{
		Jar:     jar,
		Timeout: 10 * time.Second,
	}

	// 1. Тест корневого endpoint
	fmt.Println("1. Testing root endpoint...")
	resp, err := client.Get(*baseURL + "/")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}
	fmt.Println()

	// 2. Тест регистрации пользователя
	fmt.Println("2. Testing user registration...")
	credentials := UserCredentials{
		Login:    "testuser",
		Password: "testpass",
	}

	jsonData, _ := json.Marshal(credentials)
	resp, err = client.Post(*baseURL+"/api/user/register", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body))
		resp.Body.Close()
	}
	fmt.Println()

	// 3. Тест входа пользователя
	fmt.Println("3. Testing user login...")
	resp, err = client.Post(*baseURL+"/api/user/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("Response: %s\n", string(body))
		resp.Body.Close()
	}
	fmt.Println()

	// 4. Тест загрузки заказа
	fmt.Println("4. Testing order upload...")
	resp, err = client.Post(*baseURL+"/api/user/orders", "text/plain", bytes.NewBufferString("12345678903"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}
	fmt.Println()

	// 5. Тест получения заказов
	fmt.Println("5. Testing get orders...")
	resp, err = client.Get(*baseURL + "/api/user/orders")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			fmt.Printf("Response: %s\n", string(body))
		}
		resp.Body.Close()
	}
	fmt.Println()

	// 6. Тест получения баланса
	fmt.Println("6. Testing get balance...")
	resp, err = client.Get(*baseURL + "/api/user/balance")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			fmt.Printf("Response: %s\n", string(body))
		}
		resp.Body.Close()
	}
	fmt.Println()

	// 7. Тест списания
	fmt.Println("7. Testing withdrawal...")
	withdrawal := WithdrawalRequest{
		Order: "2377225624",
		Sum:   100,
	}

	withdrawalData, _ := json.Marshal(withdrawal)
	resp, err = client.Post(*baseURL+"/api/user/balance/withdraw", "application/json", bytes.NewBuffer(withdrawalData))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		resp.Body.Close()
	}
	fmt.Println()

	// 8. Тест получения истории списаний
	fmt.Println("8. Testing get withdrawals...")
	resp, err = client.Get(*baseURL + "/api/user/withdrawals")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Status: %d\n", resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		if len(body) > 0 {
			fmt.Printf("Response: %s\n", string(body))
		}
		resp.Body.Close()
	}
	fmt.Println()

	fmt.Println("API testing completed!")
}
