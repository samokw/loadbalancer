package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	Addr        string
	Connections int32
}

type ServerPool struct {
	servers          sync.Map
	unHealthyServers sync.Map
}

func (sp *ServerPool) AddServer(addr string) {
	sp.servers.Store(addr, &Server{
		Addr: addr,
	})
}
func (sp *ServerPool) RemoveServer(addr string) {
	server, loaded := sp.servers.Load(addr)
	if loaded {
		sp.servers.Delete(addr)
		sp.unHealthyServers.Store(addr, server)
	}
}
func (sp *ServerPool) GetServers() []*Server {
	servers := make([]*Server, 0)
	sp.servers.Range(func(_, value any) bool {
		servers = append(servers, value.(*Server))
		return true
	})
	return servers
}

func LeastConnections(servers []*Server) *Server {
	var bestServer *Server
	leastConnections := int32(1<<31 - 1) // Initailizing this to be as big of a number as possible at the start
	for _, server := range servers {
		if server.Connections < leastConnections {
			bestServer = server
			leastConnections = server.Connections
		}
	}
	return bestServer
}

func healthCheck(server *Server) bool {
	url, err := url.Parse("http://" + server.Addr)
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

func runCheck(sp *ServerPool, interval time.Duration) {
	for {
		time.Sleep(interval)

		var serversToRemove []string // Collect keys to remove

		sp.servers.Range(func(addr, server any) bool {
			if !healthCheck(server.(*Server)) {
				serversToRemove = append(serversToRemove, addr.(string))
			}
			return true
		})

		// Remove servers outside the Range loop to avoid issues
		for _, addr := range serversToRemove {
			sp.RemoveServer(addr)
			log.Printf("Server %s is down", addr) // Add logging here
		}
	}
}
func checkDeadServers(sp *ServerPool, deadcheck time.Duration) {
	for {
		time.Sleep(deadcheck)

		sp.unHealthyServers.Range(func(addr, server any) bool {
			if healthCheck(server.(*Server)) {
				sp.servers.Store(addr, server)
				sp.unHealthyServers.Delete(addr)
				log.Printf("Server %s is back online", addr) // Log when server comes back
			}
			return true
		})
	}
}

//Create a list of unhealthy servers and check if it came back online every 30 seconds

type LoadBalancer struct {
	serverPool *ServerPool
	algorithm  func(servers []*Server) *Server
	interval   time.Duration
	deadcheck  time.Duration
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	servers := lb.serverPool.GetServers()
	if len(servers) == 0 {
		http.Error(w, "There are no servers available", http.StatusServiceUnavailable)
		return
	}
	server := lb.algorithm(servers)
	if server == nil {
		http.Error(w, "There are no servers available", http.StatusServiceUnavailable)
		return
	}
	atomic.AddInt32(&server.Connections, 1)
	defer atomic.AddInt32(&server.Connections, -1)

	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   server.Addr,
	})
	proxy.ServeHTTP(w, r)
}

func main() {
	sp := &ServerPool{}
	alg := LeastConnections
	inter := 5 * time.Second
	deadcheck := 60 * time.Second

	sp.AddServer("127.0.0.1:8081")
	sp.AddServer("127.0.0.1:8082")
	sp.AddServer("127.0.0.1:8083")

	lb := &LoadBalancer{
		serverPool: sp,
		algorithm:  alg,
		interval:   inter,
		deadcheck:  deadcheck,
	}
	go runCheck(sp, inter)
	go checkDeadServers(sp, deadcheck)

	http.Handle("/", lb)
	fmt.Println("Starting load balancer server on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
