package gateway

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"shardo/pkg/hashring"
	"shardo/proto/cachepb"

	"google.golang.org/grpc"
)

type Gateway struct {
	ring     *hashring.HashRing
	nodes    map[string]string // nodeName -> address
	replicas int
}

type GatewayConfig struct {
	Nodes    map[string]string // nodeName -> address
	Replicas int
}

func NewGateway(cfg GatewayConfig) *Gateway {
	ring := hashring.New(cfg.Replicas)
	for n := range cfg.Nodes {
		ring.AddNode(n)
	}
	return &Gateway{
		ring:     ring,
		nodes:    cfg.Nodes,
		replicas: cfg.Replicas,
	}
}

func (g *Gateway) Serve(port string) {
	http.HandleFunc("/get", g.handleGet)
	http.HandleFunc("/set", g.handleSet)
	http.HandleFunc("/delete", g.handleDelete)
	http.HandleFunc("/benchmark", g.handleBenchmark)
	log.Printf("Gateway listening on %s", port)
	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      nil,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway server error: %v", err)
	}
}

func (g *Gateway) getNodeAddr(key string) string {
	node := g.ring.GetNode(key)
	return g.nodes[node]
}

func (g *Gateway) handleGet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	addr := g.getNodeAddr(key)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		http.Error(w, "node unavailable", 500)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing gRPC connection: %v", err)
		}
	}()
	client := cachepb.NewCacheServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	resp, err := client.Get(ctx, &cachepb.GetRequest{Key: key})
	if err != nil || !resp.Found {
		http.Error(w, "not found", 404)
		return
	}
	if _, err := w.Write(resp.Value); err != nil {
		log.Printf("error writing response: %v", err)
	}
}

func (g *Gateway) handleSet(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	ttlStr := r.URL.Query().Get("ttl")
	value, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "invalid body", 400)
		return
	}
	addr := g.getNodeAddr(key)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		http.Error(w, "node unavailable", 500)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing gRPC connection: %v", err)
		}
	}()
	client := cachepb.NewCacheServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	ttl, _ := strconv.Atoi(ttlStr)
	if _, err := client.Set(ctx, &cachepb.SetRequest{Key: key, Value: value, Ttl: int64(ttl)}); err != nil {
		http.Error(w, "set failed", 500)
		return
	}
	w.WriteHeader(200)
}

func (g *Gateway) handleDelete(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	addr := g.getNodeAddr(key)
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		http.Error(w, "node unavailable", 500)
		return
	}
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("error closing gRPC connection: %v", err)
		}
	}()
	client := cachepb.NewCacheServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if _, err := client.Delete(ctx, &cachepb.DeleteRequest{Key: key}); err != nil {
		http.Error(w, "delete failed", 500)
		return
	}
	w.WriteHeader(200)
}

func (g *Gateway) handleBenchmark(w http.ResponseWriter, r *http.Request) {
	keys := 1000
	if k := r.URL.Query().Get("keys"); k != "" {
		keys, _ = strconv.Atoi(k)
	}
	start := time.Now()
	dist := make(map[string]int)
	for i := 0; i < keys; i++ {
		key := "bench" + strconv.Itoa(i)
		node := g.ring.GetNode(key)
		dist[node]++
		addr := g.nodes[node]
		conn, err := grpc.Dial(addr, grpc.WithInsecure())
		if err == nil {
			client := cachepb.NewCacheServiceClient(conn)
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			if _, err := client.Set(ctx, &cachepb.SetRequest{Key: key, Value: []byte("value"), Ttl: 60}); err != nil {
				log.Printf("benchmark set error: %v", err)
			}
			if _, err := client.Get(ctx, &cachepb.GetRequest{Key: key}); err != nil {
				log.Printf("benchmark get error: %v", err)
			}
			cancel()
			if err := conn.Close(); err != nil {
				log.Printf("error closing gRPC connection: %v", err)
			}
		}
	}
	elapsed := time.Since(start)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"latency_ms":   elapsed.Milliseconds(),
		"distribution": dist,
	}); err != nil {
		log.Printf("error encoding benchmark response: %v", err)
	}
}
