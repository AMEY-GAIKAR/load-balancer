Simple Load Balancer in Go

Algorithms implemented:
1) Round Robin
2) Weighted Round Robin
3) Least Connections
4) Weighted Least Connections
5) IP Hash

Usage

```bash
go run cmd/loadbalancer/main.go -port <int> -a <RR | WRR | LC | WLC | IPH> urls_for_backend_servers...
```

TODO
- implement config file
- implement wrr, wlc
