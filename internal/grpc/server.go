package grpcserver

import (
	"context"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"shardo/pkg/cache"
	"shardo/proto/cachepb"
)

type CacheServiceServer interface {
	Get(context.Context, *GetRequest) (*GetResponse, error)
	Set(context.Context, *SetRequest) (*SetResponse, error)
	Delete(context.Context, *DeleteRequest) (*DeleteResponse, error)
	Metrics(context.Context, *MetricsRequest) (*MetricsResponse, error)
}

type server struct {
	cache *cache.Cache
	cachepb.UnimplementedCacheServiceServer
}

type GetRequest struct {
	Key string
}
type GetResponse struct {
	Value []byte
	Found bool
}
type SetRequest struct {
	Key   string
	Value []byte
	Ttl   int64 // seconds
}
type SetResponse struct{}
type DeleteRequest struct {
	Key string
}
type DeleteResponse struct{}
type MetricsRequest struct{}
type MetricsResponse struct {
	Hits   int32
	Misses int32
	Size   int32
}

func safeInt32(val int) int32 {
	if val > math.MaxInt32 {
		return math.MaxInt32
	}
	if val < math.MinInt32 {
		return math.MinInt32
	}
	return int32(val)
}

func (s *server) Get(ctx context.Context, req *cachepb.GetRequest) (*cachepb.GetResponse, error) {
	val, ok := s.cache.Get(req.Key)
	return &cachepb.GetResponse{Value: val, Found: ok}, nil
}
func (s *server) Set(ctx context.Context, req *cachepb.SetRequest) (*cachepb.SetResponse, error) {
	s.cache.Set(req.Key, req.Value, time.Duration(req.Ttl)*time.Second)
	return &cachepb.SetResponse{}, nil
}
func (s *server) Delete(ctx context.Context, req *cachepb.DeleteRequest) (*cachepb.DeleteResponse, error) {
	s.cache.Delete(req.Key)
	return &cachepb.DeleteResponse{}, nil
}
func (s *server) Metrics(ctx context.Context, req *cachepb.MetricsRequest) (*cachepb.MetricsResponse, error) {
	hits, misses, size := s.cache.Metrics()
	return &cachepb.MetricsResponse{
		Hits:   safeInt32(hits),
		Misses: safeInt32(misses),
		Size:   safeInt32(size),
	}, nil
}

func StartGRPCServer(port string, cacheCap int) {
	c := cache.New(cacheCap)
	s := grpc.NewServer()
	cachepb.RegisterCacheServiceServer(s, &server{cache: c})

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		srv := &http.Server{
			Addr:         ":" + os.Getenv("METRICS_PORT"),
			Handler:      mux,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		log.Printf("Prometheus metrics on :%s/metrics", os.Getenv("METRICS_PORT"))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("metrics server error: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("gRPC cache node listening on %s", port)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
