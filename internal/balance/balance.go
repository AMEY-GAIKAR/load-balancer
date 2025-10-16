package balance

import (
	"crypto/md5"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/AMEY-GAIKAR/load-balancer.git/internal/server"
)

const (
	RoundRobin               = "RR"
	WeightedRoundRobin       = "WRR"
	LeastConnections         = "LC"
	WeightedLeastConnections = "WLC"
	IPHash                   = "IPH"
)

type LoadBalancer struct {
	Method   string
	backends []*server.Server
	current  uint64
}

func (lb *LoadBalancer) AddBackend(urlString string, weight int) {
	newURL, err := url.Parse(urlString)
	if err != nil {
		log.Println(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(newURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
	}

	newBackend := &server.Server{
		Url:          newURL,
		Weight:       weight,
		ReverseProxy: proxy,
	}

	newBackend.SetStatus(lb.CheckBackendStatus(newBackend))

	lb.backends = append(lb.backends, newBackend)

	log.Printf("Backend added at %s\n", newURL.Host)
}

func (lb *LoadBalancer) GetBackends() []string {
	var backends []string

	for _, backend := range lb.backends {
		backends = append(backends, fmt.Sprintf("{URL: %s, Alive: %t, Weight: %d}\n", backend.Url, backend.IsAlive(), backend.Weight))
	}

	return backends
}

func (lb *LoadBalancer) SetBackendWeight(backend *server.Server, weight int) {
	backend.SetWeight(weight)
}

func (lb *LoadBalancer) HealthCheck() {
	for _, backend := range lb.backends {
		status := lb.CheckBackendStatus(backend)
		backend.SetStatus(status)
		if status {
			continue
			// log.Printf("Server at %s is alive", backend.Url.Host)
		} else {
			log.Printf("Server at %s is dead", backend.Url.Host)
		}
	}
}

func (lb *LoadBalancer) CheckBackendStatus(backend *server.Server) bool {
	var timeout time.Duration = 2 * time.Second
	conn, err := net.DialTimeout("tcp", backend.Url.Host, timeout)
	if err != nil {
		log.Printf("Falied to connect to backend at %s\n", backend.Url.Host)
		log.Println(err)
		return false
	}
	defer conn.Close()
	return true
}

func (lb *LoadBalancer) HealthCheckPeriodically(interval time.Duration) {
	t := time.NewTicker(interval)
	for {
		select {
		case <-t.C:
			lb.HealthCheck()
		}
	}
}

func (lb *LoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	backend := lb.NextPeer(r.Host)
	if backend == nil {
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}

	atomic.AddInt64(&backend.Connections, 1)
	defer atomic.AddInt64(&backend.Connections, -1)

	backend.ReverseProxy.ServeHTTP(w, r)
}

func InitLB(method string) *LoadBalancer {
	return &LoadBalancer{
		Method: method,
	}
}

func (lb *LoadBalancer) NextPeer(host string) *server.Server {
	switch lb.Method {
	case RoundRobin:
		return lb.RoundRobin()
	case WeightedRoundRobin:
		return lb.WeightedRoundRobin()
	case LeastConnections:
		return lb.LeastConnections()
	case WeightedLeastConnections:
		return lb.WeightedLeastConnections()
	case IPHash:
		return lb.IPHash(host)
	}
	return lb.RoundRobin()
}

func (lb *LoadBalancer) RoundRobin() *server.Server {
	var next uint64 = atomic.AddUint64(&lb.current, uint64(1)) % uint64(len(lb.backends))

	for i := next; i < uint64(len(lb.backends)); i++ {
		if lb.backends[i%uint64(len(lb.backends))].IsAlive() {
			return lb.backends[i%uint64(len(lb.backends))]
		}
	}
	return nil
}

func (lb *LoadBalancer) WeightedRoundRobin() *server.Server {
	return lb.RoundRobin()
}

func (lb *LoadBalancer) LeastConnections() *server.Server {
	var best *server.Server

	for _, backend := range lb.backends {
		if best == nil || atomic.LoadInt64(&backend.Connections) < atomic.LoadInt64(&best.Connections) && backend.IsAlive() {
			best = backend
		}
	}

	return best
}

func (lb *LoadBalancer) WeightedLeastConnections() *server.Server {
	var best *server.Server
	var bestLoad float64

	for _, backend := range lb.backends {
		load := float64(atomic.LoadInt64(&backend.Connections)) / float64(backend.Weight)
		if best == nil || load < bestLoad && backend.IsAlive() {
			best = backend
			bestLoad = load
		}
	}

	return best
}

func (lb *LoadBalancer) IPHash(host string) *server.Server {
	var backend *server.Server

	hash := md5.Sum([]byte(host))
	backend = lb.backends[int(hash[0])%len(lb.backends)]

	if !backend.IsAlive() {
		return lb.RoundRobin()
	}

	return backend
}
