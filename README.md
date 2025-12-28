Simple Load Balancer in Go

Algorithms implemented:
1) Round Robin
2) Weighted Round Robin *
3) Least Connections
4) Weighted Least Connections
5) IP Hash

Usage
```bash
go run cmd/loadbalancer/main.go -port <int> -a <RR | WRR | LC | WLC | IPH> urls_for_backend_servers...
```

Example usage
```bash
go run cmd/loadbalancer/main.go -port 8080 -a RR http://localhost:8081/ http://localhost:8082/ http://localhost:8083/

```

TODO
- implement wrr, wlc
- implement config file
