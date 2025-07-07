package main

import (
	"log"
	"os"
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
	cfg := gateway.GatewayConfig{
		Nodes:    nodes,
		Replicas: 100,
	}
	log.Printf("Starting gateway on port %s", port)
	gateway.NewGateway(cfg).Serve(port)
}
