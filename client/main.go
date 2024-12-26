package main

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

func main() {
	requests := 1
	for {
		fmt.Printf("Total number of requests servered: %d\n", requests)
		testLoadBalancer("localhost:8080")
		requests++
	}
}

func testLoadBalancer(addr string) bool {
	url, err := url.Parse("http://" + addr)
	if err != nil {
		return false
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Get(url.String())
	if err != nil {
		return false
	}
	defer response.Body.Close()
	return response.StatusCode == http.StatusOK
}
