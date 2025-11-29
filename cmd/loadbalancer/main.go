package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/AMEY-GAIKAR/load-balancer.git/internal/balance"
)

const PORT = 8080

func main() {
	algo := flag.String(
		"a",
		balance.RoundRobin,
		"load balancing algorithm, defaults to round robin",
	)
	port := flag.Int("port", PORT, "port to set up the load balancer")

	flag.Parse()

	balance := balance.InitLB(*algo)

	for _, backend := range flag.Args() {
		balance.AddBackend(backend, 0)
	}

	log.Println(balance.GetBackends())

	go balance.HealthCheckPeriodically(time.Second)

	log.Printf("%s load balancer set on port %d\n", balance.Method, *port)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", *port), balance); err != nil {
		log.Fatal(err)
	}
}
