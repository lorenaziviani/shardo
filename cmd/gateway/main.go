package main

import (
	"log"
	"os"
	"strconv"
	"strings"

	"shardo/internal/gateway"
)

func main() {
	port := os.Getenv("GATEWAY_HTTP_PORT")
	if port == "" {
		port = "8080"
	}
	nodesEnv := os.Getenv("NODES")
	nodes := make(map[string]string)
	for _, pair := range strings.Split(nodesEnv, ",") {
		parts := strings.Split(pair, ":")
		if len(parts) == 3 {
			nodes[parts[0]] = parts[1] + ":" + parts[2]
		}
	}
	replicationFactor := 2
	if v := os.Getenv("SHARDO_REPLICATION_FACTOR"); v != "" {
		if val, err := strconv.Atoi(v); err == nil && val > 0 {
			replicationFactor = val
		}
	}
	cfg := gateway.GatewayConfig{
		Nodes:    nodes,
		Replicas: replicationFactor,
	}
	log.Printf("Starting gateway on port %s with replication factor %d", port, replicationFactor)
	gateway.NewGateway(cfg).Serve(port)
}
