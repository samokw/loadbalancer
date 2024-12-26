# Load Balancer - Distributing Requests Among Servers

A simple load balancer implemented in Go that distributes incoming requests among three backend servers. It handles server health checks and periodically attempts to add servers back after once they go back online. The load balancer uses a least connections algorithm to distribute requests.

## Features

*   Distributes requests among three backend servers.
*   Performs health checks on servers to detect failures.
*   Automatically recovers from server failures by adding them back to the pool.
*   Uses a least connections algorithm for fair distribution of requests.

## Components

*   **Server:** Represents a backend server with its address. It also maintains a connection count for least connections selection.
*   **Server Pool:** Manages a collection of servers, storing them in both healthy and unhealthy states. Uses `sync.Map` for concurrent access.
*   **Load Balancer:** Selects a healthy server from the pool and forwards the request using a reverse proxy. It also manages the health checks and recovery process.

## Implementation Details

*   The `sync.Map` type is used to efficiently manage concurrent access to the server pool.
*   The `http.Client` with a timeout is used for health checks.
*   The `httputil.NewSingleHostReverseProxy` is used to forward requests to the selected server.
*   The `LeastConnections` function implements the least connections algorithm, selecting the server with the fewest current connections.

## Running the Load Balancer

*  `go run main.go`

## Running the Servers

*   `go run server1\main.go`
*   `go run server2\main.go`
*   `go run server3\main.go`

## Running the Client (Infinite Request)

*   `go run client\main.go`


## Disclaimer

This is a basic implementation for educational purposes.


## Future Plan

Maybe try to implement my own rate limiter