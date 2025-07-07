package main

import (
	"log"
	"os"
	"strconv"

	grpcserver "shardo/internal/grpc"
)

func main() {
	port := os.Getenv("NODE_GRPC_PORT")
	if port == "" {
		port = "50051"
	}
	cacheSize := 128
	if v := os.Getenv("CACHE_SIZE_MB"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			cacheSize = n
		}
	}
	log.Printf("Starting node on port %s with cache size %dMB", port, cacheSize)
	grpcserver.StartGRPCServer(port, cacheSize)
}
